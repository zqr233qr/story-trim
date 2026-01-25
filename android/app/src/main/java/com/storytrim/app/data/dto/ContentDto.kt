package com.storytrim.app.data.dto

import com.google.gson.annotations.SerializedName

data class BatchChapterContentReq(
    @SerializedName("ids") val ids: List<Long>
)

data class ChapterContentResp(
    @SerializedName("chapter_id") val chapterId: Long,
    @SerializedName("chapter_md5") val chapterMd5: String,
    @SerializedName("content") val content: String
)
