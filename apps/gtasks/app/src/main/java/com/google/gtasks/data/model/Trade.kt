package com.google.gtasks.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class Trade(
    @SerialName("ID") val id: String,
    @SerialName("TaskExecutionID") val taskExecutionId: String,
    @SerialName("InitiatorID") val initiatorId: String,
    @SerialName("ProposedAssigneeID") val proposedAssigneeId: String,
    @SerialName("Status") val status: String = "PENDING",
    @SerialName("CreatedAt") val createdAt: String? = null
)
