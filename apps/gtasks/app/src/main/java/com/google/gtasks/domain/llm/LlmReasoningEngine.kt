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

package com.google.gtasks.domain.llm

import kotlinx.coroutines.flow.StateFlow

interface LlmReasoningEngine {
    /**
     * Flag indicating whether the model/client is fully loaded, configured, and ready for inference.
     */
    val isReady: StateFlow<Boolean>

    /**
     * Exposes a status message (e.g., "Model loaded successfully" or "gemma.bin not found").
     */
    val statusMessage: StateFlow<String>

    /**
     * Generate a response for a given text prompt.
     */
    suspend fun generateResponse(prompt: String): Result<String>
}
