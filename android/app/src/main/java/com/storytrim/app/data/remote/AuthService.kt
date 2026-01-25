package com.storytrim.app.data.remote

import com.storytrim.app.data.dto.ApiResp
import com.storytrim.app.data.dto.LoginReq
import com.storytrim.app.data.dto.LoginResp
import com.storytrim.app.data.dto.RegisterReq
import retrofit2.http.Body
import retrofit2.http.POST

interface AuthService {
    @POST("auth/login")
    suspend fun login(@Body req: LoginReq): ApiResp<LoginResp>

    @POST("auth/register")
    suspend fun register(@Body req: RegisterReq): ApiResp<Unit>
}
