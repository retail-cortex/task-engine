package com.google.gtasks.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class User(
    @SerialName("id") val id: String,
    @SerialName("email") val email: String,
    @SerialName("name") val name: String? = null,
    @SerialName("oauth_provider") val oAuthProvider: String? = null,
    @SerialName("oauth_id") val oAuthId: String? = null
)

@Serializable
data class Role(
    @SerialName("id") val id: String,
    @SerialName("name") val name: String
)

@Serializable
data class Organization(
    @SerialName("id") val id: String,
    @SerialName("name") val name: String
)

@Serializable
data class UserDTO(
    @SerialName("id") val id: String,
    @SerialName("name") val name: String? = null,
    @SerialName("email") val email: String,
    @SerialName("roles") val roles: List<Role> = emptyList(),
    @SerialName("organizations") val organizations: List<Organization> = emptyList(),
    @SerialName("sites") val sites: List<Site> = emptyList()
)
