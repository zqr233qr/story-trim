package com.storytrim.app.ui.shelf

import android.content.Context
import android.net.Uri
import android.provider.OpenableColumns
import android.widget.Toast
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
import kotlinx.coroutines.flow.MutableStateFlow
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
    val books: LiveData<List<Book>> = _forceRefresh.flatMapLatest { _ ->
        bookRepository.getBooksStream(0)
    }.asLiveData()

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

                // 获取输入流
                val inputStream = contentResolver.openInputStream(uri)
                if (inputStream != null) {
                    // 开始导入
                    withContext(Dispatchers.Main) {
                        Toast.makeText(context, "正在解析 $fileName...", Toast.LENGTH_SHORT).show()
                    }
                    
                    val result = bookRepository.importBook(inputStream, fileName)
                    
                    withContext(Dispatchers.Main) {
                        if (result.isSuccess) {
                            Toast.makeText(context, "导入成功", Toast.LENGTH_SHORT).show()
                        } else {
                            Toast.makeText(context, "导入失败: ${result.exceptionOrNull()?.message}", Toast.LENGTH_LONG).show()
                        }
                    }
                }
            } catch (e: Exception) {
                withContext(Dispatchers.Main) {
                    Toast.makeText(context, "导入出错: ${e.message}", Toast.LENGTH_LONG).show()
                }
            }
        }
    }
}
