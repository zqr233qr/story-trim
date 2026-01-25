package com.storytrim.app.core.database

import android.content.Context
import com.storytrim.app.core.database.dao.BookDao
import com.storytrim.app.core.database.dao.ChapterDao
import com.storytrim.app.core.database.dao.ContentDao
import com.storytrim.app.core.database.dao.ReadingHistoryDao
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object DatabaseModule {

    @Provides
    @Singleton
    fun provideAppDatabase(@ApplicationContext context: Context): AppDatabase {
        return AppDatabase.getDatabase(context)
    }

    @Provides
    fun provideBookDao(database: AppDatabase): BookDao {
        return database.bookDao()
    }

    @Provides
    fun provideChapterDao(database: AppDatabase): ChapterDao {
        return database.chapterDao()
    }

    @Provides
    fun provideContentDao(database: AppDatabase): ContentDao {
        return database.contentDao()
    }

    @Provides
    fun provideReadingHistoryDao(database: AppDatabase): ReadingHistoryDao {
        return database.readingHistoryDao()
    }
}
