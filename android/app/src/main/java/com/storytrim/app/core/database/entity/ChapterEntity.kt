package com.storytrim.app.core.database.entity

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "chapters")
data class ChapterEntity(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    @ColumnInfo(name = "book_id") val bookId: Long,
    @ColumnInfo(name = "cloud_id") val cloudId: Long = 0,
    @ColumnInfo(name = "chapter_index") val chapterIndex: Int,
    val title: String,
    val md5: String,
    @ColumnInfo(name = "words_count") val wordsCount: Int = 0
)