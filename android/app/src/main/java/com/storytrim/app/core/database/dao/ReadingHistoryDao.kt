package com.storytrim.app.core.database.dao

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import com.storytrim.app.core.database.entity.ReadingHistoryEntity

@Dao
interface ReadingHistoryDao {
    @Query("SELECT * FROM reading_history WHERE book_id = :bookId")
    suspend fun getReadingHistory(bookId: Long): ReadingHistoryEntity?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertReadingHistory(history: ReadingHistoryEntity)
}
