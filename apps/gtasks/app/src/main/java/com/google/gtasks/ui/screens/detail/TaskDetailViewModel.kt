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

    fun loadTaskDetails(executionId: String) {
        _uiState.value = TaskDetailUiState.Loading
        viewModelScope.launch {
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

    /**
     * Toggles a checklist step completion state and immediately synchronizes the new state to the database!
     */
    fun toggleChecklistItem(executionId: String, stepIndex: Int) {
        val index = checklist.indexOfFirst { it.step == stepIndex }
        if (index == -1) return

        val item = checklist[index]
        val nowStr = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss'Z'", Locale.US).apply {
            timeZone = TimeZone.getTimeZone("UTC")
        }.format(Date())

        val updatedItem = item.copy(
            completed = !item.completed,
            completedAt = if (!item.completed) nowStr else null
        )
        
        checklist[index] = updatedItem
        checkIfCompletable()

        // Sync back to database immediately (marking as IN_PROGRESS to preserve state)
        viewModelScope.launch {
            val checklistJson = Json.encodeToString(checklist.toList())
            taskRepository.updateTaskStatus(
                siteId = siteId,
                executionId = executionId,
                status = "IN_PROGRESS",
                checklistStateJson = checklistJson
            )
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
                loadTaskDetails(executionId)
            }.onFailure { error ->
                _uiState.value = TaskDetailUiState.Error(error.localizedMessage ?: "Failed submitting override")
            }
        }
    }

    val currentUserId: String get() = authRepository.activeUserId ?: ""

    val colleagues = listOf(
        Colleague("20e8ddca-7ade-49dd-b4b8-310c36e4c486", "Seattle Cashier", "Cashier"),
        Colleague("b5a2c5ec-8b21-40e6-a8cc-79cc7dda87fb", "Arlington Stocker", "Stock Associate"),
        Colleague("ef912d7b-7095-4e93-aeeb-ed4a7d339883", "New Orleans Cashier", "Cashier"),
        Colleague("b75c1a02-c884-40ed-a3f8-8b95f3ff7539", "Dallas Manager (Ryan)", "Store Manager")
    ).filter { it.id != currentUserId }

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
