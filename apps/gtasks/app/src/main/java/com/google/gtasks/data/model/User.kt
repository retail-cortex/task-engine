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
