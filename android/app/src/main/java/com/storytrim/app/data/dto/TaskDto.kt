package com.storytrim.app.data.dto

import com.google.gson.annotations.SerializedName

data class ChapterTrimStatusResp(
    @SerializedName("trimmed_chapter_ids")
    val trimmedChapterIds: List<Long>? = null,
    @SerializedName("processing_chapter_ids")
    val processingChapterIds: List<Long>? = null
)

data class ChapterTrimTaskReq(
    @SerializedName("book_id")
    val bookId: Long,
    @SerializedName("prompt_id")
    val promptId: Int,
    @SerializedName("chapter_ids")
    val chapterIds: List<Long>
)

data class ChapterTrimTaskResp(
    @SerializedName("task_id")
    val taskId: String
)

data class PointsBalanceResp(
    @SerializedName("balance")
    val balance: Int
)
