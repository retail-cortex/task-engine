package com.google.gtasks.data.api

import com.google.gtasks.data.model.*
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import retrofit2.http.*

interface ApiInterface {

    @GET("api/v1/organizations/{orgId}/me")
    suspend fun getMe(
        @Path("orgId") orgId: String
    ): UserDTO

    @GET("api/v1/organizations/{orgId}/sites")
    suspend fun getSites(
        @Path("orgId") orgId: String
    ): List<Site>

    @GET("api/v1/organizations/{orgId}/sites/{siteId}/tasks")
    suspend fun getSiteTasks(
        @Path("orgId") orgId: String,
        @Path("siteId") siteId: String
    ): List<TaskExecution>

    @GET("api/v1/organizations/{orgId}/sites/{siteId}/users/{userId}/tasks")
    suspend fun getUserSiteTasks(
        @Path("orgId") orgId: String,
        @Path("siteId") siteId: String,
        @Path("userId") userId: String
    ): List<TaskExecution>

    @PATCH("api/v1/organizations/{orgId}/sites/{siteId}/tasks/{id}/status")
    suspend fun updateTaskStatus(
        @Path("orgId") orgId: String,
        @Path("siteId") siteId: String,
        @Path("id") id: String,
        @Body body: UpdateStatusRequest
    ): Map<String, String>

    @POST("api/v1/organizations/{orgId}/sites/{siteId}/tasks/{id}/override")
    suspend fun overrideAsset(
        @Path("orgId") orgId: String,
        @Path("siteId") siteId: String,
        @Path("id") id: String,
        @Body body: OverrideAssetRequest
    ): Map<String, String>

    @POST("api/v1/organizations/{orgId}/sites/{siteId}/users/{userId}/sessions/shift/{shiftId}/message")
    suspend fun sendMessage(
        @Path("orgId") orgId: String,
        @Path("siteId") siteId: String,
        @Path("userId") userId: String,
        @Path("shiftId") shiftId: String,
        @Body body: MessageRequest
    ): MessageResponse

    @POST("api/v1/organizations/{orgId}/sites/{siteId}/tasks/{id}/claim")
    suspend fun claimTask(
        @Path("orgId") orgId: String,
        @Path("siteId") siteId: String,
        @Path("id") id: String
    ): Map<String, String>

    @POST("api/v1/organizations/{orgId}/sites/{siteId}/trades")
    suspend fun proposeTrade(
        @Path("orgId") orgId: String,
        @Path("siteId") siteId: String,
        @Body body: ProposeTradeRequest
    ): Map<String, String>

    @GET("api/v1/organizations/{orgId}/sites/{siteId}/trades")
    suspend fun listPendingTrades(
        @Path("orgId") orgId: String,
        @Path("siteId") siteId: String
    ): List<Trade>

    @POST("api/v1/organizations/{orgId}/sites/{siteId}/trades/{tradeId}/accept")
    suspend fun acceptTrade(
        @Path("orgId") orgId: String,
        @Path("siteId") siteId: String,
        @Path("tradeId") tradeId: String
    ): Map<String, String>

    @POST("api/v1/organizations/{orgId}/sites/{siteId}/trades/{tradeId}/reject")
    suspend fun rejectTrade(
        @Path("orgId") orgId: String,
        @Path("siteId") siteId: String,
        @Path("tradeId") tradeId: String
    ): Map<String, String>

    @GET("api/v1/organizations/{orgId}/sites/{siteId}/associates")
    suspend fun getSiteAssociates(
        @Path("orgId") orgId: String,
        @Path("siteId") siteId: String
    ): List<UserDTO>
}

// Request DTOs

@Serializable
data class UpdateStatusRequest(
    val status: String,
    @SerialName("checklist_state") val checklistState: String
)

@Serializable
data class OverrideAssetRequest(
    @SerialName("asset_id") val assetId: String,
    val justification: String
)

@Serializable
data class MessageRequest(
    val message: String
)

@Serializable
data class ProposeTradeRequest(
    @SerialName("task_execution_id") val taskExecutionId: String,
    @SerialName("proposed_assignee_id") val proposedAssigneeId: String
)
