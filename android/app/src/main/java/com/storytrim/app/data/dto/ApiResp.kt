package com.storytrim.app.data.dto

import com.google.gson.annotations.SerializedName

data class ApiResp<T>(
    @SerializedName("code") val code: Int,
    @SerializedName("msg") val msg: String,
    @SerializedName("data") val data: T? = null
) {
    companion object {
        fun <T> error(e: Exception): ApiResp<T> {
            return ApiResp(-1, e.message ?: "Unknown error", null)
        }
    }
}
