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

package com.google.gtasks.ui.screens.detail

import androidx.compose.runtime.mutableStateListOf
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.ViewModelProvider.AndroidViewModelFactory.Companion.APPLICATION_KEY
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.CreationExtras
import com.google.gtasks.GTasksApplication
import com.google.gtasks.data.repository.AuthRepository
import com.google.gtasks.data.repository.TaskRepository
import com.google.gtasks.data.model.ChecklistItem
import com.google.gtasks.data.model.TaskExecution
import com.google.gtasks.ui.screens.tasks.Colleague
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import java.text.SimpleDateFormat
import java.util.*

class TaskDetailViewModel(
    private val taskRepository: TaskRepository,
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<TaskDetailUiState>(TaskDetailUiState.Loading)
    val uiState: StateFlow<TaskDetailUiState> = _uiState.asStateFlow()

    // Interactive checklist items list
    val checklist = mutableStateListOf<ChecklistItem>()

    private val _isCompletable = MutableStateFlow(false)
    val isCompletable: StateFlow<Boolean> = _isCompletable.asStateFlow()

    private val _isSubmitting = MutableStateFlow(false)
    val isSubmitting: StateFlow<Boolean> = _isSubmitting.asStateFlow()

    val siteId: String
        get() = authRepository.activeSiteId ?: ""

    val siteName: String
        get() = authRepository.activeSiteName ?: "Unknown Store"

    fun loadTaskDetails(executionId: String, silent: Boolean = false) {
        if (!silent) {
            _uiState.value = TaskDetailUiState.Loading
        }
        viewModelScope.launch {
            // Fetch active on-shift colleagues in parallel
            launch {
                val associatesResult = taskRepository.getSiteAssociates(siteId)
                associatesResult.onSuccess { users ->
                    val mapped = users.filter { it.id != currentUserId }.map { user ->
                        Colleague(
                            id = user.id,
                            name = user.name ?: "Unknown Associate",
                            roleName = user.roles.firstOrNull()?.name ?: "Associate"
                        )
                    }
                    _colleagues.value = mapped
                }
            }

            val result = taskRepository.getSiteTasks(siteId)
            result.onSuccess { tasks ->
                val task = tasks.find { it.id == executionId }
                if (task != null) {
                    checklist.clear()
                    checklist.addAll(task.checklistState)
                    checkIfCompletable()
                    _uiState.value = TaskDetailUiState.Success(task)
                } else {
                    _uiState.value = TaskDetailUiState.Error("Task not found in active queue.")
                }
            }.onFailure { error ->
                _uiState.value = TaskDetailUiState.Error(error.localizedMessage ?: "Failed loading task details")
            }
        }
    }

    /**
     * Reactively check if all required steps in the checklist are completed.
     */
    private fun checkIfCompletable() {
        val allRequiredDone = checklist.all { !it.required || it.completed }
        _isCompletable.value = allRequiredDone && checklist.isNotEmpty()
    }

    // --- Task-Level State Machine Drivers ---

    fun startTask(executionId: String) {
        viewModelScope.launch {
            _isSubmitting.value = true
            val result = taskRepository.updateTaskStatus(siteId, executionId, "IN_PROGRESS", "")
            _isSubmitting.value = false
            result.onSuccess {
                loadTaskDetails(executionId, silent = true)
            }.onFailure { error ->
                _uiState.value = TaskDetailUiState.Error(error.localizedMessage ?: "Failed starting task")
            }
        }
    }

    fun pauseTask(executionId: String) {
        viewModelScope.launch {
            _isSubmitting.value = true
            val result = taskRepository.updateTaskStatus(siteId, executionId, "PAUSED", "")
            _isSubmitting.value = false
            result.onSuccess {
                loadTaskDetails(executionId, silent = true)
            }.onFailure { error ->
                _uiState.value = TaskDetailUiState.Error(error.localizedMessage ?: "Failed pausing task")
            }
        }
    }

    fun resumeTask(executionId: String) {
        viewModelScope.launch {
            _isSubmitting.value = true
            val result = taskRepository.updateTaskStatus(siteId, executionId, "IN_PROGRESS", "")
            _isSubmitting.value = false
            result.onSuccess {
                loadTaskDetails(executionId, silent = true)
            }.onFailure { error ->
                _uiState.value = TaskDetailUiState.Error(error.localizedMessage ?: "Failed resuming task")
            }
        }
    }

    // --- Step-Level State Machine Drivers (Option A) ---

    fun startStep(executionId: String, stepIndex: Int) {
        viewModelScope.launch {
            val delta = ChecklistDelta(step = stepIndex, action = "START", completed = false)
            val deltaJson = Json.encodeToString(delta)
            taskRepository.updateTaskStatus(siteId, executionId, "IN_PROGRESS", deltaJson)
            loadTaskDetails(executionId, silent = true)
        }
    }

    fun pauseStep(executionId: String, stepIndex: Int) {
        viewModelScope.launch {
            val delta = ChecklistDelta(step = stepIndex, action = "PAUSE", completed = false)
            val deltaJson = Json.encodeToString(delta)
            taskRepository.updateTaskStatus(siteId, executionId, "IN_PROGRESS", deltaJson)
            loadTaskDetails(executionId, silent = true)
        }
    }

    fun resumeStep(executionId: String, stepIndex: Int) {
        viewModelScope.launch {
            val delta = ChecklistDelta(step = stepIndex, action = "RESUME", completed = false)
            val deltaJson = Json.encodeToString(delta)
            taskRepository.updateTaskStatus(siteId, executionId, "IN_PROGRESS", deltaJson)
            loadTaskDetails(executionId, silent = true)
        }
    }

    fun completeStep(executionId: String, stepIndex: Int) {
        viewModelScope.launch {
            val delta = ChecklistDelta(step = stepIndex, action = "COMPLETE", completed = true)
            val deltaJson = Json.encodeToString(delta)
            taskRepository.updateTaskStatus(siteId, executionId, "IN_PROGRESS", deltaJson)
            loadTaskDetails(executionId, silent = true)
        }
    }

    /**
     * Transition task to completed state.
     */
    fun completeTask(executionId: String, onCompleteSuccess: () -> Unit) {
        _isSubmitting.value = true
        viewModelScope.launch {
            val checklistJson = Json.encodeToString(checklist.toList())
            val result = taskRepository.updateTaskStatus(
                siteId = siteId,
                executionId = executionId,
                status = "COMPLETED",
                checklistStateJson = checklistJson
            )
            _isSubmitting.value = false
            result.onSuccess {
                onCompleteSuccess()
            }.onFailure { error ->
                // Handle failure (re-surface details)
                _uiState.value = TaskDetailUiState.Error(error.localizedMessage ?: "Failed completing task")
            }
        }
    }

    /**
     * Submit a supervisor justification to bypass a hard blocker asset constraint.
     */
    fun submitAssetOverride(executionId: String, assetId: String, justification: String) {
        _isSubmitting.value = true
        viewModelScope.launch {
            val result = taskRepository.overrideAssetConstraint(
                siteId = siteId,
                executionId = executionId,
                assetId = assetId,
                justification = justification
            )
            _isSubmitting.value = false
            result.onSuccess {
                // Reload details to reflect override
                loadTaskDetails(executionId, silent = true)
            }.onFailure { error ->
                _uiState.value = TaskDetailUiState.Error(error.localizedMessage ?: "Failed submitting override")
            }
        }
    }

    val currentUserId: String get() = authRepository.activeUserId ?: ""

    val activeRoleName: String
        get() = authRepository.currentUser.value?.roles?.firstOrNull()?.name ?: "SITE_ASSOCIATE"

    // Active colleagues in the store fetched dynamically from backend
    private val _colleagues = MutableStateFlow<List<Colleague>>(emptyList())
    val colleagues: StateFlow<List<Colleague>> = _colleagues.asStateFlow()

    fun proposeTrade(taskExecutionId: String, colleagueId: String, onSuccess: () -> Unit) {
        viewModelScope.launch {
            _isSubmitting.value = true
            val result = taskRepository.proposeTrade(siteId, taskExecutionId, colleagueId)
            _isSubmitting.value = false
            result.onSuccess {
                onSuccess()
            }
        }
    }

    companion object {
        val Factory: ViewModelProvider.Factory = object : ViewModelProvider.Factory {
            @Suppress("UNCHECKED_CAST")
            override fun <T : ViewModel> create(modelClass: Class<T>, extras: CreationExtras): T {
                val application = checkNotNull(extras[APPLICATION_KEY]) as GTasksApplication
                return TaskDetailViewModel(
                    taskRepository = application.container.taskRepository,
                    authRepository = application.container.authRepository
                ) as T
            }
        }
    }
}

sealed interface TaskDetailUiState {
    data object Loading : TaskDetailUiState
    data class Success(val task: TaskExecution) : TaskDetailUiState
    data class Error(val message: String) : TaskDetailUiState
}

@Serializable
data class ChecklistDelta(
    val step: Int,
    val action: String,
    val completed: Boolean
)
