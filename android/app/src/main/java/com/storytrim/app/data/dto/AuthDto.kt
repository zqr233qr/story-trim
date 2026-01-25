package com.storytrim.app.data.dto

import com.google.gson.annotations.SerializedName

data class LoginReq(
    @SerializedName("username") val username: String,
    @SerializedName("password") val password: String
)

data class LoginResp(
    @SerializedName("token") val token: String
)

data class RegisterReq(
    @SerializedName("username") val username: String,
    @SerializedName("password") val password: String
)