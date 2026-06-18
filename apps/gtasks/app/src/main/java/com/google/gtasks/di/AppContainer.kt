package com.google.gtasks.di

import android.content.Context
import android.content.SharedPreferences
import com.google.gtasks.data.api.ApiClient
import com.google.gtasks.data.repository.AuthRepository
import com.google.gtasks.data.repository.ChatRepository
import com.google.gtasks.data.repository.TaskRepository
import com.google.gtasks.domain.llm.LlmReasoningEngine
import com.google.gtasks.domain.llm.LocalGemmaEngine
import com.google.gtasks.domain.llm.RemoteGeminiEngine

class AppContainer(private val context: Context) {

    private val prefs: SharedPreferences = context.getSharedPreferences("gtasks_prefs", Context.MODE_PRIVATE)

    companion object {
        private const val PREF_USE_LOCAL_REASONING = "use_local_reasoning"
    }

    // 1. Core Network Client
    val apiClient: ApiClient by lazy {
        ApiClient(context)
    }

    // 2. Repositories
    val authRepository: AuthRepository by lazy {
        AuthRepository(context, apiClient).also { authRepo ->
            apiClient.onUnauthorizedListener = {
                authRepo.logout()
            }
        }
    }

    val taskRepository: TaskRepository by lazy {
        TaskRepository(apiClient, authRepository)
    }

    val chatRepository: ChatRepository by lazy {
        ChatRepository(apiClient, authRepository)
    }

    // 3. Reasoning Engines
    val localGemmaEngine: LocalGemmaEngine by lazy {
        LocalGemmaEngine(context)
    }

    val remoteGeminiEngine: RemoteGeminiEngine by lazy {
        RemoteGeminiEngine(chatRepository, authRepository)
    }

    // Dynamic configuration for active reasoning engine
    var useLocalReasoning: Boolean
        get() = prefs.getBoolean(PREF_USE_LOCAL_REASONING, false)
        set(value) {
            prefs.edit().putBoolean(PREF_USE_LOCAL_REASONING, value).apply()
        }

    val activeLlmEngine: LlmReasoningEngine
        get() = if (useLocalReasoning && localGemmaEngine.isReady.value) {
            localGemmaEngine
        } else {
            remoteGeminiEngine
        }
}
