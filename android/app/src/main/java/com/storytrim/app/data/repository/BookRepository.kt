package com.storytrim.app.data.repository

import android.content.Context
import com.storytrim.app.core.database.AppDatabase
import com.storytrim.app.core.database.dao.BookDao
import com.storytrim.app.core.database.dao.ChapterDao
import com.storytrim.app.core.database.dao.ContentDao
import com.storytrim.app.core.database.entity.BookEntity
import com.storytrim.app.core.database.entity.ChapterEntity
import com.storytrim.app.core.database.entity.ContentEntity
import com.storytrim.app.core.network.ApiClient
import com.storytrim.app.core.network.TrimService
import com.storytrim.app.core.parser.FileParser
import com.storytrim.app.core.utils.ZipUtils
import com.storytrim.app.data.dto.BatchChapterContentReq
import com.storytrim.app.data.dto.BatchTrimRequest
import com.storytrim.app.data.dto.ChapterStatusReq
import com.storytrim.app.data.dto.ContentStatusReq
import com.storytrim.app.data.dto.ChapterTrimOption
import com.storytrim.app.data.dto.ChapterTrimStatusResp
import com.storytrim.app.data.dto.ChapterTrimTaskReq
import com.storytrim.app.data.dto.ChapterTrimTaskResp
import com.storytrim.app.data.dto.PointsBalanceResp
import com.storytrim.app.data.dto.TrimStatus
import com.storytrim.app.data.model.Book
import com.storytrim.app.data.model.Chapter
import com.storytrim.app.data.model.Prompt
import com.storytrim.app.feature.book.BookService
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import okhttp3.ResponseBody
import java.io.File
import java.io.FileOutputStream
import java.io.InputStream
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class BookRepository @Inject constructor(
    private val apiClient: ApiClient,
    private val bookService: BookService,
    private val trimService: TrimService,
    private val bookDao: BookDao,
    private val chapterDao: ChapterDao,
    private val contentDao: ContentDao,
    private val fileParser: FileParser,
    private val appDatabase: AppDatabase,
    private val zipUtils: ZipUtils,
    @ApplicationContext private val context: Context
) {
    private var cachedPrompts: List<Prompt> = emptyList()
    private val trimCache = mutableMapOf<String, String>()

    fun getBooksStream(userId: Long): Flow<List<Book>> {
        return bookDao.getAllBooks().map { entities ->
            entities.map { entity -> mapBookToDomain(entity) }
        }
    }

    suspend fun getPrompts(): List<Prompt> {
        if (cachedPrompts.isNotEmpty()) {
            android.util.Log.d("BookRepository", "getPrompts: returning cached ${cachedPrompts.size}")
            return cachedPrompts
        }
        android.util.Log.d("BookRepository", "getPrompts: fetching from server")
        return try {
            val res = apiClient.safeCall { bookService.getPrompts() }
            android.util.Log.d("BookRepository", "getPrompts: response code=${res.code}, data=${res.data}")
            if (res.code == 0 && res.data != null) {
                cachedPrompts = res.data.map {
                    Prompt(it.id, it.name, it.description, it.is_default)
                }
                android.util.Log.d("BookRepository", "getPrompts: loaded ${cachedPrompts.size} prompts")
                cachedPrompts
            } else {
                android.util.Log.e("BookRepository", "getPrompts: failed with code ${res.code}")
                emptyList()
            }
        } catch (e: Exception) {
            android.util.Log.e("BookRepository", "getPrompts: exception", e)
            emptyList()
        }
    }

    suspend fun startTrimStream(
        bookId: Long,
        chapterId: Long,
        promptId: Int,
        onData: (String) -> Unit,
        onError: (String) -> Unit,
        onClosed: () -> Unit
    ) {
        trimService.connect(bookId, chapterId, promptId, onData, onError, onClosed)
    }

    suspend fun startTrimStreamByMd5(
        content: String,
        md5: String,
        promptId: Int,
        bookMd5: String,
        bookTitle: String,
        chapterTitle: String,
        chapterIndex: Int,
        onData: (String) -> Unit,
        onError: (String) -> Unit,
        onClosed: () -> Unit
    ) {
        trimService.connectByMd5(
            content, md5, promptId, bookMd5, bookTitle, chapterTitle, chapterIndex,
            onData, onError, onClosed
        )
    }

    private fun trimCacheKey(bookId: Long, chapterId: Long, promptId: Int): String {
        return "$bookId:$chapterId:$promptId"
    }

    fun cacheChapterTrim(bookId: Long, chapterId: Long, promptId: Int, content: String) {
        trimCache[trimCacheKey(bookId, chapterId, promptId)] = content
    }

    fun getCachedChapterTrim(bookId: Long, chapterId: Long, promptId: Int): String? {
        return trimCache[trimCacheKey(bookId, chapterId, promptId)]
    }

    suspend fun fetchChapterTrim(bookId: Long, chapter: Chapter, promptId: Int): String? {
        val cached = getCachedChapterTrim(bookId, chapter.id, promptId)
        if (!cached.isNullOrBlank()) return cached

        val cloudChapterId = if (chapter.cloudId > 0) chapter.cloudId else chapter.id
        return try {
            val res = apiClient.safeCall { bookService.getBatchTrimmedById(BatchTrimRequest(listOf(cloudChapterId), promptId)) }
            if (res.code == 0 && !res.data.isNullOrEmpty()) {
                val trimmed = res.data.firstOrNull()?.trimmedContent?.trim()
                if (!trimmed.isNullOrBlank()) {
                    cacheChapterTrim(bookId, chapter.id, promptId, trimmed)
                    trimmed
                } else null
            } else null
        } catch (e: Exception) {
            android.util.Log.e("BookRepository", "fetchChapterTrim failed", e)
            null
        }
    }

    suspend fun getChapterTrimStatusById(chapter: Chapter, bookMd5: String?): List<Int> {
        return try {
            val res = apiClient.safeCall {
                bookService.getChapterStatusById(
                    ChapterStatusReq(
                        chapterId = if (chapter.cloudId > 0) chapter.cloudId else chapter.id,
                        bookMd5 = bookMd5,
                        chapterMd5 = chapter.md5
                    )
                )
            }
            res.data?.promptIds ?: emptyList()
        } catch (e: Exception) {
            android.util.Log.e("BookRepository", "getChapterTrimStatusById failed", e)
            emptyList()
        }
    }

    suspend fun getChapterTrimStatusByMd5(chapterMd5: String): List<Int> {
        return try {
            val res = apiClient.safeCall { bookService.getContentStatusByMd5(ContentStatusReq(chapterMd5)) }
            res.data?.promptIds ?: emptyList()
        } catch (e: Exception) {
            android.util.Log.e("BookRepository", "getChapterTrimStatusByMd5 failed", e)
            emptyList()
        }
    }

    suspend fun importBook(inputStream: InputStream, fileName: String): Result<Long> {
        return try {
            val result = fileParser.parse(inputStream, fileName)
            val bookEntity = BookEntity(
                bookMd5 = result.bookMd5,
                title = result.title,
                totalChapters = result.chapters.size,
                syncState = 0,
                createdAt = System.currentTimeMillis()
            )
            val bookId = bookDao.insertBook(bookEntity)
            val chapters = result.chapters.map { 
                ChapterEntity(
                    bookId = bookId,
                    chapterIndex = it.index,
                    title = it.title,
                    md5 = it.md5,
                    wordsCount = it.wordsCount
                )
            }
            val contents = result.chapters.map {
                ContentEntity(
                    chapterMd5 = it.md5,
                    rawContent = it.content
                )
            }
            chapterDao.insertChapters(chapters)
            contentDao.insertContents(contents)
            Result.success(bookId)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun refreshBooks(userId: Long): Result<Unit> {
        return try {
            val response = apiClient.safeCall { bookService.getBooks() }
            if (response.code == 0 && response.data != null) {
                response.data.forEach { dto ->
                    val existing = bookDao.getBookByCloudId(dto.id)
                    val newSyncState = if (existing != null && existing.syncState != 2) 1 else 2
                    
                    val entity = BookEntity(
                        id = existing?.id ?: 0,
                        cloudId = dto.id,
                        userId = dto.user_id,
                        bookMd5 = dto.book_md5 ?: "",
                        title = dto.title,
                        totalChapters = dto.total_chapters,
                        syncState = newSyncState,
                        createdAt = existing?.createdAt ?: System.currentTimeMillis()
                    )
                    if (existing != null) bookDao.updateBook(entity) else bookDao.insertBook(entity)
                }
                Result.success(Unit)
            } else Result.failure(Exception(response.msg))
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getBookDetail(bookId: Long): Result<Book> {
        val local = bookDao.getBookById(bookId)
        return local?.let { Result.success(mapBookToDomain(it)) } ?: Result.failure(Exception("Not found"))
    }

    suspend fun deleteBook(bookId: Long): Result<Unit> {
        return try {
            bookDao.getBookById(bookId)?.let { local ->
                if (local.syncState == 1 && local.cloudId > 0) {
                    apiClient.safeCall { bookService.deleteBook(local.cloudId) }
                }
                chapterDao.deleteChaptersByBookId(bookId)
                bookDao.deleteBookById(bookId)
            }
            Result.success(Unit)
        } catch (e: Exception) { Result.failure(e) }
    }

    suspend fun getChapters(bookId: Long): List<Chapter> {
        val localChapters = chapterDao.getChaptersByBookId(bookId).first()
        if (localChapters.isNotEmpty()) {
            return localChapters.map { entity -> mapChapterToDomain(entity) }
        }

        val book = bookDao.getBookById(bookId)
        if (book != null && book.syncState != 0 && book.cloudId > 0) {
            try {
                android.util.Log.d("BookRepository", "Fetching chapters from cloud via getBookDetail: ${book.cloudId}")
                val res = apiClient.safeCall { bookService.getBookDetail(book.cloudId) }
                android.util.Log.d("BookRepository", "Fetch detail result: code=${res.code}, chapters=${res.data?.chapters?.size}")
                
                if (res.code == 0 && res.data != null) {
                    val entities = res.data.chapters.map { dto ->
                        val uniqueMd5 = if (dto.md5.isNullOrEmpty()) "fake_md5_${bookId}_${dto.id}" else dto.md5!!
                        ChapterEntity(
                            bookId = bookId,
                            cloudId = dto.id,
                            chapterIndex = dto.index,
                            title = dto.title,
                            md5 = uniqueMd5,
                            wordsCount = dto.words_count
                        )
                    }
                    chapterDao.insertChapters(entities)
                    return entities.map { entity -> mapChapterToDomain(entity) }
                } else {
                    android.util.Log.w("BookRepository", "Failed to fetch book detail: ${res.msg}")
                }
            } catch (e: Exception) {
                android.util.Log.e("BookRepository", "Error fetching chapters", e)
                e.printStackTrace()
            }
        } else {
            android.util.Log.w("BookRepository", "Skipping cloud fetch: book=${book?.title}, syncState=${book?.syncState}, cloudId=${book?.cloudId}")
        }
        return emptyList()
    }

    suspend fun getChapterContent(chapterId: Long): String {
        val chapter = chapterDao.getChapterById(chapterId) ?: return ""
        val localContent = contentDao.getContentByMd5(chapter.md5)?.rawContent
        if (!localContent.isNullOrEmpty()) {
            return localContent
        }

        val book = bookDao.getBookById(chapter.bookId)
        if (book != null && book.syncState != 0 && book.cloudId > 0) {
            val cloudChapterId = if (chapter.cloudId > 0) chapter.cloudId else chapter.id
            try {
                android.util.Log.d("BookRepository", "Fetching content from cloud for chapter: $cloudChapterId")
                val res = apiClient.safeCall { 
                    bookService.getBatchChapterContents(BatchChapterContentReq(listOf(cloudChapterId))) 
                }
                
                if (res.code == 0 && res.data != null && res.data.isNotEmpty()) {
                    val content = res.data[0].content
                    contentDao.insertContents(listOf(ContentEntity(chapterMd5 = chapter.md5, rawContent = content)))
                    return content
                } else {
                    android.util.Log.w("BookRepository", "Failed to fetch content: ${res.msg}")
                }
            } catch (e: Exception) {
                android.util.Log.e("BookRepository", "Error fetching content", e)
                e.printStackTrace()
            }
        }
        return ""
    }

    suspend fun downloadBookContent(bookId: Long, cloudBookId: Long, userId: Long, onProgress: ((Int) -> Unit)?): Result<Unit> {
        return try {
            onProgress?.invoke(10)
            val body = apiClient.safeCallRaw { bookService.downloadBookContent(cloudBookId) } ?: return Result.failure(Exception("Download failed"))
            val zipFile = File(context.cacheDir, "book_$cloudBookId.zip")
            body.byteStream().use { input -> FileOutputStream(zipFile).use { input.copyTo(it) } }
            onProgress?.invoke(40)
            val unzipDir = File(context.cacheDir, "book_${cloudBookId}_unzipped").apply { if (exists()) deleteRecursively() }
            zipUtils.unzip(zipFile, unzipDir)
            onProgress?.invoke(60)
            val dbFile = File(unzipDir, "book.db")
            if (!dbFile.exists()) return Result.failure(Exception("No book.db"))
            mergeDatabase(bookId, cloudBookId, dbFile.absolutePath)
            onProgress?.invoke(100)
            Result.success(Unit)
        } catch (e: Exception) { Result.failure(e) }
    }

    private fun mergeDatabase(bookId: Long, cloudBookId: Long, dbPath: String) {
        val db = appDatabase.openHelper.writableDatabase
        db.beginTransaction()
        try {
            db.execSQL("DELETE FROM chapters WHERE book_id = ?", arrayOf(bookId))
            db.execSQL("ATTACH DATABASE '$dbPath' AS downloaded")
            db.execSQL("INSERT INTO chapters (book_id, cloud_id, chapter_index, title, md5, words_count) SELECT $bookId, cloud_id, chapter_index, title, md5, words_count FROM downloaded.chapters")
            db.execSQL("INSERT OR IGNORE INTO contents (chapter_md5, raw_content) SELECT chapter_md5, raw_content FROM downloaded.contents")
            db.execSQL("UPDATE books SET cloud_id = $cloudBookId, sync_state = 1, total_chapters = (SELECT COUNT(*) FROM downloaded.chapters) WHERE id = $bookId")
            db.setTransactionSuccessful()
        } finally {
            db.endTransaction()
            try { db.execSQL("DETACH DATABASE downloaded") } catch (e: Exception) {}
        }
    }

    private suspend fun <T> ApiClient.safeCallRaw(call: suspend () -> T): T? = try { call() } catch (e: Exception) { null }
    
    private fun mapBookToDomain(entity: BookEntity) = Book(
        id = entity.id,
        cloudId = entity.cloudId,
        userId = entity.userId,
        bookMd5 = entity.bookMd5,
        title = entity.title,
        totalChapters = entity.totalChapters,
        createdAt = entity.createdAt,
        syncState = entity.syncState
    )

    suspend fun getChapterTrimStatus(
        bookId: Long,
        promptId: Int,
        chapters: List<Chapter>
    ): Result<List<ChapterTrimOption>> {
        android.util.Log.d("BookRepository", "getChapterTrimStatus: bookId=$bookId, promptId=$promptId, chapters=${chapters.size}")
        return try {
            val cloudBookId = bookDao.getBookById(bookId)?.cloudId ?: return Result.failure(Exception("Book not found"))
            val res = apiClient.safeCall { bookService.getChapterTrimStatus(cloudBookId, promptId) }
            android.util.Log.d("BookRepository", "getChapterTrimStatus: response code=${res.code}, data=${res.data}")
            if (res.code == 0 && res.data != null) {
                val trimmedList = res.data.trimmedChapterIds ?: emptyList()
                val processingList = res.data.processingChapterIds ?: emptyList()
                val trimmedSet = trimmedList.toSet()
                val processingSet = processingList.toSet()
                android.util.Log.d("BookRepository", "getChapterTrimStatus: trimmed=${trimmedSet.size}, processing=${processingSet.size}")
                val options = chapters.map { chapter ->
                    val cloudId = if (chapter.cloudId > 0) chapter.cloudId else chapter.id
                    val status = when {
                        processingSet.contains(cloudId) -> TrimStatus.PROCESSING
                        trimmedSet.contains(cloudId) -> TrimStatus.TRIMMED
                        else -> TrimStatus.AVAILABLE
                    }
                    ChapterTrimOption(
                        id = chapter.id,
                        index = chapter.index,
                        title = chapter.title,
                        status = status
                    )
                }
                Result.success(options)
            } else {
                Result.failure(Exception(res.msg))
            }
        } catch (e: Exception) {
            android.util.Log.e("BookRepository", "getChapterTrimStatus: exception", e)
            Result.failure(e)
        }
    }

    suspend fun startChapterTrimTask(
        bookId: Long,
        promptId: Int,
        chapterIds: List<Long>
    ): Result<String> {
        android.util.Log.d("BookRepository", "startChapterTrimTask: bookId=$bookId, promptId=$promptId, chapterIds=${chapterIds.size}")
        return try {
            val cloudBookId = bookDao.getBookById(bookId)?.cloudId ?: return Result.failure(Exception("Book not found"))
            val cloudChapterIds = chapterIds.map { id ->
                val chapter = chapterDao.getChapterById(id)
                if (chapter?.cloudId ?: 0 > 0) chapter!!.cloudId else id
            }
            val req = ChapterTrimTaskReq(cloudBookId, promptId, cloudChapterIds)
            val res = apiClient.safeCall { bookService.startChapterTrimTask(req) }
            android.util.Log.d("BookRepository", "startChapterTrimTask: response code=${res.code}")
            if (res.code == 0 && res.data != null) {
                Result.success(res.data.taskId)
            } else {
                Result.failure(Exception(res.msg))
            }
        } catch (e: Exception) {
            android.util.Log.e("BookRepository", "startChapterTrimTask: exception", e)
            Result.failure(e)
        }
    }

    suspend fun getPointsBalance(): Result<Int> {
        android.util.Log.d("BookRepository", "getPointsBalance: fetching")
        return try {
            val res = apiClient.safeCall { bookService.getPointsBalance() }
            android.util.Log.d("BookRepository", "getPointsBalance: response code=${res.code}, balance=${res.data?.balance}")
            if (res.code == 0 && res.data != null) {
                Result.success(res.data.balance)
            } else {
                Result.failure(Exception(res.msg))
            }
        } catch (e: Exception) {
            android.util.Log.e("BookRepository", "getPointsBalance: exception", e)
            Result.failure(e)
        }
    }

    private fun mapChapterToDomain(entity: ChapterEntity) = Chapter(
        id = entity.id,
        bookId = entity.bookId,
        cloudId = entity.cloudId,
        index = entity.chapterIndex,
        title = entity.title,
        md5 = entity.md5,
        content = ""
    )
}
