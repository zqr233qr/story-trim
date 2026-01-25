package com.storytrim.app.data.dto

import com.google.gson.annotations.SerializedName

data class BookListRespDto(
    @SerializedName("id") val id: Long,
    @SerializedName("user_id") val user_id: Long,
    @SerializedName("book_md5") val book_md5: String?,
    @SerializedName("title") val title: String,
    @SerializedName("total_chapters") val total_chapters: Int
)

data class BookDetailRespDto(
    @SerializedName("book") val book: BookDetailItemDto,
    @SerializedName("chapters") val chapters: List<ChapterDto>
)

data class BookDetailItemDto(
    @SerializedName("id") val id: Long,
    @SerializedName("user_id") val user_id: Long,
    @SerializedName("title") val title: String,
    @SerializedName("book_md5") val book_md5: String?,
    @SerializedName("total_chapters") val total_chapters: Int
)

data class ChapterDto(
    @SerializedName("id") val id: Long,
    @SerializedName("index") val index: Int,
    @SerializedName("title") val title: String,
    @SerializedName("md5") val md5: String?,
    @SerializedName("words_count") val words_count: Int
)
