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

package com.google.gtasks.ui.screens.login

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.ViewModelProvider.AndroidViewModelFactory.Companion.APPLICATION_KEY
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.CreationExtras
import com.google.gtasks.GTasksApplication
import com.google.gtasks.data.repository.AuthRepository
import com.google.gtasks.data.model.UserDTO
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class LoginViewModel(private val authRepository: AuthRepository) : ViewModel() {

    private val _uiState = MutableStateFlow<LoginUiState>(LoginUiState.Idle)
    val uiState: StateFlow<LoginUiState> = _uiState.asStateFlow()

    // List of pre-seeded developer bypass user IDs for quick testing
    val seedUsers = listOf(
        SeedUser("Dallas Store Manager", "b75c1a02-c884-40ed-a3f8-8b95f3ff7539"),
        SeedUser("Seattle Cashier Colleague", "20e8ddca-7ade-49dd-b4b8-310c36e4c486"),
        SeedUser("Arlington Stock Associate", "b5a2c5ec-8b21-40e6-a8cc-79cc7dda87fb"),
        SeedUser("New Orleans Cashier", "ef912d7b-7095-4e93-aeeb-ed4a7d339883")
    )

    fun loginWithGoogle(idToken: String) {
        _uiState.value = LoginUiState.Loading
        viewModelScope.launch {
            val result = authRepository.loginWithGoogleToken(idToken)
            result.onSuccess { user ->
                _uiState.value = LoginUiState.Success(user)
            }.onFailure { error ->
                _uiState.value = LoginUiState.Error(error.localizedMessage ?: "Google authentication failed")
            }
        }
    }

    fun loginWithBypass(userId: String) {
        _uiState.value = LoginUiState.Loading
        viewModelScope.launch {
            val result = authRepository.loginWithBypassUserId(userId)
            result.onSuccess { user ->
                _uiState.value = LoginUiState.Success(user)
            }.onFailure { error ->
                _uiState.value = LoginUiState.Error(error.localizedMessage ?: "Bypass login failed")
            }
        }
    }

    fun setError(message: String) {
        _uiState.value = LoginUiState.Error(message)
    }

    fun resetState() {
        _uiState.value = LoginUiState.Idle
    }

    companion object {
        val Factory: ViewModelProvider.Factory = object : ViewModelProvider.Factory {
            @Suppress("UNCHECKED_CAST")
            override fun <T : ViewModel> create(modelClass: Class<T>, extras: CreationExtras): T {
                val application = checkNotNull(extras[APPLICATION_KEY]) as GTasksApplication
                return LoginViewModel(application.container.authRepository) as T
            }
        }
    }
}

sealed interface LoginUiState {
    data object Idle : LoginUiState
    data object Loading : LoginUiState
    data class Success(val user: UserDTO) : LoginUiState
    data class Error(val message: String) : LoginUiState
}

data class SeedUser(val displayName: String, val uuid: String)
