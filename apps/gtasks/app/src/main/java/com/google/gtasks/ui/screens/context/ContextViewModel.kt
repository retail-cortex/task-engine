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

package com.google.gtasks.ui.screens.context

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.ViewModelProvider.AndroidViewModelFactory.Companion.APPLICATION_KEY
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.CreationExtras
import com.google.gtasks.GTasksApplication
import com.google.gtasks.data.repository.AuthRepository
import com.google.gtasks.data.repository.TaskRepository
import com.google.gtasks.data.model.Site
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class ContextViewModel(
    private val taskRepository: TaskRepository,
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<ContextUiState>(ContextUiState.Loading)
    val uiState: StateFlow<ContextUiState> = _uiState.asStateFlow()

    init {
        loadSites()
    }

    fun loadSites() {
        _uiState.value = ContextUiState.Loading
        viewModelScope.launch {
            val result = taskRepository.getSites()
            result.onSuccess { sites ->
                if (sites.isEmpty()) {
                    _uiState.value = ContextUiState.Error("No stores found for this organization.")
                } else {
                    _uiState.value = ContextUiState.Success(sites)
                }
            }.onFailure { error ->
                _uiState.value = ContextUiState.Error(error.localizedMessage ?: "Failed loading stores")
            }
        }
    }

    fun selectSite(site: Site) {
        authRepository.activeSiteId = site.id
        authRepository.activeSiteName = site.name
    }

    companion object {
        val Factory: ViewModelProvider.Factory = object : ViewModelProvider.Factory {
            @Suppress("UNCHECKED_CAST")
            override fun <T : ViewModel> create(modelClass: Class<T>, extras: CreationExtras): T {
                val application = checkNotNull(extras[APPLICATION_KEY]) as GTasksApplication
                return ContextViewModel(
                    taskRepository = application.container.taskRepository,
                    authRepository = application.container.authRepository
                ) as T
            }
        }
    }
}

sealed interface ContextUiState {
    data object Loading : ContextUiState
    data class Success(val sites: List<Site>) : ContextUiState
    data class Error(val message: String) : ContextUiState
}
