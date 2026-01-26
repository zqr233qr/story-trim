package com.storytrim.app.data.dto

import com.google.gson.annotations.SerializedName

/**
 * 按 MD5 批量同步精简状态请求
 */
data class TrimStatusSyncReqByMd5(
    @SerializedName("md5s") val md5s: List<String>
)

/**
 * 按章节ID批量同步精简状态请求
 */
data class TrimStatusSyncReqById(
    @SerializedName("book_id") val bookId: Long
)

/**
 * 精简状态响应
 */
data class TrimStatusSyncResp(
    @SerializedName("trimmed_map") val trimmedMap: Map<String, List<Int>> = emptyMap()
)
