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

package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// TraceProxyHandler proxies OTLP JSON traces from the frontend UI to Google Cloud Trace.
type TraceProxyHandler struct {
	httpClient *http.Client
	projectID  string
}

// NewTraceProxyHandler instantiates a new authenticated OTLP trace proxy.
func NewTraceProxyHandler(ctx context.Context) (*TraceProxyHandler, error) {
	// Find Google Application Default Credentials (ADC) programmatically
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/trace.append")
	if err != nil {
		return nil, err
	}

	// Configure a custom transport to skip TLS verification
	// This bypasses enterprise SSL-decrypting proxies that present custom CAs
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Wrap the custom transport with the oauth2 TokenSource to handle automatic authorization headers
	client := &http.Client{
		Transport: &oauth2.Transport{
			Source: creds.TokenSource,
			Base:   customTransport,
		},
	}

	// Resolve the target GCP Project ID once on boot
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = "cs-poc-gvosjaln9q6gcudiayjqdzq"
	}

	return &TraceProxyHandler{
		httpClient: client,
		projectID:  projectID,
	}, nil
}

// ProxyTraces handles the POST request from the frontend OTLP exporter.
func (h *TraceProxyHandler) ProxyTraces(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Inject the server-configured Google Cloud Project ID into the OTLP payload
	// This grounds the traces to the correct project without exposing it in the browser UI
	enrichedBody, err := h.injectProjectID(body)
	if err != nil {
		log.Printf("Warning: Failed to inject gcp.project_id into traces: %v. Sending raw payload.", err)
		enrichedBody = body
	}

	// Target Google Cloud OTLP Trace ingest endpoint
	// Google's universal OTLP/HTTP endpoint is https://telemetry.googleapis.com/v1/traces
	// The target GCP project is automatically resolved from the authenticated credentials.
	gcpEndpoint := "https://telemetry.googleapis.com/v1/traces"

	req, err := http.NewRequestWithContext(c.Request.Context(), "POST", gcpEndpoint, bytes.NewReader(enrichedBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upstream request"})
		return
	}

	// Copy content type and other relevant headers
	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	req.Header.Set("Content-Type", contentType)

	// Send the authenticated request to Google Cloud Trace
	resp, err := h.httpClient.Do(req)
	if err != nil {
		log.Printf("Error proxying traces to GCP: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to forward traces to Google Cloud"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read upstream response"})
		return
	}

	// Forward the status code and response body back to the frontend
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// injectProjectID parses the incoming OTLP JSON payload and inserts the "gcp.project_id"
// attribute into the resource section of every resource span.
func (h *TraceProxyHandler) injectProjectID(body []byte) ([]byte, error) {
	if len(body) == 0 {
		return body, nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	resourceSpans, ok := payload["resourceSpans"].([]interface{})
	if !ok {
		return body, nil // Return unmodified if not standard OTLP JSON structure
	}

	for _, rs := range resourceSpans {
		rsMap, ok := rs.(map[string]interface{})
		if !ok {
			continue
		}

		resource, ok := rsMap["resource"].(map[string]interface{})
		if !ok {
			resource = make(map[string]interface{})
			rsMap["resource"] = resource
		}

		attributes, ok := resource["attributes"].([]interface{})
		if !ok {
			attributes = make([]interface{}, 0)
		}

		// Check if gcp.project_id attribute already exists, and overwrite it
		exists := false
		for i, attr := range attributes {
			attrMap, ok := attr.(map[string]interface{})
			if !ok {
				continue
			}
			if attrMap["key"] == "gcp.project_id" {
				attrMap["value"] = map[string]interface{}{"stringValue": h.projectID}
				attributes[i] = attrMap
				exists = true
				break
			}
		}

		// If it doesn't exist, append it to the resource attributes
		if !exists {
			attributes = append(attributes, map[string]interface{}{
				"key": "gcp.project_id",
				"value": map[string]interface{}{
					"stringValue": h.projectID,
				},
			})
		}

		resource["attributes"] = attributes
	}

	return json.Marshal(payload)
}
