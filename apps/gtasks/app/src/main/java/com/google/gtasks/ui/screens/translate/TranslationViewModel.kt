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

package com.google.gtasks.ui.screens.translate

import android.annotation.SuppressLint
import android.media.AudioFormat
import android.media.AudioRecord
import android.media.MediaPlayer
import android.media.MediaRecorder
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.ViewModelProvider.AndroidViewModelFactory.Companion.APPLICATION_KEY
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.CreationExtras
import com.google.gtasks.GTasksApplication
import com.google.gtasks.data.api.HDVoice
import com.google.gtasks.data.api.ProfileDTO
import com.google.gtasks.data.repository.AuthRepository
import com.google.gtasks.data.repository.TranslationRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import java.io.ByteArrayOutputStream
import java.io.File

class TranslationViewModel(
    private val translationRepository: TranslationRepository,
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<TranslationUiState>(TranslationUiState.Loading)
    val uiState: StateFlow<TranslationUiState> = _uiState.asStateFlow()

    private val _isRecording = MutableStateFlow(false)
    val isRecording: StateFlow<Boolean> = _isRecording.asStateFlow()

    private val _activeRecordingMode = MutableStateFlow<RecordingMode?>(null) // TALK or LISTEN
    val activeRecordingMode: StateFlow<RecordingMode?> = _activeRecordingMode.asStateFlow()

    private val _conversation = MutableStateFlow<List<Utterance>>(emptyList())
    val conversation: StateFlow<List<Utterance>> = _conversation.asStateFlow()

    private val _voices = MutableStateFlow<List<HDVoice>>(emptyList())
    val voices: StateFlow<List<HDVoice>> = _voices.asStateFlow()

    private val _profile = MutableStateFlow<ProfileDTO?>(null)
    val profile: StateFlow<ProfileDTO?> = _profile.asStateFlow()

    private val _errorMessage = MutableStateFlow<String?>(null)
    val errorMessage: StateFlow<String?> = _errorMessage.asStateFlow()

    private var audioRecord: AudioRecord? = null
    private val recordingBuffer = ByteArrayOutputStream()
    private var mediaPlayer: MediaPlayer? = null

    // Audio recording configuration (PCM 16-bit Mono 16kHz for Google Speech-to-Text LINEAR16)
    private val sampleRate = 16000
    private val channelConfig = AudioFormat.CHANNEL_IN_MONO
    private val audioFormat = AudioFormat.ENCODING_PCM_16BIT

    init {
        loadData()
    }

    fun loadData() {
        viewModelScope.launch {
            _uiState.value = TranslationUiState.Loading
            _errorMessage.value = null

            val userId = authRepository.activeUserId
            if (userId == null) {
                _uiState.value = TranslationUiState.Error("User session not found")
                return@launch
            }

            // Parallel load voices and user profile
            val voicesResult = translationRepository.listVoices()
            val profileResult = translationRepository.getProfile(userId)

            if (voicesResult.isSuccess && profileResult.isSuccess) {
                _voices.value = voicesResult.getOrThrow()
                _profile.value = profileResult.getOrThrow()
                _uiState.value = TranslationUiState.Success
            } else {
                val voiceErr = voicesResult.exceptionOrNull()?.message ?: ""
                val profileErr = profileResult.exceptionOrNull()?.message ?: ""
                _uiState.value = TranslationUiState.Error("Failed to initialize: $voiceErr $profileErr")
            }
        }
    }

    fun updateVoicePreferences(gender: String, voiceName: String) {
        val currentProfile = _profile.value ?: return
        viewModelScope.launch {
            val updated = currentProfile.copy(
                voiceGenderPreference = gender,
                voiceNamePreference = voiceName
            )
            val result = translationRepository.updateProfile(currentProfile.id, updated)
            if (result.isSuccess) {
                _profile.value = result.getOrThrow()
            } else {
                _errorMessage.value = "Failed to update profile: ${result.exceptionOrNull()?.message}"
            }
        }
    }

    @SuppressLint("MissingPermission")
    fun startRecording(mode: RecordingMode) {
        if (_isRecording.value) return
        _isRecording.value = true
        _activeRecordingMode.value = mode
        _errorMessage.value = null
        recordingBuffer.reset()

        viewModelScope.launch(Dispatchers.IO) {
            try {
                val minBufferSize = AudioRecord.getMinBufferSize(sampleRate, channelConfig, audioFormat)
                if (minBufferSize == AudioRecord.ERROR || minBufferSize == AudioRecord.ERROR_BAD_VALUE) {
                    throw IllegalStateException("Invalid buffer size returned")
                }

                val recorder = AudioRecord(
                    MediaRecorder.AudioSource.MIC,
                    sampleRate,
                    channelConfig,
                    audioFormat,
                    minBufferSize
                )
                audioRecord = recorder

                if (recorder.state != AudioRecord.STATE_INITIALIZED) {
                    throw IllegalStateException("AudioRecord initialization failed")
                }

                recorder.startRecording()
                val buffer = ByteArray(minBufferSize)
                while (_isRecording.value) {
                    val read = recorder.read(buffer, 0, buffer.size)
                    if (read > 0) {
                        recordingBuffer.write(buffer, 0, read)
                    }
                }
            } catch (e: SecurityException) {
                _isRecording.value = false
                _activeRecordingMode.value = null
                _errorMessage.value = "Microphone recording permission denied."
            } catch (e: Exception) {
                _isRecording.value = false
                _activeRecordingMode.value = null
                _errorMessage.value = "Audio recording failed: ${e.message}"
            } finally {
                cleanupRecorder()
            }
        }
    }

    fun stopRecordingAndTranslate(targetLanguage: String, customerGender: String = "FEMALE") {
        if (!_isRecording.value) return
        _isRecording.value = false
        val mode = _activeRecordingMode.value
        _activeRecordingMode.value = null

        val audioBytes = recordingBuffer.toByteArray()
        cleanupRecorder()

        if (audioBytes.isEmpty() || mode == null) return

        val associateId = authRepository.activeUserId ?: return

        viewModelScope.launch {
            _uiState.value = TranslationUiState.Translating

            val result = if (mode == RecordingMode.TALK) {
                translationRepository.translateTalk(associateId, targetLanguage, audioBytes)
            } else {
                translationRepository.translateListen(associateId, customerGender, targetLanguage, audioBytes)
            }

            if (result.isSuccess) {
                val (translatedAudio, text) = result.getOrThrow()
                
                val utterance = Utterance(
                    speaker = if (mode == RecordingMode.TALK) Speaker.ASSOCIATE else Speaker.CUSTOMER,
                    originalText = if (mode == RecordingMode.TALK) "Talk Recording" else "Customer Speech",
                    translatedText = text,
                    audioBytes = translatedAudio
                )
                
                _conversation.value = _conversation.value + utterance
                _uiState.value = TranslationUiState.Success

                // Auto-play the synthesized translation stream immediately!
                playAudioBytes(translatedAudio)
            } else {
                _errorMessage.value = "Translation failed: ${result.exceptionOrNull()?.message}"
                _uiState.value = TranslationUiState.Success
            }
        }
    }

    fun stopRecordingOnly() {
        if (!_isRecording.value) return
        _isRecording.value = false
        _activeRecordingMode.value = null
        cleanupRecorder()
    }

    fun cloneVoice(onSuccess: () -> Unit) {
        val audioBytes = recordingBuffer.toByteArray()
        val userId = authRepository.activeUserId ?: return
        if (audioBytes.isEmpty()) {
            _errorMessage.value = "Please record your voice first before cloning."
            return
        }

        viewModelScope.launch {
            _uiState.value = TranslationUiState.Translating
            val result = translationRepository.cloneVoice(userId, audioBytes)
            if (result.isSuccess) {
                _profile.value = _profile.value?.copy(clonedVoiceKey = result.getOrThrow().clonedVoiceKey)
                _uiState.value = TranslationUiState.Success
                onSuccess()
            } else {
                _errorMessage.value = "Voice cloning failed: ${result.exceptionOrNull()?.message}"
                _uiState.value = TranslationUiState.Success
            }
        }
    }

    fun playVoicePreview(voiceName: String, languageCode: String) {
        viewModelScope.launch {
            _uiState.value = TranslationUiState.Translating
            val result = translationRepository.previewVoice(voiceName, languageCode)
            if (result.isSuccess) {
                playAudioBytes(result.getOrThrow())
                _uiState.value = TranslationUiState.Success
            } else {
                _errorMessage.value = "Failed to play voice preview: ${result.exceptionOrNull()?.message}"
                _uiState.value = TranslationUiState.Success
            }
        }
    }

    fun playAudioBytes(audioBytes: ByteArray?) {
        if (audioBytes == null || audioBytes.isEmpty()) return
        viewModelScope.launch(Dispatchers.IO) {
            try {
                // Fetch the cache directory
                val application = authRepository.context
                val tempFile = File.createTempFile("gte_trans_", ".mp3", application.cacheDir)
                tempFile.deleteOnExit()
                tempFile.writeBytes(audioBytes)

                mediaPlayer?.release()
                mediaPlayer = MediaPlayer().apply {
                    setDataSource(tempFile.absolutePath)
                    prepare()
                    start()
                }
            } catch (e: Exception) {
                _errorMessage.value = "Audio playback failed: ${e.message}"
            }
        }
    }

    private fun cleanupRecorder() {
        try {
            audioRecord?.stop()
            audioRecord?.release()
        } catch (e: Exception) {
            // Already released or stopped
        }
        audioRecord = null
    }

    override fun onCleared() {
        super.onCleared()
        cleanupRecorder()
        mediaPlayer?.release()
        mediaPlayer = null
    }

    companion object {
        val Factory: ViewModelProvider.Factory = object : ViewModelProvider.Factory {
            @Suppress("UNCHECKED_CAST")
            override fun <T : ViewModel> create(modelClass: Class<T>, extras: CreationExtras): T {
                val application = checkNotNull(extras[APPLICATION_KEY]) as GTasksApplication
                return TranslationViewModel(
                    translationRepository = application.container.translationRepository,
                    authRepository = application.container.authRepository
                ) as T
            }
        }
    }
}

// Data Classes & Enums

sealed interface TranslationUiState {
    data object Loading : TranslationUiState
    data object Translating : TranslationUiState
    data object Success : TranslationUiState
    data class Error(val message: String) : TranslationUiState
}

enum class RecordingMode { TALK, LISTEN }
enum class Speaker { ASSOCIATE, CUSTOMER }

data class Utterance(
    val speaker: Speaker,
    val originalText: String,
    val translatedText: String,
    val audioBytes: ByteArray?
)
