package com.storytrim.app.core.network

import android.util.Log
import com.google.gson.Gson
import com.google.gson.annotations.SerializedName
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TrimService @Inject constructor(
    private val okHttpClient: OkHttpClient,
    private val authInterceptor: AuthInterceptor
) {
    private val gson = Gson()
    private val BASE_WS_URL_BY_ID = "ws://110.41.38.214:8088/api/v1/trim/stream/by-id"
    private val BASE_WS_URL_BY_MD5 = "ws://110.41.38.214:8088/api/v1/trim/stream/by-md5"

    data class TrimPayload(
        @SerializedName("book_id") val bookId: Long,
        @SerializedName("chapter_id") val chapterId: Long,
        @SerializedName("prompt_id") val promptId: Int
    )

    data class TrimMd5Payload(
        @SerializedName("content") val content: String,
        @SerializedName("md5") val md5: String,
        @SerializedName("prompt_id") val promptId: Int,
        @SerializedName("book_md5") val bookMd5: String,
        @SerializedName("book_title") val bookTitle: String,
        @SerializedName("chapter_title") val chapterTitle: String,
        @SerializedName("chapter_index") val chapterIndex: Int
    )

    data class StreamResponse(
        @SerializedName("c") val content: String?,
        @SerializedName("error") val error: String?
    )

    suspend fun connect(
        bookId: Long,
        chapterId: Long,
        promptId: Int,
        onData: (String) -> Unit,
        onError: (String) -> Unit,
        onClosed: () -> Unit
    ): WebSocket {
        val token = authInterceptor.getToken()
        val url = "$BASE_WS_URL_BY_ID?token=$token"
        Log.d("TrimService", "WS connect by id: $url bookId=$bookId chapterId=$chapterId promptId=$promptId")
        
        val request = Request.Builder().url(url).build()
        
        val listener = createListener(
            onOpen = {
                val payload = TrimPayload(bookId, chapterId, promptId)
                Log.d("TrimService", "WS send by id: ${gson.toJson(payload)}")
                it.send(gson.toJson(payload))
            },
            onData = onData,
            onError = onError,
            onClosed = onClosed
        )

        return okHttpClient.newWebSocket(request, listener)
    }

    suspend fun connectByMd5(
        content: String,
        md5: String,
        promptId: Int,
        bookMd5: String,
        bookTitle: String,
        chapterTitle: String,
        chapterIndex: Int,
        onData: (String) -> Unit,
        onError: (String) -> Unit,
        onClosed: () -> Unit
    ): WebSocket {
        val token = authInterceptor.getToken()
        val url = "$BASE_WS_URL_BY_MD5?token=$token"
        Log.d("TrimService", "WS connect by md5: $url md5=$md5 promptId=$promptId")

        val request = Request.Builder().url(url).build()

        val listener = createListener(
            onOpen = {
                val payload = TrimMd5Payload(
                    content, md5, promptId, bookMd5, bookTitle, chapterTitle, chapterIndex
                )
                Log.d(
                    "TrimService",
                    "WS send by md5: md5=$md5 promptId=$promptId contentLength=${content.length}"
                )
                it.send(gson.toJson(payload))
            },
            onData = onData,
            onError = onError,
            onClosed = onClosed
        )

        return okHttpClient.newWebSocket(request, listener)
    }

    private fun createListener(
        onOpen: (WebSocket) -> Unit,
        onData: (String) -> Unit,
        onError: (String) -> Unit,
        onClosed: () -> Unit
    ): WebSocketListener {
        return object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                Log.d("TrimService", "WebSocket Opened")
                onOpen(webSocket)
            }

            override fun onMessage(webSocket: WebSocket, text: String) {
                try {
                    val resp = gson.fromJson(text, StreamResponse::class.java)
                    if (resp.error != null) {
                        onError(resp.error)
                        webSocket.close(1000, "Server Error")
                    } else if (resp.content != null) {
                        onData(resp.content)
                    }
                } catch (e: Exception) {
                    Log.e("TrimService", "Parse error", e)
                }
            }

            override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                Log.e("TrimService", "WebSocket Failure", t)
                onError(t.message ?: "Unknown error")
            }

            override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
                webSocket.close(1000, null)
                onClosed()
            }
            
            override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                onClosed()
            }
        }
    }
}
