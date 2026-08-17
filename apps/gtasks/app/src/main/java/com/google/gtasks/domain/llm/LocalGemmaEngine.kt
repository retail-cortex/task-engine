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

import android.content.Context
import android.util.Log
import com.google.mediapipe.tasks.genai.llminference.LlmInference
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.withContext
import java.io.File

class LocalGemmaEngine(private val context: Context) : LlmReasoningEngine {

    private val _isReady = MutableStateFlow(false)
    override val isReady: StateFlow<Boolean> = _isReady.asStateFlow()

    private val _statusMessage = MutableStateFlow("Initializing local reasoning...")
    override val statusMessage: StateFlow<String> = _statusMessage.asStateFlow()

    private var llmInference: LlmInference? = null
    
    companion object {
        private const val TAG = "LocalGemmaEngine"
        private const val MODEL_FILENAME = "gemma-2b-it-gpu-int4.bin"
    }

    init {
        // Automatically attempt initialization on startup
        initialize()
    }

    fun initialize() {
        val modelFile = File(context.filesDir, MODEL_FILENAME)
        if (!modelFile.exists()) {
            // Check if the model is packaged in the APK assets
            val hasAsset = try {
                context.assets.list("")?.contains(MODEL_FILENAME) == true
            } catch (e: Exception) {
                false
            }

            if (hasAsset) {
                try {
                    _statusMessage.value = "Extracting Gemma model from assets (this only happens once)..."
                    Log.i(TAG, "Extracting $MODEL_FILENAME from assets to filesDir...")
                    context.assets.open(MODEL_FILENAME).use { input ->
                        modelFile.outputStream().use { output ->
                            input.copyTo(output)
                        }
                    }
                    Log.i(TAG, "Gemma model extraction completed successfully.")
                } catch (e: Exception) {
                    val extractError = "Failed to extract Gemma from assets: ${e.localizedMessage}"
                    Log.e(TAG, extractError, e)
                    _statusMessage.value = extractError
                    _isReady.value = false
                    return
                }
            } else {
                val missingMsg = "Local Gemma model not found. Copy '$MODEL_FILENAME' to the app files directory to enable offline reasoning: ${context.filesDir.absolutePath}"
                Log.w(TAG, missingMsg)
                _statusMessage.value = missingMsg
                _isReady.value = false
                return
            }
        }

        try {
            _statusMessage.value = "Loading on-device Gemma model (this may take a few seconds)..."
            val options = LlmInference.LlmInferenceOptions.builder()
                .setModelPath(modelFile.absolutePath)
                .setMaxTokens(1024)
                .setTopK(40)
                .setTemperature(0.7f)
                .build()
                
            llmInference = LlmInference.createFromOptions(context, options)
            _statusMessage.value = "Local Gemma 2B model loaded and ready."
            _isReady.value = true
            Log.i(TAG, "Local Gemma 2B successfully initialized.")
        } catch (e: Exception) {
            val errorMsg = "Failed loading local Gemma: ${e.localizedMessage}"
            Log.e(TAG, errorMsg, e)
            _statusMessage.value = errorMsg
            _isReady.value = false
        }
    }

    override suspend fun generateResponse(prompt: String): Result<String> = withContext(Dispatchers.Default) {
        val inference = llmInference
        if (inference == null || !_isReady.value) {
            return@withContext Result.failure(IllegalStateException("Local Gemma model is not ready. Status: ${_statusMessage.value}"))
        }

        return@withContext try {
            val reply = inference.generateResponse(prompt)
            Result.success(reply)
        } catch (e: Exception) {
            Log.e(TAG, "Inference error", e)
            Result.failure(e)
        }
    }
}
