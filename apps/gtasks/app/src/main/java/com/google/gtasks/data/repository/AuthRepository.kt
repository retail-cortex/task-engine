package com.google.gtasks.data.repository

import android.content.Context
import android.content.SharedPreferences
import com.google.gtasks.data.api.ApiClient
import com.google.gtasks.data.model.UserDTO
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

class AuthRepository(
    private val context: Context,
    private val apiClient: ApiClient
) {
    private val prefs: SharedPreferences = context.getSharedPreferences("gtasks_prefs", Context.MODE_PRIVATE)

    companion object {
        private const val PREF_USER_ID = "active_user_id"
        private const val PREF_USER_NAME = "active_user_name"
        private const val PREF_USER_EMAIL = "active_user_email"
        private const val PREF_ORG_ID = "active_org_id"
        private const val PREF_SITE_ID = "active_site_id"
        private const val PREF_SITE_NAME = "active_site_name"
        
        // Default organization ID from Dallas Store #1000 operational scopes
        const val DEFAULT_ORG_ID = "33333333-3333-3333-3333-333333333333"
    }

    private val _isLoggedIn = MutableStateFlow(apiClient.token != null)
    val isLoggedIn: StateFlow<Boolean> = _isLoggedIn.asStateFlow()

    private val _currentUser = MutableStateFlow<UserDTO?>(null)
    val currentUser: StateFlow<UserDTO?> = _currentUser.asStateFlow()

    var activeOrgId: String
        get() = prefs.getString(PREF_ORG_ID, DEFAULT_ORG_ID) ?: DEFAULT_ORG_ID
        set(value) {
            prefs.edit().putString(PREF_ORG_ID, value).apply()
        }

    var activeSiteId: String?
        get() = prefs.getString(PREF_SITE_ID, null)
        set(value) {
            prefs.edit().putString(PREF_SITE_ID, value).apply()
        }

    var activeSiteName: String?
        get() = prefs.getString(PREF_SITE_NAME, null)
        set(value) {
            prefs.edit().putString(PREF_SITE_NAME, value).apply()
        }

    var activeUserId: String?
        get() = prefs.getString(PREF_USER_ID, null)
        private set(value) {
            prefs.edit().putString(PREF_USER_ID, value).apply()
        }

    init {
        // Recover cached user profile if token exists
        if (apiClient.token != null) {
            val cachedId = prefs.getString(PREF_USER_ID, null)
            val cachedName = prefs.getString(PREF_USER_NAME, null)
            val cachedEmail = prefs.getString(PREF_USER_EMAIL, null)
            if (cachedId != null && cachedEmail != null) {
                _currentUser.value = UserDTO(
                    id = cachedId,
                    name = cachedName,
                    email = cachedEmail
                )
            }
        }
    }

    /**
     * Authenticate using a Google ID Token (JWT) obtained from Google Sign-In.
     */
    suspend fun loginWithGoogleToken(idToken: String, orgId: String = DEFAULT_ORG_ID): Result<UserDTO> {
        return try {
            apiClient.token = idToken
            activeOrgId = orgId
            val profile = apiClient.apiService.getMe(orgId)
            
            // Persist session context
            activeUserId = profile.id
            prefs.edit()
                .putString(PREF_USER_NAME, profile.name)
                .putString(PREF_USER_EMAIL, profile.email)
                .apply()
                
            _currentUser.value = profile
            _isLoggedIn.value = true
            Result.success(profile)
        } catch (e: Exception) {
            logout()
            Result.failure(e)
        }
    }

    /**
     * Development/Offline bypass login. Sets a pre-seeded user ID as the Bearer token.
     * The Go backend's development middleware accepts this user ID directly.
     */
    suspend fun loginWithBypassUserId(bypassUserId: String, orgId: String = DEFAULT_ORG_ID): Result<UserDTO> {
        return try {
            apiClient.token = bypassUserId
            activeOrgId = orgId
            val profile = apiClient.apiService.getMe(orgId)
            
            // Persist session context
            activeUserId = profile.id
            prefs.edit()
                .putString(PREF_USER_NAME, profile.name)
                .putString(PREF_USER_EMAIL, profile.email)
                .apply()
                
            _currentUser.value = profile
            _isLoggedIn.value = true
            Result.success(profile)
        } catch (e: Exception) {
            logout()
            Result.failure(e)
        }
    }

    fun logout() {
        apiClient.clearSession()
        activeUserId = null
        activeSiteId = null
        prefs.edit()
            .remove(PREF_USER_NAME)
            .remove(PREF_USER_EMAIL)
            .remove(PREF_SITE_ID)
            .apply()
        _currentUser.value = null
        _isLoggedIn.value = false
    }
}
