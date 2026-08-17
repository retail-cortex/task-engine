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

package com.google.gtasks.ui.screens.chat

import androidx.compose.runtime.mutableStateListOf
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.ViewModelProvider.AndroidViewModelFactory.Companion.APPLICATION_KEY
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.CreationExtras
import com.google.gtasks.GTasksApplication
import com.google.gtasks.data.repository.AuthRepository
import com.google.gtasks.data.repository.ChatRepository
import com.google.gtasks.data.repository.TaskRepository
import com.google.gtasks.ui.a2ui.A2UITransaction
import com.google.gtasks.ui.a2ui.ButtonAction
import com.google.gtasks.domain.llm.LlmReasoningEngine
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.jsonPrimitive
import java.text.SimpleDateFormat
import java.util.*

class ChatViewModel(
    private val chatRepository: ChatRepository,
    private val taskRepository: TaskRepository,
    private val authRepository: AuthRepository,
    private val localGemmaEngine: LlmReasoningEngine,
    private val remoteGeminiEngine: LlmReasoningEngine,
    private val application: GTasksApplication
) : ViewModel() {

    private val _messages = MutableStateFlow<List<ChatMessage>>(emptyList())
    val messages: StateFlow<List<ChatMessage>> = _messages.asStateFlow()

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    // Expose local Gemma readiness state
    val isLocalGemmaReady: StateFlow<Boolean> = localGemmaEngine.isReady
    val localGemmaStatus: StateFlow<String> = localGemmaEngine.statusMessage

    // Toggle for Local vs Remote
    private val _useLocalReasoning = MutableStateFlow(application.container.useLocalReasoning)
    val useLocalReasoning: StateFlow<Boolean> = _useLocalReasoning.asStateFlow()

    private val siteId: String get() = authRepository.activeSiteId ?: ""
    private val userId: String get() = authRepository.activeUserId ?: ""
    


    init {
        // Seed initial greeting from Hanna
        _messages.value = listOf(
            ChatMessage(
                id = "hanna-welcome",
                role = "assistant",
                content = "Well, look who finally clocked in. I'm Hanna, your operational coach. Ask me about cash drawer drops, stock corrections, or swap a shift with a coworker. Try not to mess up today."
            )
        )
    }

    fun toggleReasoningMode() {
        val newValue = !_useLocalReasoning.value
        _useLocalReasoning.value = newValue
        application.container.useLocalReasoning = newValue
        
        // Append a system guidance message
        val modeText = if (newValue) "On-Device Gemma 2B (Offline reasoning)" else "Cloud Gemini Grounded (RAG + full A2UI support)"
        val sysMsg = ChatMessage(
            id = "sys-${System.currentTimeMillis()}",
            role = "system",
            content = "Reasoning engine switched to: $modeText"
        )
        _messages.value = _messages.value + sysMsg
    }

    fun sendMessage(text: String) {
        if (text.isBlank()) return

        val userMsg = ChatMessage(
            id = "user-${System.currentTimeMillis()}",
            role = "user",
            content = text
        )
        
        _messages.value = _messages.value + userMsg
        _isLoading.value = true

        viewModelScope.launch {
            val isLocal = _useLocalReasoning.value && isLocalGemmaReady.value
            
            if (isLocal) {
                // 1. Run Local Gemma 2B Reasoning on NPU/GPU
                val result = localGemmaEngine.generateResponse(text)
                _isLoading.value = false
                result.onSuccess { reply ->
                    val replyMsg = ChatMessage(
                        id = "gemma-${System.currentTimeMillis()}",
                        role = "assistant",
                        content = reply
                    )
                    _messages.value = _messages.value + replyMsg
                }.onFailure { error ->
                    val replyMsg = ChatMessage(
                        id = "gemma-err-${System.currentTimeMillis()}",
                        role = "assistant",
                        content = "Local reasoning failed: ${error.localizedMessage}. Falling back to cloud."
                    )
                    _messages.value = _messages.value + replyMsg
                    // Fallback to cloud
                    runCloudGemini(text)
                }
            } else {
                // 2. Run Remote Cloud Gemini Grounded Reasoning
                runCloudGemini(text)
            }
        }
    }

    private suspend fun runCloudGemini(text: String) {
        val result = chatRepository.sendMessage(
            siteId = siteId,
            userId = userId,
            shiftId = SHIFT_SESSION_ID,
            message = text
        )
        _isLoading.value = false
        result.onSuccess { response ->
            val replyMsg = ChatMessage(
                id = response.id,
                role = response.role,
                content = response.content,
                transaction = response.a2uiData
            )
            _messages.value = _messages.value + replyMsg
        }.onFailure { error ->
            val replyMsg = ChatMessage(
                id = "cloud-err-${System.currentTimeMillis()}",
                role = "assistant",
                content = "Sigh, I couldn't reach the server. Let's blame the store Wi-Fi. Error: ${error.localizedMessage}"
            )
            _messages.value = _messages.value + replyMsg
        }
    }

    /**
     * Intercept and handle A2UI interactive Card actions natively in the Android client!
     */
    fun handleA2UIAction(action: ButtonAction, dataModel: Map<String, JsonElement>) {
        val actionType = action.type
        val timestamp = System.currentTimeMillis()

        viewModelScope.launch {
            _isLoading.value = true
            when (actionType) {
                "OVERRIDE" -> {
                    // Extract vault drop override parameters from dataModel or action contexts
                    val taskID = dataModel["taskExecutionID"]?.jsonPrimitive?.contentOrNull
                        ?: action.context.find { it.key == "taskExecutionID" }?.value?.literalString
                        ?: ""
                    val assetID = dataModel["assetID"]?.jsonPrimitive?.contentOrNull
                        ?: action.context.find { it.key == "assetID" }?.value?.literalString
                        ?: ""
                    val justification = dataModel["justification"]?.jsonPrimitive?.contentOrNull
                        ?: action.context.find { it.key == "justification" }?.value?.literalString
                        ?: "Verified cash drops pouches secured inside main backend vault"

                    val result = taskRepository.overrideAssetConstraint(siteId, taskID, assetID, justification)
                    _isLoading.value = false
                    result.onSuccess {
                        val sysMsg = ChatMessage(
                            id = "sys-$timestamp",
                            role = "system",
                            content = "Transaction Success: Supervisor cash vault drop override authorized! Audit ledger updated."
                        )
                        _messages.value = _messages.value + sysMsg
                    }.onFailure { error ->
                        val sysMsg = ChatMessage(
                            id = "sys-$timestamp",
                            role = "system",
                            content = "Transaction Failed: ${error.localizedMessage}"
                        )
                        _messages.value = _messages.value + sysMsg
                    }
                }
                "TRADE" -> {
                    // Extract task trade parameters
                    val taskID = dataModel["taskExecutionID"]?.jsonPrimitive?.contentOrNull ?: ""
                    val proposedAssigneeID = dataModel["proposedAssigneeID"]?.jsonPrimitive?.contentOrNull ?: ""
                    
                    // Trigger a mock network trade success message (since we demonstrate the flow)
                    _isLoading.value = false
                    val sysMsg = ChatMessage(
                        id = "sys-$timestamp",
                        role = "system",
                        content = "Transaction Success: Handover trade proposal submitted. Awaiting supervisor approval rule triggers."
                    )
                    _messages.value = _messages.value + sysMsg
                }
                else -> {
                    _isLoading.value = false
                    val sysMsg = ChatMessage(
                        id = "sys-$timestamp",
                        role = "system",
                        content = "Action '${actionType}' triggered with data: ${dataModel.entries.joinToString { "${it.key}=${it.value}" }}"
                    )
                    _messages.value = _messages.value + sysMsg
                }
            }
        }
    }

    companion object {
        private const val SHIFT_SESSION_ID = "11111111-1111-1111-1111-111111111111"

        val Factory: ViewModelProvider.Factory = object : ViewModelProvider.Factory {
            @Suppress("UNCHECKED_CAST")
            override fun <T : ViewModel> create(modelClass: Class<T>, extras: CreationExtras): T {
                val application = checkNotNull(extras[APPLICATION_KEY]) as GTasksApplication
                return ChatViewModel(
                    chatRepository = application.container.chatRepository,
                    taskRepository = application.container.taskRepository,
                    authRepository = application.container.authRepository,
                    localGemmaEngine = application.container.localGemmaEngine,
                    remoteGeminiEngine = application.container.remoteGeminiEngine,
                    application = application
                ) as T
            }
        }
    }
}

data class ChatMessage(
    val id: String,
    val role: String, // "user", "assistant", "system"
    val content: String,
    val transaction: A2UITransaction? = null
)
