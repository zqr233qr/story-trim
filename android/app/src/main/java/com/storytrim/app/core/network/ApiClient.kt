package com.storytrim.app.core.network

import com.storytrim.app.data.dto.ApiResp
import javax.inject.Inject

suspend fun <T> safeCall(call: suspend () -> ApiResp<T>): ApiResp<T> {
    return try {
        call()
    } catch (e: Exception) {
        ApiResp.error(e)
    }
}

class ApiClient @Inject constructor() {
    suspend fun <T> safeCall(call: suspend () -> ApiResp<T>): ApiResp<T> {
        return try {
            call()
        } catch (e: Exception) {
            ApiResp.error(e)
        }
    }
}
