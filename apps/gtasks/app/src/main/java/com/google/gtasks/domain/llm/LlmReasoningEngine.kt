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
