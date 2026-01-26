package com.storytrim.app.data.dto

/**
 * 阅读进度上报请求体
 * @param chapterId 章节ID（云端ID）
 * @param promptId 精简模式ID（0表示原文）
 */
data class ReadingProgressReq(
    val chapterId: Long,
    val promptId: Int
)
