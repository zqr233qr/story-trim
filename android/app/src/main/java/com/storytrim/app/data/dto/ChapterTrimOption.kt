package com.storytrim.app.data.dto

import com.google.gson.annotations.SerializedName

data class ChapterTrimOption(
    val id: Long,
    val index: Int,
    val title: String,
    val status: TrimStatus
)

enum class TrimStatus {
    AVAILABLE,
    TRIMMED,
    PROCESSING
}
