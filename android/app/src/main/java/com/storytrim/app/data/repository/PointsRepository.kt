package com.storytrim.app.data.repository

import com.storytrim.app.core.network.ApiClient
import com.storytrim.app.feature.book.BookService
import com.storytrim.app.data.dto.PointsBalanceResp
import com.storytrim.app.data.dto.PointsLedgerResp
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class PointsRepository @Inject constructor(
    private val bookService: BookService,
    private val apiClient: ApiClient
) {
    suspend fun getBalance(): Result<PointsBalanceResp> {
        return try {
            val res = apiClient.safeCall { bookService.getPointsBalance() }
            if (res.code == 0 && res.data != null) {
                Result.success(res.data)
            } else {
                Result.failure(Exception(res.msg))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getLedger(page: Int, size: Int): Result<PointsLedgerResp> {
        return try {
            val res = apiClient.safeCall { bookService.getPointsLedger(page, size) }
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
