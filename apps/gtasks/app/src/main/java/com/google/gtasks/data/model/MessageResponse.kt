package com.google.gtasks.data.model

import com.google.gtasks.ui.a2ui.A2UITransaction
import kotlinx.serialization.Serializable

@Serializable
data class MessageResponse(
    val id: String,
    val role: String,
    val content: String,
    val a2uiType: String? = null,
    val a2uiData: A2UITransaction? = null
)
