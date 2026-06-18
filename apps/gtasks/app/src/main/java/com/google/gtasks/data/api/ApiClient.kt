package com.google.gtasks.data.api

import android.content.Context
import android.content.SharedPreferences
import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit

class ApiClient(private val context: Context) {

    private val prefs: SharedPreferences = context.getSharedPreferences("gtasks_prefs", Context.MODE_PRIVATE)

    companion object {
        // Standard Android Emulator loopback to host machine's localhost where Go API server runs
        private const val DEFAULT_BASE_URL = "http://10.0.2.2:8080/"
        private const val PREF_JWT_TOKEN = "jwt_token"
        private const val PREF_BASE_URL = "api_base_url"
    }

    // Reactive Token Getter and Setter
    var token: String?
        get() = prefs.getString(PREF_JWT_TOKEN, null)
        set(value) {
            prefs.edit().putString(PREF_JWT_TOKEN, value).apply()
        }

    // Dynamic Base URL Configuration
    var baseUrl: String
        get() = prefs.getString(PREF_BASE_URL, DEFAULT_BASE_URL) ?: DEFAULT_BASE_URL
        set(value) {
            val formattedUrl = if (value.endsWith("/")) value else "$value/"
            prefs.edit().putString(PREF_BASE_URL, formattedUrl).apply()
            // Reset retrofit instance when base URL is changed
            _apiService = null
        }

    var onUnauthorizedListener: (() -> Unit)? = null

    fun clearSession() {
        prefs.edit().remove(PREF_JWT_TOKEN).apply()
    }

    private val okHttpClient: OkHttpClient by lazy {
        val loggingInterceptor = HttpLoggingInterceptor().apply {
            level = HttpLoggingInterceptor.Level.BODY
        }

        OkHttpClient.Builder()
            .addInterceptor { chain ->
                val requestBuilder = chain.request().newBuilder()
                token?.let { t ->
                    requestBuilder.addHeader("Authorization", "Bearer $t")
                }
                val response = chain.proceed(requestBuilder.build())
                if (response.code == 401) {
                    onUnauthorizedListener?.invoke()
                }
                response
            }
            .addInterceptor(loggingInterceptor)
            .build()
    }

    private val json: Json by lazy {
        Json {
            ignoreUnknownKeys = true
            coerceInputValues = true
            prettyPrint = true
            isLenient = true
        }
    }

    private var _apiService: ApiInterface? = null

    val apiService: ApiInterface
        get() {
            if (_apiService == null) {
                val contentType = "application/json".toMediaType()
                val retrofit = Retrofit.Builder()
                    .baseUrl(baseUrl)
                    .client(okHttpClient)
                    .addConverterFactory(json.asConverterFactory(contentType))
                    .build()
                _apiService = retrofit.create(ApiInterface::class.java)
            }
            return _apiService!!
        }
}
