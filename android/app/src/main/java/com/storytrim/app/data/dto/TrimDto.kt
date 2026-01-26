package com.storytrim.app.data.dto

import com.google.gson.annotations.SerializedName

data class BatchTrimRequest(
    @SerializedName("ids") val ids: List<Long>,
    @SerializedName("prompt_id") val promptId: Int
)

data class BatchTrimByMd5Request(
    @SerializedName("md5s") val md5s: List<String>,
    @SerializedName("prompt_id") val promptId: Int
)

data class TrimmedContentResp(
    @SerializedName("chapter_id") val chapterId: Long? = null,
    @SerializedName("chapter_md5") val chapterMd5: String? = null,
    @SerializedName("prompt_id") val promptId: Int,
    @SerializedName("trimmed_content") val trimmedContent: String
)

data class ChapterStatusReq(
    @SerializedName("chapter_id") val chapterId: Long,
    @SerializedName("book_md5") val bookMd5: String? = null,
    @SerializedName("chapter_md5") val chapterMd5: String? = null
)

data class ContentStatusReq(
    @SerializedName("chapter_md5") val chapterMd5: String
)

data class PromptStatusResp(
    @SerializedName("prompt_ids") val promptIds: List<Int>? = null
)
