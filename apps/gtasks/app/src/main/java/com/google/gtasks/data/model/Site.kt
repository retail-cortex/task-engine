package com.google.gtasks.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class Site(
    @SerialName("id") val id: String,
    @SerialName("name") val name: String,
    @SerialName("organization_id") val organizationId: String
)
