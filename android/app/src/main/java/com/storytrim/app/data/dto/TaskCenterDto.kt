package com.storytrim.app.data.dto

import com.google.gson.annotations.SerializedName

data class TaskItemResp(
    @SerializedName("id") val id: String,
    @SerializedName("book_id") val bookId: Long,
    @SerializedName("book_title") val bookTitle: String,
    @SerializedName("prompt_id") val promptId: Int,
    @SerializedName("prompt_name") val promptName: String,
    @SerializedName("status") val status: String,
    @SerializedName("progress") val progress: Int,
    @SerializedName("error") val error: String? = null,
    @SerializedName("created_at") val createdAt: String
)

data class TaskProgressResp(
    @SerializedName("status") val status: String,
    @SerializedName("progress") val progress: Int,
    @SerializedName("error") val error: String? = null
)
