package com.storytrim.app.core.network

import android.content.Context
import android.util.Log
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.runBlocking
import okhttp3.Interceptor
import okhttp3.Response
import java.io.IOException
import javax.inject.Inject
import javax.inject.Singleton

private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "auth_prefs")
private val TOKEN_KEY = stringPreferencesKey("auth_token")
private val USERNAME_KEY = stringPreferencesKey("auth_username")

@Singleton
class AuthInterceptor @Inject constructor(
    @ApplicationContext private val context: Context
) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val requestBuilder = chain.request().newBuilder()
        
        // 使用 runBlocking 确保拦截器能拿到 token，虽然有性能损耗但在 OkHttp 线程池中尚可接受
        // 增加异常捕获确保不会 crash
        val token = try {
            runBlocking { readToken() }
        } catch (e: Exception) {
            Log.e("AuthInterceptor", "Failed to read token", e)
            ""
        }
        
        if (token.isNotEmpty()) {
            requestBuilder.addHeader("Authorization", "Bearer $token")
        }
        
        return chain.proceed(requestBuilder.build())
    }

    suspend fun saveToken(token: String) {
        try {
            context.dataStore.edit { preferences ->
                preferences[TOKEN_KEY] = token
            }
            Log.d("AuthInterceptor", "Token saved: ${token.take(10)}...")
        } catch (e: Exception) {
            Log.e("AuthInterceptor", "Failed to save token", e)
        }
    }

    suspend fun saveUsername(username: String) {
        try {
            context.dataStore.edit { preferences ->
                preferences[USERNAME_KEY] = username
            }
        } catch (e: Exception) {
            Log.e("AuthInterceptor", "Failed to save username", e)
        }
    }

    suspend fun getToken(): String = readToken()

    private suspend fun readToken(): String {
        return try {
            val preferences = context.dataStore.data.first()
            preferences[TOKEN_KEY] ?: ""
        } catch (e: IOException) {
            Log.e("AuthInterceptor", "Error reading preferences", e)
            ""
        }
    }

    suspend fun clearToken() {
        try {
            context.dataStore.edit { preferences ->
                preferences.remove(TOKEN_KEY)
                preferences.remove(USERNAME_KEY)
            }
        } catch (e: Exception) {
            Log.e("AuthInterceptor", "Failed to clear token", e)
        }
    }

    suspend fun getUsername(): String {
        return try {
            val preferences = context.dataStore.data.first()
            preferences[USERNAME_KEY] ?: ""
        } catch (e: IOException) {
            Log.e("AuthInterceptor", "Error reading username", e)
            ""
        }
    }

    fun getUsernameFlow(): Flow<String> {
        return context.dataStore.data
            .catch { emit(androidx.datastore.preferences.core.emptyPreferences()) }
            .map { preferences ->
                preferences[USERNAME_KEY] ?: ""
            }
    }

    fun getTokenFlow(): Flow<String> {
        return context.dataStore.data
            .catch { emit(androidx.datastore.preferences.core.emptyPreferences()) }
            .map { preferences ->
                preferences[TOKEN_KEY] ?: ""
            }
    }
}
