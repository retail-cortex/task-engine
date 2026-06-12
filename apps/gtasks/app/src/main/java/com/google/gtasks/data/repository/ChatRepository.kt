package com.google.gtasks.data.repository

import com.google.gtasks.data.api.ApiClient
import com.google.gtasks.data.api.MessageRequest
import com.google.gtasks.data.model.MessageResponse
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

class ChatRepository(
    private val apiClient: ApiClient,
    private val authRepository: AuthRepository
) {
    private val apiService get() = apiClient.apiService
    private val orgId get() = authRepository.activeOrgId

    /**
     * Send a conversational message to the backend Gemini ADK agent.
     * Returns a MessageResponse which can contain markdown text replies and/or structured A2UI cards.
     */
    suspend fun sendMessage(
        siteId: String,
        userId: String,
        shiftId: String,
        message: String
    ): Result<MessageResponse> = withContext(Dispatchers.IO) {
        return@withContext try {
            val response = apiService.sendMessage(
                orgId = orgId,
                siteId = siteId,
                userId = userId,
                shiftId = shiftId,
                body = MessageRequest(message = message)
            )
            Result.success(response)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
