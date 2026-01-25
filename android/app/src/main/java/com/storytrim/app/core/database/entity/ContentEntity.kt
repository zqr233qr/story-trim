package com.storytrim.app.core.database.entity

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "contents")
data class ContentEntity(
    @PrimaryKey
    @ColumnInfo(name = "chapter_md5") val chapterMd5: String,
    @ColumnInfo(name = "raw_content") val rawContent: String
)