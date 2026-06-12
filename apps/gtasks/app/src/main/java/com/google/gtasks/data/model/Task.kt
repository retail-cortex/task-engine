package com.google.gtasks.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class Task(
    @SerialName("ID") val id: String,
    @SerialName("Name") val name: String,
    @SerialName("Description") val description: String? = null,
    @SerialName("TaskType") val taskType: String = "STANDARD",
    @SerialName("Priority") val priority: Int = 3,
    @SerialName("EstimatedDurationMinutes") val estimatedDurationMinutes: Int = 0
)

@Serializable
data class TaskExecution(
    @SerialName("ID") val id: String,
    @SerialName("TaskTemplateID") val taskTemplateID: String,
    @SerialName("Task") val task: Task,
    @SerialName("ExecutionType") val executionType: String = "STANDARD",
    @SerialName("AssigneeID") val assigneeID: String? = null,
    @SerialName("Description") val description: String? = null,
    @SerialName("Status") val status: String = "PENDING",
    @SerialName("Priority") val priority: Int = 3,
    @SerialName("DueAt") val dueAt: String? = null,
    @SerialName("CompletedAt") val completedAt: String? = null,
    @SerialName("ChecklistState") val checklistState: List<ChecklistItem> = emptyList()
)

@Serializable
data class ChecklistItem(
    val step: Int,
    val action: String,
    val required: Boolean = true,
    val completed: Boolean = false,
    @SerialName("completed_at") val completedAt: String? = null
)
