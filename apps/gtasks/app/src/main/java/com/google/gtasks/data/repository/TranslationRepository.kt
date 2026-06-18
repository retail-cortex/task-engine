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
import com.google.gtasks.data.api.CloneVoiceResponse
import com.google.gtasks.data.api.HDVoice
import com.google.gtasks.data.api.ProfileDTO
import com.google.gtasks.data.api.PreviewVoiceRequest
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.RequestBody.Companion.toRequestBody
import java.io.IOException

class TranslationRepository(private val apiClient: ApiClient) {
    private val apiService get() = apiClient.apiService

    suspend fun getProfile(userId: String): Result<ProfileDTO> = withContext(Dispatchers.IO) {
        return@withContext try {
            val response = apiService.getProfile(userId)
            Result.success(response)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun saveProfile(userId: String, profile: ProfileDTO): Result<ProfileDTO> = withContext(Dispatchers.IO) {
        return@withContext try {
            val response = apiService.saveProfile(userId, profile)
            Result.success(response)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun updateProfile(userId: String, profile: ProfileDTO): Result<ProfileDTO> = withContext(Dispatchers.IO) {
        return@withContext try {
            val response = apiService.updateProfile(userId, profile)
            Result.success(response)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun cloneVoice(userId: String, audioBytes: ByteArray): Result<CloneVoiceResponse> = withContext(Dispatchers.IO) {
        return@withContext try {
            val requestBody = audioBytes.toRequestBody("audio/wav".toMediaTypeOrNull())
            val response = apiService.cloneVoice(userId, requestBody)
            Result.success(response)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun translateTalk(
        associateId: String,
        targetLanguage: String,
        audioBytes: ByteArray
    ): Result<Pair<ByteArray, String>> = withContext(Dispatchers.IO) {
        return@withContext try {
            val requestBody = audioBytes.toRequestBody("application/octet-stream".toMediaTypeOrNull())
            val response = apiService.translateTalk(associateId, targetLanguage, requestBody)
            
            if (response.isSuccessful) {
                val audioOut = response.body()?.bytes() ?: ByteArray(0)
                val encodedText = response.headers()["X-Translated-Text"] ?: ""
                val translatedText = try {
                    java.net.URLDecoder.decode(encodedText, "UTF-8")
                } catch (e: Exception) {
                    encodedText
                }
                Result.success(Pair(audioOut, translatedText))
            } else {
                val errorMsg = response.errorBody()?.string() ?: "Unknown API error"
                Result.failure(IOException("HTTP ${response.code()}: $errorMsg"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun translateListen(
        associateId: String,
        gender: String,
        targetLanguage: String,
        audioBytes: ByteArray
    ): Result<Pair<ByteArray, String>> = withContext(Dispatchers.IO) {
        return@withContext try {
            val requestBody = audioBytes.toRequestBody("application/octet-stream".toMediaTypeOrNull())
            val response = apiService.translateListen(associateId, gender, targetLanguage, requestBody)
            
            if (response.isSuccessful) {
                val audioOut = response.body()?.bytes() ?: ByteArray(0)
                val encodedText = response.headers()["X-Translated-Text"] ?: ""
                val translatedText = try {
                    java.net.URLDecoder.decode(encodedText, "UTF-8")
                } catch (e: Exception) {
                    encodedText
                }
                Result.success(Pair(audioOut, translatedText))
            } else {
                val errorMsg = response.errorBody()?.string() ?: "Unknown API error"
                Result.failure(IOException("HTTP ${response.code()}: $errorMsg"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun listVoices(): Result<List<HDVoice>> = withContext(Dispatchers.IO) {
        return@withContext try {
            val response = apiService.listVoices()
            Result.success(response)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun previewVoice(voiceName: String, languageCode: String): Result<ByteArray> = withContext(Dispatchers.IO) {
        return@withContext try {
            val response = apiService.previewVoice(PreviewVoiceRequest(voiceName, languageCode))
            if (response.isSuccessful) {
                val audioOut = response.body()?.bytes() ?: ByteArray(0)
                Result.success(audioOut)
            } else {
                val errorMsg = response.errorBody()?.string() ?: "Unknown API error"
                Result.failure(IOException("HTTP ${response.code()}: $errorMsg"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
