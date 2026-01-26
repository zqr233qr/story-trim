package com.storytrim.app.feature.book

import com.storytrim.app.data.dto.ApiResp
import com.storytrim.app.data.dto.BatchChapterContentReq
import com.storytrim.app.data.dto.BatchTrimByMd5Request
import com.storytrim.app.data.dto.BookDetailRespDto
import com.storytrim.app.data.dto.BookListRespDto
import com.storytrim.app.data.dto.ChapterContentResp
import com.storytrim.app.data.dto.ChapterTrimStatusResp
import com.storytrim.app.data.dto.BatchTrimRequest
import com.storytrim.app.data.dto.BookUploadResp
import com.storytrim.app.data.dto.TrimmedContentResp
import com.storytrim.app.data.dto.ChapterStatusReq
import com.storytrim.app.data.dto.ContentStatusReq
import com.storytrim.app.data.dto.PromptStatusResp
import com.storytrim.app.data.dto.ChapterTrimTaskReq
import com.storytrim.app.data.dto.ChapterTrimTaskResp
import com.storytrim.app.data.dto.PointsBalanceResp
import com.storytrim.app.data.dto.PointsLedgerResp
import com.storytrim.app.data.dto.TaskItemResp
import com.storytrim.app.data.dto.TaskProgressResp
import com.storytrim.app.data.dto.PromptRespDto
import com.storytrim.app.data.dto.ReadingProgressReq
import com.storytrim.app.data.dto.ReadingProgressResp
import com.storytrim.app.data.dto.TrimStatusSyncReqById
import com.storytrim.app.data.dto.TrimStatusSyncReqByMd5
import com.storytrim.app.data.dto.TrimStatusSyncResp
import okhttp3.ResponseBody
import okhttp3.MultipartBody
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.Path
import retrofit2.http.Query
import retrofit2.http.Streaming
import retrofit2.http.Part

interface BookService {
    @GET("books")
    suspend fun getBooks(): ApiResp<List<BookListRespDto>>

    @GET("common/prompts")
    suspend fun getPrompts(): ApiResp<List<PromptRespDto>>

    @GET("books/{id}")
    suspend fun getBookDetail(@Path("id") bookId: Long): ApiResp<BookDetailRespDto>

    @DELETE("books/{id}")
    suspend fun deleteBook(@Path("id") bookId: Long): ApiResp<Unit>

    @GET("books/{id}/content-db")
    @Streaming
    suspend fun downloadBookContent(@Path("id") bookId: Long): ResponseBody

    @Multipart
    @POST("books/upload-zip")
    suspend fun uploadBookZip(
        @Query("book_name") bookName: String,
        @Query("book_md5") bookMd5: String,
        @Query("total_chapters") totalChapters: Int,
        @Part file: MultipartBody.Part
    ): ApiResp<BookUploadResp>

    @POST("chapters/content")
    suspend fun getBatchChapterContents(@Body req: BatchChapterContentReq): ApiResp<List<ChapterContentResp>>

    @POST("chapters/trim")
    suspend fun getBatchTrimmedById(@Body req: BatchTrimRequest): ApiResp<List<TrimmedContentResp>>

    @POST("contents/trim")
    suspend fun getBatchTrimmedByMd5(@Body req: BatchTrimByMd5Request): ApiResp<List<TrimmedContentResp>>

    @POST("chapters/status")
    suspend fun getChapterStatusById(@Body req: ChapterStatusReq): ApiResp<PromptStatusResp>

    @POST("contents/status")
    suspend fun getContentStatusByMd5(@Body req: ContentStatusReq): ApiResp<PromptStatusResp>

    @POST("contents/sync-status")
    suspend fun syncTrimmedStatusByMd5(@Body req: TrimStatusSyncReqByMd5): ApiResp<TrimStatusSyncResp>

    @POST("chapters/sync-status")
    suspend fun syncTrimmedStatusById(@Body req: TrimStatusSyncReqById): ApiResp<TrimStatusSyncResp>

    @GET("chapters/trim-status")
    suspend fun getChapterTrimStatus(
        @Query("book_id") bookId: Long,
        @Query("prompt_id") promptId: Int
    ): ApiResp<ChapterTrimStatusResp>

    @POST("chapters/trim-task")
    suspend fun startChapterTrimTask(@Body req: ChapterTrimTaskReq): ApiResp<ChapterTrimTaskResp>

    @GET("users/me/points")
    suspend fun getPointsBalance(): ApiResp<PointsBalanceResp>

    @GET("users/me/points/ledger")
    suspend fun getPointsLedger(
        @Query("page") page: Int,
        @Query("size") size: Int
    ): ApiResp<PointsLedgerResp>

    @GET("tasks/active")
    suspend fun getActiveTasks(): ApiResp<List<TaskItemResp>>

    @GET("tasks/{id}/progress")
    suspend fun getTaskProgress(@Path("id") taskId: String): ApiResp<TaskProgressResp>

    /**
     * 获取阅读进度
     * @param bookId 云端书籍ID
     */
    @GET("books/{id}/progress")
    suspend fun getReadingProgress(
        @Path("id") bookId: Long
    ): ApiResp<ReadingProgressResp>

    /**
     * 上报阅读进度
     * @param bookId 云端书籍ID
     * @param req 进度请求体
     */
    @POST("books/{id}/progress")
    suspend fun updateReadingProgress(
        @Path("id") bookId: Long,
        @Body req: ReadingProgressReq
    ): ApiResp<Unit>
}
