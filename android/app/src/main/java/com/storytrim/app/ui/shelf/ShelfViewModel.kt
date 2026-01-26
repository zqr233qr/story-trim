package com.storytrim.app.ui.shelf

import android.content.Context
import android.net.Uri
import android.provider.OpenableColumns
import com.storytrim.app.ui.common.ToastHelper
import androidx.lifecycle.LiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.asLiveData
import androidx.lifecycle.viewModelScope
import com.storytrim.app.data.model.Book
import com.storytrim.app.data.repository.BookRepository
import com.storytrim.app.data.repository.AuthRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.flatMapLatest
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import javax.inject.Inject

@HiltViewModel
class ShelfViewModel @Inject constructor(
    private val bookRepository: BookRepository,
    private val authRepository: AuthRepository,
    @ApplicationContext private val context: Context
) : ViewModel() {

    private val _forceRefresh = MutableStateFlow(0)
    
    // Explicitly typing the Flow helps type inference
    @OptIn(ExperimentalCoroutinesApi::class)
    val books: LiveData<List<Book>> = _forceRefresh.flatMapLatest { _ ->
        bookRepository.getBooksStream(0)
    }.asLiveData()

    val isLoggedIn: LiveData<Boolean> = authRepository.isLoggedInFlow()
        .map { it.isNotBlank() }
        .asLiveData()

    init {
        refresh()
    }

    fun refresh() {
        viewModelScope.launch {
            bookRepository.refreshBooks(0)
        }
    }
    
    fun deleteBook(book: Book) {
        viewModelScope.launch {
            bookRepository.deleteBook(book.id)
        }
    }

    fun downloadBook(book: Book, onProgress: (Int) -> Unit) {
        viewModelScope.launch {
            try {
                onProgress(1)
                val result = bookRepository.downloadBookContent(book.id, book.cloudId, book.userId, onProgress)
                withContext(Dispatchers.Main) {
                    if (result.isSuccess) {
                        ToastHelper.show(context, "下载完成")
                        refresh()
                    } else {
                        ToastHelper.show(context, "下载失败: ${result.exceptionOrNull()?.message}", long = true)
                    }
                }
            } catch (e: Exception) {
                withContext(Dispatchers.Main) {
                    ToastHelper.show(context, "下载失败: ${e.message}", long = true)
                }
            } finally {
                onProgress(0)
            }
        }
    }

    fun syncBook(book: Book, onProgress: (Int) -> Unit) {
        viewModelScope.launch {
            try {
                onProgress(1)
                val result = bookRepository.uploadBookZip(book.id, onProgress)
                withContext(Dispatchers.Main) {
                    if (result.isSuccess) {
                        ToastHelper.show(context, "同步成功")
                        refresh()
                    } else {
                        val message = result.exceptionOrNull()?.message.orEmpty()
                        val display = when {
                            message.contains("已存在") || message.contains("1001") -> "云端已存在本书"
                            message.contains("401") || message.contains("未授权") -> "登录已过期，请重新登录"
                            message.isNotBlank() -> "同步失败: $message"
                            else -> "同步失败"
                        }
                        ToastHelper.show(context, display, long = true)
                    }
                }
            } catch (e: Exception) {
                withContext(Dispatchers.Main) {
                    val message = e.message.orEmpty()
                    val display = when {
                        message.contains("已存在") || message.contains("1001") -> "云端已存在本书"
                        message.contains("401") || message.contains("未授权") -> "登录已过期，请重新登录"
                        message.isNotBlank() -> "同步失败: $message"
                        else -> "同步失败"
                    }
                    ToastHelper.show(context, display, long = true)
                }
            } finally {
                onProgress(0)
            }
        }
    }

    fun importBook(uri: Uri) {
        viewModelScope.launch(Dispatchers.IO) {
            try {
                val contentResolver = context.contentResolver
                
                // 获取文件名
                var fileName = "Unknown.txt"
                contentResolver.query(uri, null, null, null, null)?.use { cursor ->
                    if (cursor.moveToFirst()) {
                        val nameIndex = cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME)
                        if (nameIndex >= 0) {
                            fileName = cursor.getString(nameIndex)
                        }
                    }
                }

                val lowerName = fileName.lowercase()
                if (!lowerName.endsWith(".txt") && !lowerName.endsWith(".epub")) {
                    withContext(Dispatchers.Main) {
                        ToastHelper.show(context, "不支持的文件格式，仅支持 TXT 或 EPUB")
                    }
                    return@launch
                }

                // 获取输入流
                val inputStream = contentResolver.openInputStream(uri)
                if (inputStream != null) {
                    // 开始导入
                    withContext(Dispatchers.Main) {
                        ToastHelper.show(context, "正在解析 $fileName...")
                    }
                    
                    val result = bookRepository.importBook(inputStream, fileName)
                    
                    withContext(Dispatchers.Main) {
                        if (result.isSuccess) {
                            ToastHelper.show(context, "导入成功")
                        } else {
                            ToastHelper.show(context, "导入失败: ${result.exceptionOrNull()?.message}", long = true)
                        }
                    }
                }
            } catch (e: Exception) {
                withContext(Dispatchers.Main) {
                    ToastHelper.show(context, "导入出错: ${e.message}", long = true)
                }
            }
        }
    }
}
