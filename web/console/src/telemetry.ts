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

import { WebTracerProvider } from '@opentelemetry/sdk-trace-web';
import { BatchSpanProcessor } from '@opentelemetry/sdk-trace-base';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { registerInstrumentations } from '@opentelemetry/instrumentation';
import { FetchInstrumentation } from '@opentelemetry/instrumentation-fetch';
import { ZoneContextManager } from '@opentelemetry/context-zone';
import { Resource } from '@opentelemetry/resources';
import { ATTR_SERVICE_NAME } from '@opentelemetry/semantic-conventions';

export const initTelemetry = (serviceName: string) => {
  const provider = new WebTracerProvider({
    resource: new Resource({
      [ATTR_SERVICE_NAME]: serviceName,
    }),
  });

  // Configure OTLP HTTP exporter pointing to our Go API secure trace proxy endpoint
  const exporter = new OTLPTraceExporter({
    url: '/api/v1/traces',
  });

  // Use BatchSpanProcessor for optimal performance, flushing traces in batches
  provider.addSpanProcessor(
    new BatchSpanProcessor(exporter, {
      maxQueueSize: 100,
      maxExportBatchSize: 10,
      scheduledDelayMillis: 1000, // Aggressive flushing: flush every second
    })
  );

  provider.register({
    contextManager: new ZoneContextManager(),
  });

  // Register automatic fetch instrumentation to auto-capture HTTP requests
  registerInstrumentations({
    instrumentations: [
      new FetchInstrumentation({
        propagateTraceHeaderCorsUrls: [
          /.*/, // Propagate W3C trace headers to all requests (API and Agent)
        ],
      }),
    ],
    tracerProvider: provider,
  });

  console.log(`OpenTelemetry Web Tracing initialized for ${serviceName}`);
};
