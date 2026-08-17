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
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

func TestInitTelemetry(t *testing.T) {
	// Backup and restore environment
	oldProj := os.Getenv("GOOGLE_CLOUD_PROJECT")
	defer os.Setenv("GOOGLE_CLOUD_PROJECT", oldProj)

	t.Run("default fallback project ID when env not set", func(t *testing.T) {
		os.Unsetenv("GOOGLE_CLOUD_PROJECT")

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		shutdown, err := InitTelemetry(ctx, "test-service")
		// Even if ADC is missing or fails, we verify the tracer provider registration or error handling
		if err == nil {
			assert.NotNil(t, shutdown)
			tp := otel.GetTracerProvider()
			assert.NotNil(t, tp)

			tracer := tp.Tracer("test-tracer")
			_, span := tracer.Start(context.Background(), "test-span")
			span.End()

			shutdown()
		} else {
			assert.Contains(t, err.Error(), "failed to create Google Cloud Trace exporter")
		}
	})

	t.Run("explicit project ID set", func(t *testing.T) {
		os.Setenv("GOOGLE_CLOUD_PROJECT", "custom-test-project-123")

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		shutdown, err := InitTelemetry(ctx, "test-service-2")
		if err == nil {
			assert.NotNil(t, shutdown)
			shutdown()
		}
	})
}
