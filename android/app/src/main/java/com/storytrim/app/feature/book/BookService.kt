package com.storytrim.app.feature.book

import com.storytrim.app.data.dto.ApiResp
import com.storytrim.app.data.dto.BatchChapterContentReq
import com.storytrim.app.data.dto.BookDetailRespDto
import com.storytrim.app.data.dto.BookListRespDto
import com.storytrim.app.data.dto.ChapterContentResp
import com.storytrim.app.data.dto.ChapterTrimStatusResp
import com.storytrim.app.data.dto.BatchTrimRequest
import com.storytrim.app.data.dto.TrimmedContentResp
import com.storytrim.app.data.dto.ChapterStatusReq
import com.storytrim.app.data.dto.ContentStatusReq
import com.storytrim.app.data.dto.PromptStatusResp
import com.storytrim.app.data.dto.ChapterTrimTaskReq
import com.storytrim.app.data.dto.ChapterTrimTaskResp
import com.storytrim.app.data.dto.PointsBalanceResp
import com.storytrim.app.data.dto.PromptRespDto
import okhttp3.ResponseBody
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Path
import retrofit2.http.Query
import retrofit2.http.Streaming

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

    @POST("chapters/content")
    suspend fun getBatchChapterContents(@Body req: BatchChapterContentReq): ApiResp<List<ChapterContentResp>>

    @POST("chapters/trim")
    suspend fun getBatchTrimmedById(@Body req: BatchTrimRequest): ApiResp<List<TrimmedContentResp>>

    @POST("chapters/status")
    suspend fun getChapterStatusById(@Body req: ChapterStatusReq): ApiResp<PromptStatusResp>

    @POST("contents/status")
    suspend fun getContentStatusByMd5(@Body req: ContentStatusReq): ApiResp<PromptStatusResp>

    @GET("chapters/trim-status")
    suspend fun getChapterTrimStatus(
        @Query("book_id") bookId: Long,
        @Query("prompt_id") promptId: Int
    ): ApiResp<ChapterTrimStatusResp>

    @POST("chapters/trim-task")
    suspend fun startChapterTrimTask(@Body req: ChapterTrimTaskReq): ApiResp<ChapterTrimTaskResp>

    @GET("users/me/points")
    suspend fun getPointsBalance(): ApiResp<PointsBalanceResp>
}
