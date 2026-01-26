package com.storytrim.app.data.repository

import com.storytrim.app.core.network.ApiClient
import com.storytrim.app.data.dto.TaskItemResp
import com.storytrim.app.data.dto.TaskProgressResp
import com.storytrim.app.feature.book.BookService
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TaskRepository @Inject constructor(
    private val bookService: BookService,
    private val apiClient: ApiClient
) {
    suspend fun getActiveTasks(): Result<List<TaskItemResp>> {
        return try {
            val res = apiClient.safeCall { bookService.getActiveTasks() }
            if (res.code == 0 && res.data != null) {
                Result.success(res.data)
            } else {
                Result.failure(Exception(res.msg))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getTaskProgress(taskId: String): Result<TaskProgressResp> {
        return try {
            val res = apiClient.safeCall { bookService.getTaskProgress(taskId) }
            if (res.code == 0 && res.data != null) {
                Result.success(res.data)
            } else {
                Result.failure(Exception(res.msg))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
