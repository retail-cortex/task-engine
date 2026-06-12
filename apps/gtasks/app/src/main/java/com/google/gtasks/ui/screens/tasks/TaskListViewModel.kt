package com.google.gtasks.ui.screens.tasks

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.ViewModelProvider.AndroidViewModelFactory.Companion.APPLICATION_KEY
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.CreationExtras
import com.google.gtasks.GTasksApplication
import com.google.gtasks.data.model.TaskExecution
import com.google.gtasks.data.model.Trade
import com.google.gtasks.data.repository.AuthRepository
import com.google.gtasks.data.repository.TaskRepository
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch

class TaskListViewModel(
    private val taskRepository: TaskRepository,
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _assignedTasks = MutableStateFlow<List<TaskExecution>>(emptyList())
    private val _availableTasks = MutableStateFlow<List<TaskExecution>>(emptyList())
    private val _pendingTrades = MutableStateFlow<List<Trade>>(emptyList())

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    private val _errorMessage = MutableStateFlow<String?>(null)
    val errorMessage: StateFlow<String?> = _errorMessage.asStateFlow()

    // Filters and sorting View States
    private val _statusFilter = MutableStateFlow(StatusFilter.ALL_ACTIVE)
    val statusFilter: StateFlow<StatusFilter> = _statusFilter.asStateFlow()

    private val _priorityFilter = MutableStateFlow(PriorityFilter.ALL)
    val priorityFilter: StateFlow<PriorityFilter> = _priorityFilter.asStateFlow()

    private val _sortBy = MutableStateFlow(SortBy.PRIORITY)
    val sortBy: StateFlow<SortBy> = _sortBy.asStateFlow()

    // Exposed central UI State combining all data structures reactively!
    private val _uiState = MutableStateFlow<TaskListUiState>(TaskListUiState.Loading)
    val uiState: StateFlow<TaskListUiState> = _uiState.asStateFlow()

    // Expose active site name for the UI banner
    val activeSiteName: String get() = authRepository.activeSiteName ?: "Unknown Site"
    val activeRoleName: String get() = authRepository.currentUser.value?.roles?.firstOrNull()?.name ?: "SITE_ASSOCIATE"
    val currentUserId: String get() = authRepository.activeUserId ?: ""

    // Seeds representing other colleagues in the active store to select for trades
    val colleagues = listOf(
        Colleague("20e8ddca-7ade-49dd-b4b8-310c36e4c486", "Seattle Cashier", "Cashier"),
        StringColleague("b5a2c5ec-8b21-40e6-a8cc-79cc7dda87fb", "Arlington Stocker", "Stock Associate"),
        StringColleague("ef912d7b-7095-4e93-aeeb-ed4a7d339883", "New Orleans Cashier", "Cashier"),
        StringColleague("b75c1a02-c884-40ed-a3f8-8b95f3ff7539", "Dallas Manager (Ryan)", "Store Manager")
    ).filter { it.id != currentUserId } // Exclude yourself

    init {
        loadData()

        // Reactively combine data sets and filters to compile a robust UI state
        val filterStateFlow = combine(_statusFilter, _priorityFilter, _sortBy) { status, priority, sort ->
            FilterState(status, priority, sort)
        }

        val loadingAndErrorFlow = combine(_isLoading, _errorMessage) { loading, error ->
            Pair(loading, error)
        }

        viewModelScope.launch {
            combine(
                _assignedTasks,
                _availableTasks,
                _pendingTrades,
                filterStateFlow,
                loadingAndErrorFlow
            ) { assigned, available, trades, filter, loadErr ->
                val (loading, error) = loadErr
                if (loading && assigned.isEmpty() && available.isEmpty()) return@combine TaskListUiState.Loading
                if (error != null && assigned.isEmpty()) return@combine TaskListUiState.Error(error)

                // Apply filtering and sorting to "My Work" (assigned)
                val filteredAssigned = filterAndSort(assigned, filter)
                
                // Apply filtering and sorting to "Available" (unassigned)
                val filteredAvailable = filterAndSort(available, filter)

                TaskListUiState.Success(
                    assignedTasks = filteredAssigned,
                    availableTasks = filteredAvailable,
                    pendingTrades = trades
                )
            }.collect { newState ->
                _uiState.value = newState
            }
        }
    }

    private fun filterAndSort(tasks: List<TaskExecution>, filter: FilterState): List<TaskExecution> {
        // 1. Filter completed vs active tasks
        var filtered = if (filter.status == StatusFilter.COMPLETED) {
            tasks.filter { it.status == "COMPLETED" }
        } else {
            tasks.filter { it.status != "COMPLETED" }
        }

        // 2. Apply status filter
        filtered = when (filter.status) {
            StatusFilter.ALL_ACTIVE -> filtered
            StatusFilter.PENDING -> filtered.filter { it.status == "PENDING" }
            StatusFilter.IN_PROGRESS -> filtered.filter { it.status == "IN_PROGRESS" }
            StatusFilter.COMPLETED -> filtered
        }

        // 3. Apply priority filter
        filtered = when (filter.priority) {
            PriorityFilter.ALL -> filtered
            PriorityFilter.HIGH -> filtered.filter { it.priority == 1 }
            PriorityFilter.MEDIUM -> filtered.filter { it.priority == 2 }
            PriorityFilter.LOW -> filtered.filter { it.priority >= 3 }
        }

        // 4. Apply sorting
        return when (filter.sortBy) {
            SortBy.PRIORITY -> filtered.sortedBy { it.priority }
            SortBy.CREATION_TIME -> filtered.sortedByDescending { it.id }
            SortBy.STATUS -> filtered.sortedBy { it.status }
        }
    }

    /**
     * Parallel-fetch all operational datasets (My Work, Available Site Tasks, Pending Trades)
     */
    fun loadData() {
        val siteId = authRepository.activeSiteId
        val userId = authRepository.activeUserId
        if (siteId == null || userId == null) {
            _errorMessage.value = "Missing storefront site or user context."
            return
        }

        viewModelScope.launch {
            _isLoading.value = true
            _errorMessage.value = null

            try {
                // 1. Fetch tasks assigned specifically to you ("My Work")
                val assignedResult = taskRepository.getUserTasks(siteId, userId)
                
                // 2. Fetch all tasks for the site, and filter for unassigned ("Available" to claim)
                val siteTasksResult = taskRepository.getSiteTasks(siteId)
                
                // 3. Fetch pending trades
                val tradesResult = taskRepository.getPendingTrades(siteId)

                assignedResult.onSuccess { tasks ->
                    _assignedTasks.value = tasks
                }.onFailure { e ->
                    _errorMessage.value = "Failed to load your tasks: ${e.localizedMessage}"
                }

                siteTasksResult.onSuccess { tasks ->
                    // Available tasks are active tasks where the assignee is unassigned (null or empty)
                    val unassigned = tasks.filter { it.assigneeID.isNullOrEmpty() }
                    _availableTasks.value = unassigned
                }.onFailure { e ->
                    if (_errorMessage.value == null) {
                        _errorMessage.value = "Failed to load store available tasks: ${e.localizedMessage}"
                    }
                }

                tradesResult.onSuccess { trades ->
                    _pendingTrades.value = trades
                }.onFailure { e ->
                    // Fail silently or log (trades are secondary to core task queues)
                }

            } catch (e: Exception) {
                _errorMessage.value = e.localizedMessage
            } finally {
                _isLoading.value = false
            }
        }
    }

    /**
     * Claim/Take an unassigned task
     */
    fun claimTask(executionId: String) {
        val siteId = authRepository.activeSiteId ?: return
        viewModelScope.launch {
            _isLoading.value = true
            val result = taskRepository.claimTask(siteId, executionId)
            result.onSuccess {
                loadData() // Reload all queues to trigger dynamic animated shifting
            }.onFailure { e ->
                _errorMessage.value = "Failed to claim task: ${e.localizedMessage}"
                _isLoading.value = false
            }
        }
    }

    /**
     * Propose a shift/task trade to a colleague
     */
    fun proposeTrade(taskExecutionId: String, colleagueId: String) {
        val siteId = authRepository.activeSiteId ?: return
        viewModelScope.launch {
            _isLoading.value = true
            val result = taskRepository.proposeTrade(siteId, taskExecutionId, colleagueId)
            result.onSuccess {
                loadData() // Reload
            }.onFailure { e ->
                _errorMessage.value = "Failed to propose trade: ${e.localizedMessage}"
                _isLoading.value = false
            }
        }
    }

    /**
     * Accept a pending trade offer from a peer
     */
    fun acceptTrade(tradeId: String) {
        val siteId = authRepository.activeSiteId ?: return
        viewModelScope.launch {
            _isLoading.value = true
            val result = taskRepository.acceptTrade(siteId, tradeId)
            result.onSuccess {
                loadData() // Reload
            }.onFailure { e ->
                _errorMessage.value = "Failed to accept trade: ${e.localizedMessage}"
                _isLoading.value = false
            }
        }
    }

    /**
     * Reject a pending trade offer
     */
    fun rejectTrade(tradeId: String) {
        val siteId = authRepository.activeSiteId ?: return
        viewModelScope.launch {
            _isLoading.value = true
            val result = taskRepository.rejectTrade(siteId, tradeId)
            result.onSuccess {
                loadData() // Reload
            }.onFailure { e ->
                _errorMessage.value = "Failed to reject trade: ${e.localizedMessage}"
                _isLoading.value = false
            }
        }
    }

    // Filter/Sort Setters
    fun setStatusFilter(filter: StatusFilter) { _statusFilter.value = filter }
    fun setPriorityFilter(filter: PriorityFilter) { _priorityFilter.value = filter }
    fun setSortBy(sort: SortBy) { _sortBy.value = sort }

    companion object {
        val Factory: ViewModelProvider.Factory = object : ViewModelProvider.Factory {
            @Suppress("UNCHECKED_CAST")
            override fun <T : ViewModel> create(modelClass: Class<T>, extras: CreationExtras): T {
                val application = checkNotNull(extras[APPLICATION_KEY]) as GTasksApplication
                return TaskListViewModel(
                    taskRepository = application.container.taskRepository,
                    authRepository = application.container.authRepository
                ) as T
            }
        }
    }
}

// Data Classes & Enums

sealed interface TaskListUiState {
    data object Loading : TaskListUiState
    data class Success(
        val assignedTasks: List<TaskExecution>,
        val availableTasks: List<TaskExecution>,
        val pendingTrades: List<Trade>
    ) : TaskListUiState
    data class Error(val message: String) : TaskListUiState
}

enum class StatusFilter { ALL_ACTIVE, PENDING, IN_PROGRESS, COMPLETED }
enum class PriorityFilter { ALL, HIGH, MEDIUM, LOW }
enum class SortBy { PRIORITY, CREATION_TIME, STATUS }

private data class FilterState(
    val status: StatusFilter,
    val priority: PriorityFilter,
    val sortBy: SortBy
)

data class Colleague(
    val id: String,
    val name: String,
    val roleName: String
)
// Helper constructor for easy clean listing
private fun StringColleague(id: String, name: String, roleName: String) = Colleague(id, name, roleName)
