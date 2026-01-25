package com.storytrim.app.core.network

import com.storytrim.app.data.dto.ApiResp
import com.google.gson.Gson
import com.google.gson.JsonSyntaxException
import java.io.IOException
import java.net.HttpURLConnection
import retrofit2.HttpException
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ApiErrorHandler @Inject constructor(private val gson: Gson) {
    fun handle(throwable: Throwable): ApiResp<*> {
        return when (throwable) {
            is HttpException -> {
                val code = throwable.code()
                val errorBody = throwable.response()?.errorBody()?.string()
                try {
                    if (errorBody != null) {
                        val apiResp = gson.fromJson(errorBody, ApiResp::class.java)
                        apiResp ?: ApiResp(code, "Network error: $code", null)
                    } else {
                        ApiResp(code, "Network error: $code", null)
                    }
                } catch (e: JsonSyntaxException) {
                    ApiResp(code, "Network error: $code", null)
                }
            }
            is IOException -> ApiResp(-1, "Network error: ${throwable.message}", null)
            else -> ApiResp(-1, "Unknown error: ${throwable.message}", null)
        }
    }
}