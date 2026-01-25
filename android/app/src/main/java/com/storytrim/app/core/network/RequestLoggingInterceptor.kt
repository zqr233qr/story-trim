package com.storytrim.app.core.network

import android.util.Log
import okhttp3.Interceptor
import okhttp3.Response

class RequestLoggingInterceptor : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()
        val startNs = System.nanoTime()
        Log.d("ApiRequest", "${request.method} ${request.url}")

        val response = chain.proceed(request)
        val tookMs = (System.nanoTime() - startNs) / 1_000_000
        Log.d("ApiResponse", "${response.code} ${request.method} ${request.url} (${tookMs}ms)")
        return response
    }
}
