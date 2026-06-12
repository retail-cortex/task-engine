// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package telemetry

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTelemetry initializes OpenTelemetry with the Google Cloud Trace exporter.
// It returns a shutdown function to be called on application exit to flush remaining traces.
func InitTelemetry(ctx context.Context, serviceName string) (func(), error) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		// Fallback to a default project if not set (helps in local offline development)
		projectID = "cs-poc-gvosjaln9q6gcudiayjqdzq"
		log.Printf("Warning: GOOGLE_CLOUD_PROJECT environment variable not set. Using default project: %s", projectID)
	}

	log.Printf("Initializing OpenTelemetry tracing for service '%s' targeting Google Cloud Project: %s", serviceName, projectID)

	// Create Google Cloud Trace exporter
	exporter, err := texporter.New(texporter.WithProjectID(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Cloud Trace exporter: %w", err)
	}

	// Create resource definition
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry resource: %w", err)
	}

	// Configure the BatchSpanProcessor with aggressive flushing parameters
	// - BatchTimeout: 60 seconds (ensures traces are flushed at least every minute)
	// - MaxQueueSize: 2048
	// - MaxExportBatchSize: 512
	bsp := sdktrace.NewBatchSpanProcessor(
		exporter,
		sdktrace.WithBatchTimeout(60*time.Second),
		sdktrace.WithMaxQueueSize(2048),
		sdktrace.WithMaxExportBatchSize(512),
	)

	// Build the TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Always sample for POC/Demo traces
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// Set global TracerProvider
	otel.SetTracerProvider(tp)

	// Eagerly warm up the connection pool in a background goroutine to absorb the initial ADC/TLS handshake penalty
	go func() {
		log.Println("[Telemetry] Eagerly warming up Google Cloud Trace connection pool in the background...")
		tracer := otel.Tracer("warmup")
		_, span := tracer.Start(context.Background(), "WarmupConnectionPool")
		span.End()

		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.ForceFlush(flushCtx); err != nil {
			log.Printf("[Telemetry] Warmup trace flush failed or timed out: %v. This is normal if offline.", err)
		} else {
			log.Println("[Telemetry] Warmup trace flushed successfully! Connection pool is warm and ready.")
		}
	}()

	// Set global Propagator to W3C Trace Context (crucial for cross-service propagation!)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Return a shutdown function that flushes and cleans up
	shutdown := func() {
		log.Println("Shutting down OpenTelemetry TracerProvider... Flushing remaining spans.")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := tp.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error shutting down TracerProvider: %v", err)
		} else {
			log.Println("TracerProvider shut down successfully.")
		}
	}

	return shutdown, nil
}
