package com.storytrim.app.data.dto

/**
 * 阅读进度响应体
 * @param lastChapterId 上次章节ID（云端ID）
 * @param lastPromptId 上次精简模式ID
 * @param updatedAt 更新时间戳
 */
data class ReadingProgressResp(
    val lastChapterId: Long,
    val lastPromptId: Int,
    val updatedAt: Long
)
