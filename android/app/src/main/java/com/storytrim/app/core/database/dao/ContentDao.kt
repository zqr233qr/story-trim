package com.storytrim.app.core.database.dao

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import com.storytrim.app.core.database.entity.ContentEntity
import kotlinx.coroutines.flow.Flow

@Dao
interface ContentDao {
    @Query("SELECT * FROM contents WHERE chapter_md5 = :md5 LIMIT 1")
    suspend fun getContentByMd5(md5: String): ContentEntity?

    @Query("SELECT * FROM contents WHERE chapter_md5 IN (:md5s)")
    suspend fun getContentsByMd5s(md5s: List<String>): List<ContentEntity>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertContent(content: ContentEntity)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertContents(contents: List<ContentEntity>)

    @Query("DELETE FROM contents WHERE chapter_md5 = :md5")
    suspend fun deleteContentByMd5(md5: String)
}