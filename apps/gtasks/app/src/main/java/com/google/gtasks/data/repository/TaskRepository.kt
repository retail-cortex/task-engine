package com.google.gtasks.data.repository

import com.google.gtasks.data.api.ApiClient
import com.google.gtasks.data.api.OverrideAssetRequest
import com.google.gtasks.data.api.UpdateStatusRequest
import com.google.gtasks.data.model.Site
import com.google.gtasks.data.model.TaskExecution
import com.google.gtasks.data.model.Trade
import com.google.gtasks.data.api.ProposeTradeRequest
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

class TaskRepository(
    private val apiClient: ApiClient,
    private val authRepository: AuthRepository
) {
    private val apiService get() = apiClient.apiService
    private val orgId get() = authRepository.activeOrgId

    /**
     * Fetch all active sites (stores) in the current organization.
     */
    suspend fun getSites(): Result<List<Site>> = withContext(Dispatchers.IO) {
        return@withContext try {
            val sites = apiService.getSites(orgId)
            Result.success(sites)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Fetch the complete, prioritized task queue for a specific site.
     */
    suspend fun getSiteTasks(siteId: String): Result<List<TaskExecution>> = withContext(Dispatchers.IO) {
        return@withContext try {
            val tasks = apiService.getSiteTasks(orgId, siteId)
            Result.success(tasks)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Fetch only the tasks assigned to a specific user at a specific site.
     */
    suspend fun getUserTasks(siteId: String, userId: String): Result<List<TaskExecution>> = withContext(Dispatchers.IO) {
        return@withContext try {
            val tasks = apiService.getUserSiteTasks(orgId, siteId, userId)
            Result.success(tasks)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Mutate the status of a task execution (e.g. from PENDING to IN_PROGRESS, or IN_PROGRESS to COMPLETED).
     * Also synchronizes the checklist steps state as a serialized JSON array.
     */
    suspend fun updateTaskStatus(
        siteId: String,
        executionId: String,
        status: String,
        checklistStateJson: String
    ): Result<Unit> = withContext(Dispatchers.IO) {
        return@withContext try {
            apiService.updateTaskStatus(
                orgId = orgId,
                siteId = siteId,
                id = executionId,
                body = UpdateStatusRequest(status = status, checklistState = checklistStateJson)
            )
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Submit a supervisor justification to bypass a hard blocker asset constraint.
     */
    suspend fun overrideAssetConstraint(
        siteId: String,
        executionId: String,
        assetId: String,
        justification: String
    ): Result<Unit> = withContext(Dispatchers.IO) {
        return@withContext try {
            apiService.overrideAsset(
                orgId = orgId,
                siteId = siteId,
                id = executionId,
                body = OverrideAssetRequest(assetId = assetId, justification = justification)
            )
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Claim an unassigned task in the store.
     */
    suspend fun claimTask(siteId: String, executionId: String): Result<Unit> = withContext(Dispatchers.IO) {
        return@withContext try {
            apiService.claimTask(orgId, siteId, executionId)
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Propose a task trade with another colleague.
     */
    suspend fun proposeTrade(siteId: String, taskExecutionId: String, proposedAssigneeId: String): Result<Unit> = withContext(Dispatchers.IO) {
        return@withContext try {
            apiService.proposeTrade(
                orgId = orgId,
                siteId = siteId,
                body = ProposeTradeRequest(taskExecutionId, proposedAssigneeId)
            )
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * List all pending trade offers for the current user.
     */
    suspend fun getPendingTrades(siteId: String): Result<List<Trade>> = withContext(Dispatchers.IO) {
        return@withContext try {
            val trades = apiService.listPendingTrades(orgId, siteId)
            Result.success(trades)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Accept a trade offer.
     */
    suspend fun acceptTrade(siteId: String, tradeId: String): Result<Unit> = withContext(Dispatchers.IO) {
        return@withContext try {
            apiService.acceptTrade(orgId, siteId, tradeId)
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Reject a trade offer.
     */
    suspend fun rejectTrade(siteId: String, tradeId: String): Result<Unit> = withContext(Dispatchers.IO) {
        return@withContext try {
            apiService.rejectTrade(orgId, siteId, tradeId)
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
