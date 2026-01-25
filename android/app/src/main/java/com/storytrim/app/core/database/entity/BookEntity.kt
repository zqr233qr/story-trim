package com.storytrim.app.core.database.entity

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "books")
data class BookEntity(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    @ColumnInfo(name = "cloud_id") val cloudId: Long = 0,
    @ColumnInfo(name = "user_id") val userId: Long = 0,
    @ColumnInfo(name = "book_md5") val bookMd5: String,
    val title: String,
    @ColumnInfo(name = "total_chapters") val totalChapters: Int,
    @ColumnInfo(name = "sync_state") val syncState: Int = 0,
    @ColumnInfo(name = "created_at") val createdAt: Long
)