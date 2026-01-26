package com.storytrim.app.data.dto

import com.google.gson.annotations.SerializedName

/**
 * 书籍上传清单
 */
data class BookUploadManifest(
    @SerializedName("book_id") val bookId: Long,
    @SerializedName("book_name") val bookName: String,
    @SerializedName("total_chapters") val totalChapters: Int,
    @SerializedName("chapters") val chapters: List<BookUploadChapter>
)

/**
 * 上传章节清单项
 */
data class BookUploadChapter(
    @SerializedName("local_id") val localId: Long,
    @SerializedName("index") val index: Int,
    @SerializedName("title") val title: String,
    @SerializedName("chapter_md5") val chapterMd5: String,
    @SerializedName("size") val size: Int,
    @SerializedName("words_count") val wordsCount: Int,
    @SerializedName("offset") val offset: Int,
    @SerializedName("length") val length: Int
)

/**
 * 上传响应
 */
data class BookUploadResp(
    @SerializedName("book_id") val bookId: Long,
    @SerializedName("chapter_mappings") val chapterMappings: List<ChapterMapping> = emptyList()
)

data class ChapterMapping(
    @SerializedName("local_id") val localId: Long,
    @SerializedName("cloud_id") val cloudId: Long
)
