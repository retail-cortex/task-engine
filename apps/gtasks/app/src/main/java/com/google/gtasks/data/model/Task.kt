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
    @SerialName("EstimatedDurationMinutes") val estimatedDurationMinutes: Int = 0,
    @SerialName("SOPs") val sops: List<SOP> = emptyList()
)

@Serializable
data class SOP(
    @SerialName("ID") val id: String,
    @SerialName("Title") val title: String,
    @SerialName("CanonicalURL") val canonicalUrl: String? = null
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
    @SerialName("StartedAt") val startedAt: String? = null,
    @SerialName("PausedAt") val pausedAt: String? = null,
    @SerialName("TotalPausedSeconds") val totalPausedSeconds: Int = 0,
    @SerialName("CompletedAt") val completedAt: String? = null,
    @SerialName("ChecklistState") val checklistState: List<ChecklistItem> = emptyList()
)

@Serializable
data class ChecklistItem(
    val step: Int,
    val action: String,
    val required: Boolean = true,
    val completed: Boolean = false,
    val status: String = "PENDING",
    @SerialName("started_at") val startedAt: String? = null,
    @SerialName("paused_at") val pausedAt: String? = null,
    @SerialName("total_paused_seconds") val totalPausedSeconds: Int = 0,
    @SerialName("completed_at") val completedAt: String? = null,
    @SerialName("completed_by_id") val completedById: String? = null,
    @SerialName("slo_seconds") val sloSeconds: Int = 60,
    @SerialName("slo_delta_seconds") val sloDeltaSeconds: Int? = null
)
