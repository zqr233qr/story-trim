package com.storytrim.app.data.model

data class Book(
    val id: Long,
    val cloudId: Long,
    val userId: Long,
    val bookMd5: String?,
    val title: String,
    val totalChapters: Int,
    val createdAt: Long,
    val syncState: Int = 0 // 0: Local, 1: Synced, 2: CloudOnly
)

data class Chapter(
    val id: Long,
    val bookId: Long,
    val cloudId: Long,
    val index: Int,
    val title: String,
    val md5: String,
    val content: String? = null
)