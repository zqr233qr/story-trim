package com.storytrim.app.data.repository

import android.content.Context
import com.storytrim.app.core.database.AppDatabase
import com.storytrim.app.core.database.dao.BookDao
import com.storytrim.app.core.database.dao.ChapterDao
import com.storytrim.app.core.database.dao.ContentDao
import com.storytrim.app.core.database.dao.ReadingHistoryDao
import com.storytrim.app.core.database.entity.BookEntity
import com.storytrim.app.core.database.entity.ChapterEntity
import com.storytrim.app.core.database.entity.ContentEntity
import com.storytrim.app.core.database.entity.ReadingHistoryEntity
import com.storytrim.app.core.network.ApiClient
import com.storytrim.app.core.network.TrimService
import com.storytrim.app.core.parser.FileParser
import com.storytrim.app.core.utils.ZipUtils
import com.storytrim.app.data.dto.BatchChapterContentReq
import com.storytrim.app.data.dto.BatchTrimByMd5Request
import com.storytrim.app.data.dto.BatchTrimRequest
import com.storytrim.app.data.dto.BookUploadChapter
import com.storytrim.app.data.dto.BookUploadManifest
import com.storytrim.app.data.dto.BookUploadResp
import com.storytrim.app.data.dto.ChapterStatusReq
import com.storytrim.app.data.dto.ContentStatusReq
import com.storytrim.app.data.dto.ChapterTrimOption
import com.storytrim.app.data.dto.ChapterTrimStatusResp
import com.storytrim.app.data.dto.ChapterTrimTaskReq
import com.storytrim.app.data.dto.ChapterTrimTaskResp
import com.storytrim.app.data.dto.PointsBalanceResp
import com.storytrim.app.data.dto.ReadingProgressReq
import com.storytrim.app.data.dto.ReadingProgressResp
import com.storytrim.app.data.dto.TrimStatus
import com.storytrim.app.data.dto.TrimStatusSyncReqById
import com.storytrim.app.data.dto.TrimStatusSyncReqByMd5
import com.storytrim.app.data.dto.TrimStatusSyncResp
import com.storytrim.app.data.model.Book
import com.storytrim.app.data.model.Chapter
import com.storytrim.app.data.model.Prompt
import com.storytrim.app.feature.book.BookService
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import okhttp3.ResponseBody
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.MultipartBody
import okhttp3.RequestBody.Companion.asRequestBody
import com.google.gson.Gson
import java.io.File
import java.io.FileOutputStream
import java.io.InputStream
import java.util.Locale
import kotlin.collections.LinkedHashMap
import javax.inject.Inject
import javax.inject.Singleton
import java.util.zip.ZipOutputStream
import java.util.zip.ZipEntry

