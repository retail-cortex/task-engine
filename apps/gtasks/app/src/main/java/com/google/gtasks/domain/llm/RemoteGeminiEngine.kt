package com.google.gtasks.domain.llm

import com.google.gtasks.data.repository.AuthRepository
import com.google.gtasks.data.repository.ChatRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

class RemoteGeminiEngine(
    private val chatRepository: ChatRepository,
    private val authRepository: AuthRepository
) : LlmReasoningEngine {

    private val _isReady = MutableStateFlow(true)
    override val isReady: StateFlow<Boolean> = _isReady.asStateFlow()

    private val _statusMessage = MutableStateFlow("Remote Gemini Cloud Grounded reasoning is active.")
    override val statusMessage: StateFlow<String> = _statusMessage.asStateFlow()

    companion object {
        // Shift session ID constant to align with local development
        private const val DEFAULT_SHIFT_ID = "11111111-1111-1111-1111-111111111111"
    }

    override suspend fun generateResponse(prompt: String): Result<String> {
        val siteId = authRepository.activeSiteId
            ?: return Result.failure(IllegalStateException("No active site context selected."))
        val userId = authRepository.activeUserId
            ?: return Result.failure(IllegalStateException("No authenticated user profile."))

        val result = chatRepository.sendMessage(
            siteId = siteId,
            userId = userId,
            shiftId = DEFAULT_SHIFT_ID,
            message = prompt
        )

        return result.map { it.content }
    }
}
