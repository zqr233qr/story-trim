package com.storytrim.app.data.repository

import com.storytrim.app.core.network.ApiClient
import com.storytrim.app.core.network.AuthInterceptor
import com.storytrim.app.data.dto.ApiResp
import com.storytrim.app.data.dto.LoginReq
import com.storytrim.app.data.dto.LoginResp
import com.storytrim.app.data.dto.RegisterReq
import com.storytrim.app.data.model.User
import com.storytrim.app.data.remote.AuthService
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthRepository @Inject constructor(
    private val authService: AuthService,
    private val authInterceptor: AuthInterceptor,
    private val apiClient: ApiClient
) {
    suspend fun login(username: String, password: String): Result<User> {
        return try {
            val response: ApiResp<LoginResp> = apiClient.safeCall {
                authService.login(LoginReq(username, password))
            }

            if (response.code == 0 && response.data != null) {
                val token = response.data.token
                if (token.isNotEmpty()) {
                    authInterceptor.saveToken(token)
                    authInterceptor.saveUsername(username)
                    Result.success(User(0, username, ""))
                } else {
                    Result.failure(Exception("登录失败：未返回token"))
                }
            } else {
                Result.failure(Exception("登录失败：${response.msg}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun register(username: String, password: String): Result<Unit> {
        return try {
            val response: ApiResp<Unit> = apiClient.safeCall {
                authService.register(RegisterReq(username, password))
            }

            if (response.code == 0) {
                Result.success(Unit)
            } else {
                Result.failure(Exception("注册失败：${response.msg}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun logout() {
        authInterceptor.clearToken()
    }

    suspend fun getToken(): String {
        return authInterceptor.getToken()
    }

    fun isLoggedInFlow() = authInterceptor.getTokenFlow()

    fun usernameFlow() = authInterceptor.getUsernameFlow()
}
