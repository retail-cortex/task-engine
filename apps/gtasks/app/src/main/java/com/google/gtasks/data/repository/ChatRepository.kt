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
