package com.storytrim.app.core.database.entity

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "reading_history")
data class ReadingHistoryEntity(
    @PrimaryKey @ColumnInfo(name = "book_id") val bookId: Long,
    @ColumnInfo(name = "last_chapter_id") val lastChapterId: Long,
    @ColumnInfo(name = "last_prompt_id") val lastPromptId: Int,
    @ColumnInfo(name = "updated_at") val updatedAt: Long
)
