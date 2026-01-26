package com.storytrim.app.ui.tasks

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.storytrim.app.data.repository.TaskRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class TaskCenterViewModel @Inject constructor(
    private val taskRepository: TaskRepository
) : ViewModel() {

    private val _tasks = MutableLiveData<List<TaskItemUi>>(emptyList())
    val tasks: LiveData<List<TaskItemUi>> = _tasks

    private val _isLoading = MutableLiveData(false)
    val isLoading: LiveData<Boolean> = _isLoading

    private var pollJob: Job? = null

    fun startPolling() {
        if (pollJob?.isActive == true) return
        pollJob = viewModelScope.launch {
            while (isActive) {
                loadTasks()
                delay(3000)
            }
        }
    }

    fun stopPolling() {
        pollJob?.cancel()
        pollJob = null
    }

    private suspend fun loadTasks() {
        _isLoading.postValue(true)
        val result = taskRepository.getActiveTasks()
        if (result.isSuccess) {
            val items = result.getOrNull().orEmpty().map { task ->
                TaskItemUi(
                    id = task.id,
                    title = task.bookTitle,
                    subtitle = task.promptName.ifBlank { task.status },
                    progress = task.progress,
                    status = statusLabel(task.status, task.error)
                )
            }
            if (items.isEmpty()) {
                _tasks.postValue(demoTasks())
            } else {
                _tasks.postValue(items)
            }
        } else {
            _tasks.postValue(demoTasks())
        }
        _isLoading.postValue(false)
    }

    private fun statusLabel(status: String, error: String?): String {
        return when (status) {
            "running" -> "处理中"
            "pending" -> "等待中"
            "completed" -> "已完成"
            "failed" -> error?.let { "失败：$it" } ?: "失败"
            else -> status
        }
    }

    private fun demoTasks(): List<TaskItemUi> {
        return listOf(
            TaskItemUi(
                id = "demo-1",
                title = "全书精简 · 标准模式",
                subtitle = "《示例书籍》 · 50 章",
                progress = 0,
                status = "等待中"
            ),
            TaskItemUi(
                id = "demo-2",
                title = "批量精简 · 精简模式",
                subtitle = "《示例书籍》 · 20 章",
                progress = 0,
                status = "等待中"
            )
        )
    }
}