@Singleton
class BookRepository @Inject constructor(
    private val apiClient: ApiClient,
    private val bookService: BookService,
    private val trimService: TrimService,
    private val bookDao: BookDao,
    private val chapterDao: ChapterDao,
    private val contentDao: ContentDao,
    private val readingHistoryDao: ReadingHistoryDao,
    private val fileParser: FileParser,
    private val appDatabase: AppDatabase,
    private val zipUtils: ZipUtils,
    @ApplicationContext private val context: Context
) {
    private var cachedPrompts: List<Prompt> = emptyList()
    private val trimCache = mutableMapOf<String, String>()
    private val trimStatusCache = object : LinkedHashMap<String, List<Int>>(256, 0.75f, true) {
        override fun removeEldestEntry(eldest: MutableMap.MutableEntry<String, List<Int>>?): Boolean {
            return size > 256
        }
    }
    private val trimStatusLock = Any()
    private val contentMemoryCache = object : LinkedHashMap<String, String>(64, 0.75f, true) {
        override fun removeEldestEntry(eldest: MutableMap.MutableEntry<String, String>?): Boolean {
            return size > 64
        }
    }
    private val contentCacheLock = Any()
    private val chapterCacheDir: File by lazy { File(context.cacheDir, "chapter_cache") }
    private val gson = Gson()

    @Suppress("UNUSED_PARAMETER")
    fun getBooksStream(_userId: Long): Flow<List<Book>> {
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

        val book = bookDao.getBookById(bookId) ?: return null
        if (book.syncState == 0) {
            val md5 = chapter.md5
            if (md5.isBlank()) return null
            return fetchChapterTrimByMd5(bookId, chapter, promptId, md5)
        }

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

    private suspend fun fetchChapterTrimByMd5(bookId: Long, chapter: Chapter, promptId: Int, md5: String): String? {
        return try {
            val res = apiClient.safeCall { bookService.getBatchTrimmedByMd5(BatchTrimByMd5Request(listOf(md5), promptId)) }
            if (res.code == 0 && !res.data.isNullOrEmpty()) {
                val trimmed = res.data.firstOrNull()?.trimmedContent?.trim()
                if (!trimmed.isNullOrBlank()) {
                    cacheChapterTrim(bookId, chapter.id, promptId, trimmed)
                    trimmed
                } else null
            } else null
        } catch (e: Exception) {
            android.util.Log.e("BookRepository", "fetchChapterTrimByMd5 failed", e)
            null
        }
    }

    /**
     * 批量预加载精简内容（遵循批量接口上限：10）
     */
    suspend fun preloadChapterTrims(chapters: List<Chapter>, promptId: Int) {
        if (chapters.isEmpty() || promptId <= 0) return
        val book = bookDao.getBookById(chapters.first().bookId) ?: return
        if (book.syncState == 0 && book.cloudId <= 0) return

        val candidates = chapters.filter { chapter ->
            val cached = getCachedChapterTrim(chapter.bookId, chapter.id, promptId)
            cached.isNullOrBlank()
        }
        if (candidates.isEmpty()) return

        val eligible = mutableListOf<Chapter>()
        val unknown = mutableListOf<Chapter>()
        for (chapter in candidates) {
            val cachedStatus = getCachedTrimStatus(chapter)
            if (cachedStatus != null) {
                if (cachedStatus.contains(promptId)) {
                    eligible.add(chapter)
                }
            } else {
                unknown.add(chapter)
            }
        }

        if (unknown.isNotEmpty()) {
            if (book.syncState == 0) {
                val md5s = unknown.mapNotNull { it.md5.takeIf { md5 -> md5.isNotBlank() } }
                val map = syncTrimStatusByMd5(md5s)
                unknown.forEach { chapter ->
                    val promptIds = map[chapter.md5] ?: emptyList()
                    cacheTrimStatus(chapter, promptIds)
                    if (promptIds.contains(promptId)) {
                        eligible.add(chapter)
                    }
                }
            } else if (book.cloudId > 0) {
                val map = syncTrimStatusByBookId(book.cloudId)
                unknown.forEach { chapter ->
                    val promptIds = resolvePromptIdsFromMap(chapter, map) ?: emptyList()
                    cacheTrimStatus(chapter, promptIds)
                    if (promptIds.contains(promptId)) {
                        eligible.add(chapter)
                    }
                }
            }
        }

        if (eligible.isEmpty()) return

        if (book.syncState == 0) {
            val md5Map = eligible.associateBy { it.md5 }
            val md5s = md5Map.keys.filter { it.isNotBlank() }
            val batches = md5s.chunked(10)
            for (batch in batches) {
                try {
                    val res = apiClient.safeCall { bookService.getBatchTrimmedByMd5(BatchTrimByMd5Request(batch, promptId)) }
                    if (res.code == 0 && !res.data.isNullOrEmpty()) {
                        res.data.forEach { dto ->
                            val md5 = dto.chapterMd5 ?: return@forEach
                            val chapter = md5Map[md5] ?: return@forEach
                            val trimmed = dto.trimmedContent.trim()
                            if (trimmed.isNotBlank()) {
                                cacheChapterTrim(chapter.bookId, chapter.id, promptId, trimmed)
                            }
                        }
                    }
                } catch (e: Exception) {
                    android.util.Log.w("BookRepository", "preloadChapterTrimsByMd5 failed", e)
                }
            }
        } else {
            val idMap = eligible.associateBy { if (it.cloudId > 0) it.cloudId else it.id }
            val cloudIds = idMap.keys.toList()
            val batches = cloudIds.chunked(10)
            for (batch in batches) {
                try {
                    val res = apiClient.safeCall { bookService.getBatchTrimmedById(BatchTrimRequest(batch, promptId)) }
                    if (res.code == 0 && !res.data.isNullOrEmpty()) {
                        res.data.forEach { dto ->
                            val chapter = idMap[dto.chapterId] ?: return@forEach
                            val trimmed = dto.trimmedContent.trim()
                            if (trimmed.isNotBlank()) {
                                cacheChapterTrim(chapter.bookId, chapter.id, promptId, trimmed)
                            }
                        }
                    }
                } catch (e: Exception) {
                    android.util.Log.w("BookRepository", "preloadChapterTrimsById failed", e)
                }
            }
        }
    }

    private fun trimStatusCacheKey(chapter: Chapter): String {
        return if (chapter.md5.isNotBlank()) "md5:${chapter.md5}" else "id:${chapter.id}"
    }

    private fun getCachedTrimStatus(chapter: Chapter): List<Int>? {
        val key = trimStatusCacheKey(chapter)
        synchronized(trimStatusLock) {
            return trimStatusCache[key]
        }
    }

    private fun cacheTrimStatus(chapter: Chapter, promptIds: List<Int>) {
        val key = trimStatusCacheKey(chapter)
        synchronized(trimStatusLock) {
            trimStatusCache[key] = promptIds
        }
    }

    private fun cacheTrimStatusByKey(key: String, promptIds: List<Int>) {
        synchronized(trimStatusLock) {
            trimStatusCache[key] = promptIds
        }
    }

    private fun resolvePromptIdsFromMap(chapter: Chapter, map: Map<String, List<Int>>): List<Int>? {
        if (chapter.md5.isNotBlank()) {
            map[chapter.md5]?.let { return it }
        }
        val idKey = if (chapter.cloudId > 0) chapter.cloudId.toString() else chapter.id.toString()
        return map[idKey]
    }

    suspend fun getChapterTrimStatusById(chapter: Chapter, bookMd5: String?): List<Int> {
        getCachedTrimStatus(chapter)?.let { return it }

        val book = bookDao.getBookById(chapter.bookId)
        if (book != null && book.syncState == 0 && chapter.md5.isNotBlank()) {
            val result = syncTrimStatusByMd5(listOf(chapter.md5))
            val promptIds = result[chapter.md5] ?: emptyList()
            cacheTrimStatus(chapter, promptIds)
            return promptIds
        }

        if (book != null && book.cloudId > 0) {
            val result = syncTrimStatusByBookId(book.cloudId)
            val promptIds = resolvePromptIdsFromMap(chapter, result) ?: emptyList()
            cacheTrimStatus(chapter, promptIds)
            if (promptIds.isNotEmpty()) return promptIds
        }

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
            val promptIds = res.data?.promptIds ?: emptyList()
            cacheTrimStatus(chapter, promptIds)
            promptIds
        } catch (e: Exception) {
            android.util.Log.e("BookRepository", "getChapterTrimStatusById failed", e)
            emptyList()
        }
    }

    suspend fun getChapterTrimStatusByMd5(chapterMd5: String): List<Int> {
        if (chapterMd5.isBlank()) return emptyList()
        val cacheKey = "md5:$chapterMd5"
        synchronized(trimStatusLock) {
            trimStatusCache[cacheKey]?.let { return it }
        }

        val result = syncTrimStatusByMd5(listOf(chapterMd5))
        val promptIds = result[chapterMd5] ?: emptyList()
        cacheTrimStatusByKey(cacheKey, promptIds)
        if (promptIds.isNotEmpty()) return promptIds

        return try {
            val res = apiClient.safeCall { bookService.getContentStatusByMd5(ContentStatusReq(chapterMd5)) }
            val ids = res.data?.promptIds ?: emptyList()
            cacheTrimStatusByKey(cacheKey, ids)
            ids
        } catch (e: Exception) {
            android.util.Log.e("BookRepository", "getChapterTrimStatusByMd5 failed", e)
            emptyList()
        }
    }

    private suspend fun syncTrimStatusByMd5(md5s: List<String>): Map<String, List<Int>> {
        if (md5s.isEmpty()) return emptyMap()
        val batches = md5s.chunked(20)
        val result = mutableMapOf<String, List<Int>>()
        for (batch in batches) {
            try {
                val res = apiClient.safeCall { bookService.syncTrimmedStatusByMd5(TrimStatusSyncReqByMd5(batch)) }
                if (res.code == 0 && res.data != null) {
                    result.putAll(res.data.trimmedMap)
                    res.data.trimmedMap.forEach { (key, value) ->
                        cacheTrimStatusByKey("md5:$key", value)
                    }
                }
            } catch (e: Exception) {
                android.util.Log.w("BookRepository", "syncTrimStatusByMd5 failed", e)
            }
        }
        return result
    }

    private suspend fun syncTrimStatusByBookId(bookId: Long): Map<String, List<Int>> {
        if (bookId <= 0) return emptyMap()
        return try {
            val res = apiClient.safeCall { bookService.syncTrimmedStatusById(TrimStatusSyncReqById(bookId)) }
            if (res.code == 0 && res.data != null) {
                res.data.trimmedMap
            } else emptyMap()
        } catch (e: Exception) {
            android.util.Log.w("BookRepository", "syncTrimStatusByBookId failed", e)
            emptyMap()
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

    @Suppress("UNUSED_PARAMETER")
    suspend fun refreshBooks(_userId: Long): Result<Unit> {
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
                        val uniqueMd5 = dto.md5?.takeIf { it.isNotBlank() }
                            ?: "fake_md5_${bookId}_${dto.id}"
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
        val md5 = chapter.md5
        val memoryCached = getMemoryCachedContent(md5)
        if (!memoryCached.isNullOrEmpty()) {
            return memoryCached
        }

        val diskCached = readDiskCachedContent(md5)
        if (!diskCached.isNullOrEmpty()) {
            putMemoryCachedContent(md5, diskCached)
            return diskCached
        }

        val localContent = contentDao.getContentByMd5(chapter.md5)?.rawContent
        if (!localContent.isNullOrEmpty()) {
            cacheChapterContent(md5, localContent)
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
                    cacheChapterContent(md5, content)
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

    /**
     * 批量预加载章节内容（遵循批量接口上限：10）
     */
    suspend fun preloadChapterContents(chapters: List<Chapter>) {
        if (chapters.isEmpty()) return
        val book = bookDao.getBookById(chapters.first().bookId) ?: return

        val missing = mutableListOf<Chapter>()
        for (chapter in chapters) {
            val md5 = chapter.md5
            val memoryCached = getMemoryCachedContent(md5)
            if (!memoryCached.isNullOrEmpty()) continue

            val diskCached = readDiskCachedContent(md5)
            if (!diskCached.isNullOrEmpty()) {
                putMemoryCachedContent(md5, diskCached)
                continue
            }

            val localContent = contentDao.getContentByMd5(md5)?.rawContent
            if (!localContent.isNullOrEmpty()) {
                cacheChapterContent(md5, localContent)
                continue
            }

            if (book.syncState != 0 && book.cloudId > 0) {
                missing.add(chapter)
            }
        }

        if (missing.isEmpty()) return

        val idMap = missing.associateBy { if (it.cloudId > 0) it.cloudId else it.id }
        val cloudIds = idMap.keys.toList()
        val batches = cloudIds.chunked(10)
        for (batch in batches) {
            try {
                val res = apiClient.safeCall { bookService.getBatchChapterContents(BatchChapterContentReq(batch)) }
                if (res.code == 0 && !res.data.isNullOrEmpty()) {
                    val contents = res.data.mapNotNull { dto ->
                        val chapter = idMap[dto.chapterId]
                        val md5 = dto.chapterMd5.ifBlank { chapter?.md5 ?: "" }
                        if (md5.isBlank()) return@mapNotNull null
                        cacheChapterContent(md5, dto.content)
                        ContentEntity(chapterMd5 = md5, rawContent = dto.content)
                    }
                    if (contents.isNotEmpty()) {
                        contentDao.insertContents(contents)
                    }
                }
            } catch (e: Exception) {
                android.util.Log.w("BookRepository", "preloadChapterContents failed", e)
            }
        }
    }

    private fun cacheChapterContent(md5: String, content: String) {
        if (content.isBlank()) return
        putMemoryCachedContent(md5, content)
        writeDiskCachedContent(md5, content)
    }

    private fun getMemoryCachedContent(md5: String): String? {
        synchronized(contentCacheLock) {
            return contentMemoryCache[md5]
        }
    }

    private fun putMemoryCachedContent(md5: String, content: String) {
        synchronized(contentCacheLock) {
            contentMemoryCache[md5] = content
        }
    }

    private fun readDiskCachedContent(md5: String): String? {
        return try {
            val file = File(chapterCacheDir, "$md5.txt")
            if (!file.exists()) return null
            file.setLastModified(System.currentTimeMillis())
            file.readText(Charsets.UTF_8)
        } catch (e: Exception) {
            android.util.Log.w("BookRepository", "readDiskCachedContent failed", e)
            null
        }
    }

    private fun writeDiskCachedContent(md5: String, content: String) {
        try {
            if (!chapterCacheDir.exists()) {
                chapterCacheDir.mkdirs()
            }
            val file = File(chapterCacheDir, "$md5.txt")
            file.writeText(content, Charsets.UTF_8)
            pruneDiskCache()
        } catch (e: Exception) {
            android.util.Log.w("BookRepository", "writeDiskCachedContent failed", e)
        }
    }

    private fun pruneDiskCache() {
        val files = chapterCacheDir.listFiles() ?: return
        if (files.size <= 50) return
        val sorted = files.sortedBy { it.lastModified() }
        val toDelete = sorted.take(files.size - 50)
        toDelete.forEach { it.delete() }
    }

    @Suppress("UNUSED_PARAMETER")
    suspend fun downloadBookContent(bookId: Long, cloudBookId: Long, _userId: Long, onProgress: ((Int) -> Unit)?): Result<Unit> {
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

    /**
     * 同步本地书籍到云端（上传 zip 并写入映射）
     */
    suspend fun uploadBookZip(bookId: Long, onProgress: (Int) -> Unit): Result<BookUploadResp> {
        return try {
            val book = bookDao.getBookById(bookId) ?: return Result.failure(Exception("书籍不存在"))
            if (book.totalChapters <= 0) {
                return Result.failure(Exception("章节为空"))
            }

            val tempDir = File(context.cacheDir, "upload_books/${bookId}_${System.currentTimeMillis()}")
            if (!tempDir.exists()) tempDir.mkdirs()

            val manifest = buildUploadManifest(book, onProgress)
            val bookFile = File(tempDir, "book.txt")
            val manifestFile = File(tempDir, "manifest.json")
            bookFile.writeText(manifest.first, Charsets.UTF_8)
            manifestFile.writeText(gson.toJson(manifest.second), Charsets.UTF_8)

            val zipFile = File(context.cacheDir, "upload_books/${bookId}_${System.currentTimeMillis()}.zip")
            zipFile.parentFile?.let { parent ->
                if (!parent.exists()) parent.mkdirs()
            }
            zipDirectory(tempDir, zipFile)

            onProgress(85)

            val requestBody = zipFile.asRequestBody("application/zip".toMediaType())
            val part = MultipartBody.Part.createFormData("file", zipFile.name, requestBody)
            val response = apiClient.safeCall {
                bookService.uploadBookZip(
                    bookName = book.title,
                    bookMd5 = book.bookMd5,
                    totalChapters = book.totalChapters,
                    file = part
                )
            }

            if (response.code != 0 || response.data == null) {
                return Result.failure(Exception(response.msg))
            }

            applyUploadMappings(bookId, response.data)
            onProgress(100)
            Result.success(response.data)
        } catch (e: Exception) {
            Result.failure(e)
        } finally {
            runCatching {
                val cacheBase = File(context.cacheDir, "upload_books")
                if (cacheBase.exists()) {
                    cacheBase.listFiles()?.forEach { child ->
                        if (child.isDirectory) child.deleteRecursively() else child.delete()
                    }
                }
            }
        }
    }

    private suspend fun buildUploadManifest(
        book: BookEntity,
        onProgress: (Int) -> Unit
    ): Pair<String, BookUploadManifest> {
        val chapters = mutableListOf<BookUploadChapter>()
        val contentBuilder = StringBuilder()
        var offsetBytes = 0

        val total = book.totalChapters
        val batchSize = 200
        var offset = 0
        while (offset < total) {
            val batch = chapterDao.getChaptersByBookIdPaged(book.id, batchSize, offset)
            if (batch.isEmpty()) break

            val md5s = batch.map { it.md5 }
            val contents = contentDao.getContentsByMd5s(md5s).associateBy { it.chapterMd5 }

            batch.forEach { chapter ->
                val content = contents[chapter.md5]?.rawContent ?: ""
                val bytes = content.toByteArray(Charsets.UTF_8)
                val length = bytes.size
                chapters.add(
                    BookUploadChapter(
                        localId = chapter.id,
                        index = chapter.chapterIndex,
                        title = chapter.title,
                        chapterMd5 = chapter.md5,
                        size = length,
                        wordsCount = chapter.wordsCount,
                        offset = offsetBytes,
                        length = length
                    )
                )
                contentBuilder.append(content)
                offsetBytes += length
            }

            offset += batch.size
            val progress = ((offset.toDouble() / total) * 60).toInt().coerceIn(1, 60)
            onProgress(progress)
        }

        val manifest = BookUploadManifest(
            bookId = book.cloudId,
            bookName = book.title,
            totalChapters = total,
            chapters = chapters
        )
        return contentBuilder.toString() to manifest
    }

    private fun zipDirectory(sourceDir: File, targetZip: File) {
        ZipOutputStream(FileOutputStream(targetZip)).use { zipOut ->
            sourceDir.listFiles()?.forEach { file ->
                val entry = ZipEntry(file.name)
                zipOut.putNextEntry(entry)
                file.inputStream().use { it.copyTo(zipOut) }
                zipOut.closeEntry()
            }
        }
    }

    private fun applyUploadMappings(bookId: Long, resp: BookUploadResp) {
        if (resp.chapterMappings.isEmpty()) return
        val db = appDatabase.openHelper.writableDatabase
        db.beginTransaction()
        try {
            val batchSize = 200
            val mappings = resp.chapterMappings
            for (i in mappings.indices step batchSize) {
                val batch = mappings.subList(i, kotlin.math.min(i + batchSize, mappings.size))
                val whenParts = mutableListOf<String>()
                val params = mutableListOf<Long>()
                val ids = mutableListOf<Long>()
                batch.forEach { mapping ->
                    whenParts.add("WHEN ? THEN ?")
                    params.add(mapping.localId)
                    params.add(mapping.cloudId)
                    ids.add(mapping.localId)
                }
                val idPlaceholders = ids.joinToString(",") { "?" }
                val sql = "UPDATE chapters SET cloud_id = CASE id ${whenParts.joinToString(" ")} END WHERE id IN ($idPlaceholders)"
                db.execSQL(sql, (params + ids).toTypedArray())
            }
            db.execSQL("UPDATE books SET cloud_id = ?, sync_state = 1 WHERE id = ?", arrayOf(resp.bookId, bookId))
            db.setTransactionSuccessful()
        } finally {
            db.endTransaction()
        }
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

    /**
     * 更新阅读进度：本地入库并尝试同步到云端
     * @param bookId 本地书籍ID
     * @param chapterId 本地章节ID
     * @param promptId 精简模式ID（0表示原文）
     */
    suspend fun updateReadingProgress(bookId: Long, chapterId: Long, promptId: Int) {
        val timestamp = System.currentTimeMillis()
        readingHistoryDao.insertReadingHistory(
            ReadingHistoryEntity(
                bookId = bookId,
                lastChapterId = chapterId,
                lastPromptId = promptId,
                updatedAt = timestamp
            )
        )

        val book = bookDao.getBookById(bookId) ?: return
        if (book.cloudId <= 0) return

        val chapter = chapterDao.getChapterById(chapterId)
        val cloudChapterId = if (chapter != null && chapter.cloudId > 0) chapter.cloudId else chapterId

        try {
            apiClient.safeCall {
                bookService.updateReadingProgress(
                    book.cloudId,
                    ReadingProgressReq(chapterId = cloudChapterId, promptId = promptId)
                )
            }
        } catch (e: Exception) {
            android.util.Log.w("BookRepository", "updateReadingProgress failed", e)
        }
    }

    /**
     * 获取本地阅读进度
     */
    suspend fun getLocalReadingHistory(bookId: Long): ReadingHistoryEntity? {
        return readingHistoryDao.getReadingHistory(bookId)
    }

    /**
     * 获取云端阅读进度并映射为本地章节ID
     */
    suspend fun getCloudReadingHistory(book: Book): ReadingHistoryEntity? {
        if (book.cloudId <= 0) return null
        return try {
            val res = apiClient.safeCall { bookService.getReadingProgress(book.cloudId) }
            if (res.code != 0 || res.data == null) return null
            val data: ReadingProgressResp = res.data
            val localChapter = chapterDao.getChapterByCloudId(data.lastChapterId)
                ?: chapterDao.getChapterById(data.lastChapterId)
                ?: return null
            ReadingHistoryEntity(
                bookId = book.id,
                lastChapterId = localChapter.id,
                lastPromptId = data.lastPromptId,
                updatedAt = data.updatedAt
            )
        } catch (e: Exception) {
            android.util.Log.w("BookRepository", "getCloudReadingHistory failed", e)
            null
        }
    }
    
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
