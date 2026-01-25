package com.storytrim.app.core.database.dao

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Update
import com.storytrim.app.core.database.entity.BookEntity
import kotlinx.coroutines.flow.Flow

@Dao
interface BookDao {
    @Query("SELECT * FROM books ORDER BY created_at DESC")
    fun getAllBooks(): Flow<List<BookEntity>>

    @Query("SELECT * FROM books WHERE id = :id LIMIT 1")
    suspend fun getBookById(id: Long): BookEntity?

    @Query("SELECT * FROM books WHERE user_id = :userId ORDER BY created_at DESC")
    fun getBooksByUserId(userId: Long): Flow<List<BookEntity>>

    @Query("SELECT * FROM books WHERE cloud_id = :cloudId LIMIT 1")
    suspend fun getBookByCloudId(cloudId: Long): BookEntity?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertBook(book: BookEntity): Long

    @Update
    suspend fun updateBook(book: BookEntity)

    @Query("DELETE FROM books WHERE id = :id")
    suspend fun deleteBookById(id: Long)

    @Query("DELETE FROM books WHERE cloud_id = :cloudId")
    suspend fun deleteBookByCloudId(cloudId: Long)
}
