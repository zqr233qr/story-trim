package com.storytrim.app.ui.reader

import android.content.Context
import dagger.hilt.android.qualifiers.ApplicationContext
import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.storytrim.app.data.model.Book
import com.storytrim.app.data.model.Chapter
import com.storytrim.app.data.model.Prompt
import com.storytrim.app.data.repository.BookRepository
import com.storytrim.app.ui.reader.tts.TtsController
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.Job
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class ReaderViewModel @Inject constructor(
    val bookRepository: BookRepository,
    @ApplicationContext private val appContext: Context
) : ViewModel() {

    private val _book = MutableLiveData<Book>()
    val book: LiveData<Book> = _book

    private val _chapters = MutableLiveData<List<Chapter>>()
    val chapters: LiveData<List<Chapter>> = _chapters
    
    var currentChapterIndex = 0
        private set

    data class ContentState(val title: String, val text: String, val scrollTo: NavigationStrategy)
    enum class NavigationStrategy { START, END, KEEP }

    private val _contentState = MutableLiveData<ContentState>()
    val contentState: LiveData<ContentState> = _contentState
    
    private val _isLoading = MutableLiveData<Boolean>()
    val isLoading: LiveData<Boolean> = _isLoading

    private val _isMenuVisible = MutableLiveData(false)
    val isMenuVisible: LiveData<Boolean> = _isMenuVisible

    private val _isMagicActive = MutableLiveData(false)
    val isMagicActive: LiveData<Boolean> = _isMagicActive

    var userPreferredModeId = 0
        private set

    private val prefs by lazy {
        appContext.getSharedPreferences("reader_prefs", Context.MODE_PRIVATE)
    }

    private val prefKeyUserPreferredModeId = "user_preferred_mode_id"

    private val _activeModeName = MutableLiveData("原文")
    val activeModeName: LiveData<String> = _activeModeName

    private val _toastMessage = MutableLiveData<String>()
    val toastMessage: LiveData<String> = _toastMessage

    private val _showModeConfigEvent = MutableLiveData<Unit>()
    val showModeConfigEvent: LiveData<Unit> = _showModeConfigEvent

    private val _prompts = MutableLiveData<List<Prompt>>()
    val prompts: LiveData<List<Prompt>> = _prompts
    
    var currentModeId = 0
        private set
        
    private val _isGenerating = MutableLiveData(false)
    val isGenerating: LiveData<Boolean> = _isGenerating
    
    private val _generationStream = MutableLiveData("")
    val generationStream: LiveData<String> = _generationStream
    
    private val _showTerminalEvent = MutableLiveData<Boolean>()
    val showTerminalEvent: LiveData<Boolean> = _showTerminalEvent

    // TTS相关
    private val _ttsController = MutableLiveData<TtsController?>()
    val ttsController: LiveData<TtsController?> = _ttsController

    private val _ttsState = MutableLiveData<TtsController.TtsState>()
    val ttsState: LiveData<TtsController.TtsState> = _ttsState

    private val _currentSentenceIndex = MutableLiveData(-1)
    val currentSentenceIndex: LiveData<Int> = _currentSentenceIndex

    private val _isTtsPanelVisible = MutableLiveData(false)
    val isTtsPanelVisible: LiveData<Boolean> = _isTtsPanelVisible

    private var progressJob: Job? = null
    private var recordedChapterId: Long? = null
    private var historyPromptId: Int = 0
    private var suppressToast = false
    private var lastRenderKey: String? = null
    private var lastRenderModeId: Int? = null

    fun loadBook(bookId: Long) {
        _isLoading.value = true
        viewModelScope.launch {
            android.util.Log.d("ReaderViewModel", "Loading book: $bookId")
            bookRepository.getBookDetail(bookId).onSuccess { b ->
                _book.value = b
                android.util.Log.d("ReaderViewModel", "Book loaded: ${b.title}, cloudId: ${b.cloudId}, syncState: ${b.syncState}")
                val list = bookRepository.getChapters(bookId)
                android.util.Log.d("ReaderViewModel", "Chapters loaded: ${list.size}")
                _chapters.value = list
                val restoreResult = restoreReadingState(b, list)
                currentChapterIndex = restoreResult.first
                historyPromptId = restoreResult.second
                loadPrompts()
                if (list.isNotEmpty()) {
                    loadChapterContent(currentChapterIndex, NavigationStrategy.START, forcePreferred = true)
                } else {
                    _isLoading.value = false
                }
            }.onFailure {
                android.util.Log.e("ReaderViewModel", "Failed to load book", it)
                _isLoading.value = false
            }
        }
    }

    fun loadPrompts() {
        viewModelScope.launch {
            _prompts.value = bookRepository.getPrompts()
            initUserPreferredMode()
            if (historyPromptId > 0) {
                currentModeId = historyPromptId
                _isMagicActive.value = true
                val name = getPromptName(historyPromptId)
                if (name.isNotBlank()) {
                    _activeModeName.value = name
                }
            }
            applyPreferredModeIfNeeded(NavigationStrategy.KEEP, force = true)
        }
    }

    /**
     * 初始化用户偏好模式
     */
    private fun initUserPreferredMode() {
        val saved = prefs.getInt(prefKeyUserPreferredModeId, 0)
        if (saved > 0) {
            userPreferredModeId = saved
            return
        }

        val defaultPrompt = _prompts.value?.find { it.isDefault } ?: _prompts.value?.firstOrNull()
        if (defaultPrompt != null) {
            userPreferredModeId = defaultPrompt.id
            prefs.edit().putInt(prefKeyUserPreferredModeId, defaultPrompt.id).apply()
        }
    }
    
    fun toggleMagic() {
        android.util.Log.d("ReaderViewModel", "toggleMagic called, currentModeId=$currentModeId, userPreferredModeId=$userPreferredModeId")
        if (currentModeId != 0) {
            currentModeId = 0
            _activeModeName.value = "原文"
            _isMagicActive.value = false
            viewModelScope.launch { loadChapterContent(currentChapterIndex, NavigationStrategy.KEEP, forcePreferred = false) }
        } else {
            if (userPreferredModeId > 0) {
                viewModelScope.launch {
                    val chapter = _chapters.value?.getOrNull(currentChapterIndex) ?: return@launch
                    val applied = applyPreferredModeForChapter(chapter, NavigationStrategy.KEEP, force = true)
                    if (!applied) {
                        _showModeConfigEvent.postValue(Unit)
                        loadChapterContent(currentChapterIndex, NavigationStrategy.KEEP, forcePreferred = false)
                    } else {
                        val promptName = getPromptName(userPreferredModeId)
                        if (promptName.isNotBlank()) {
                            // toast handled on content change
                        }
                    }
                }
            } else {
                if (_prompts.value.isNullOrEmpty()) {
                    loadPrompts()
                }
                _showModeConfigEvent.postValue(Unit)
            }
        }
    }

    /**
     * 更新用户偏好模式
     */
    fun updateUserPreferredMode(modeId: Int) {
        userPreferredModeId = modeId
        prefs.edit().putInt(prefKeyUserPreferredModeId, modeId).apply()
        viewModelScope.launch {
            val applied = applyPreferredModeIfNeeded(NavigationStrategy.KEEP, force = true)
            val promptName = getPromptName(modeId)
            if (!applied) {
                if (promptName.isNotBlank()) {
                    pushToast("偏好已更新为「$promptName」，当前章节暂无该模式，已显示原文")
                } else {
                    pushToast("当前章节暂无该偏好模式，已显示原文")
                }
            } else if (promptName.isNotBlank()) {
                pushToast("偏好已更新为「$promptName」")
            }
        }
    }

    fun switchMode(modeId: Int) {
        if (modeId == currentModeId) return
        currentModeId = modeId
        _isMagicActive.value = modeId != 0

        val modeName = if (modeId == 0) {
            "原文"
        } else {
            val prompt = _prompts.value?.find { it.id == modeId }
            prompt?.name ?: "未知模式"
        }
        _activeModeName.value = modeName

        if (modeId == 0) {
            viewModelScope.launch { loadChapterContent(currentChapterIndex, NavigationStrategy.KEEP, forcePreferred = false) }
            return
        }
        viewModelScope.launch {
            val bId = book.value?.id ?: return@launch
            val list = _chapters.value ?: return@launch
            val chapter = list.getOrNull(currentChapterIndex) ?: return@launch
            val cached = bookRepository.fetchChapterTrim(bId, chapter, modeId)
            if (!cached.isNullOrBlank()) {
                renderContent(chapter.title, cached, NavigationStrategy.KEEP)
            }
            else startGeneration(bId, chapter.id, modeId)
        }
    }
    
    private suspend fun startGeneration(bookId: Long, chapterId: Long, promptId: Int) {
        _isGenerating.value = true
        _generationStream.value = ""
        _showTerminalEvent.value = true
        val builder = StringBuilder()
        
        val activeBook = _book.value
        val activeChapter = _chapters.value?.getOrNull(currentChapterIndex)
        
        if (activeBook == null || activeChapter == null) {
            _isGenerating.value = false
            return
        }

        val onData: (String) -> Unit = { chunk -> 
            builder.append(chunk)
            _generationStream.postValue(builder.toString()) 
        }
        val onError: (String) -> Unit = { _ -> 
            viewModelScope.launch {
                fallbackToOriginal(activeChapter, NavigationStrategy.KEEP, "精简失败，已返回原文")
            }
        }
        val onClosed: () -> Unit = {
            viewModelScope.launch {
                val final = builder.toString()
                if (final.isNotEmpty()) {
                    bookRepository.cacheChapterTrim(bookId, chapterId, promptId, final)
                    delay(500)
                    _isGenerating.value = false
                    _showTerminalEvent.value = false
                    _chapters.value?.getOrNull(currentChapterIndex)?.let {
                        renderContent(it.title, final, NavigationStrategy.KEEP)
                    }
                } else {
                    fallbackToOriginal(activeChapter, NavigationStrategy.KEEP, "精简无结果，已返回原文")
                }
            }
        }

        // Logic split based on syncState
        if (activeBook.syncState == 0) {
            // Local Book: Use connectByMd5
            val content = bookRepository.getChapterContent(chapterId)
            if (content.isEmpty()) {
                _isGenerating.value = false
                return
            }
            
            bookRepository.startTrimStreamByMd5(
                content = content,
                md5 = activeChapter.md5,
                promptId = promptId,
                bookMd5 = activeBook.bookMd5 ?: "",
                bookTitle = activeBook.title,
                chapterTitle = activeChapter.title,
                chapterIndex = activeChapter.index,
                onData = onData,
                onError = onError,
                onClosed = onClosed
            )
        } else {
            // sync_state=1/2: 走 ID 流式
            val cloudBookId = if (activeBook.cloudId > 0) activeBook.cloudId else activeBook.id
            val cloudChapterId = if (activeChapter.cloudId > 0) activeChapter.cloudId else activeChapter.id
            
            bookRepository.startTrimStream(cloudBookId, cloudChapterId, promptId, onData, onError, onClosed)
        }
    }

    /**
     * 精简失败兜底：恢复原文并提示
     */
    private suspend fun fallbackToOriginal(chapter: Chapter, strategy: NavigationStrategy, message: String) {
        currentModeId = 0
        _isMagicActive.postValue(false)
        _activeModeName.postValue("原文")

        val content = bookRepository.getChapterContent(chapter.id)
        if (content.isNotBlank()) {
            renderContent(chapter.title, content, strategy)
        }
        pushToast(message)
        _showTerminalEvent.postValue(false)
        _isGenerating.postValue(false)
    }
    
    private fun renderContent(title: String, text: String, strategy: NavigationStrategy) {
        val formatted = StringBuilder()
        text.split("\n").forEach { if (it.isNotBlank()) formatted.append("    ").append(it.trim()).append("\n") }
        _contentState.value = ContentState(title, formatted.toString(), strategy)

        _ttsController.value?.prepare(formatted.toString(), title)
        handleContentChange(text)
    }

    private fun handleContentChange(rawText: String) {
        val chapter = _chapters.value?.getOrNull(currentChapterIndex) ?: return
        val key = "${chapter.id}_${currentModeId}_${rawText.hashCode()}"
        if (key == lastRenderKey) return

        val previousMode = lastRenderModeId
        lastRenderKey = key
        lastRenderModeId = currentModeId

        if (currentModeId == 0) {
            if (previousMode != null && previousMode != 0) {
                pushToast("已切换为原文")
            }
            return
        }

        val promptName = getPromptName(currentModeId)
        if (promptName.isBlank()) return

        viewModelScope.launch {
            val original = bookRepository.getChapterContent(chapter.id)
            if (original.isBlank()) return@launch
            val ratio = calculateTrimRatio(original, rawText)
            pushToast("已切换为「$promptName」，精简 ${ratio}%")
        }
    }

    private suspend fun loadChapterContent(index: Int, strategy: NavigationStrategy, forcePreferred: Boolean) {
        val list = _chapters.value ?: return
        if (index !in list.indices) return
        currentChapterIndex = index
        val chapter = list[index]
        _isLoading.value = true
        
        android.util.Log.d("ReaderViewModel", "Loading chapter content: ${chapter.title} (idx: $index)")

        if (historyPromptId > 0 && currentModeId == historyPromptId) {
            val trimmed = bookRepository.fetchChapterTrim(bookId = chapter.bookId, chapter = chapter, promptId = historyPromptId)
            if (!trimmed.isNullOrBlank()) {
                _isMagicActive.value = true
                _activeModeName.value = getPromptName(historyPromptId).ifBlank { _activeModeName.value }
                renderContent(chapter.title, trimmed, strategy)
                scheduleProgressUpdate(chapter)
                preloadAdjacentChapters(list, index)
                _isLoading.value = false
                return
            } else {
                currentModeId = 0
                _isMagicActive.value = false
                _activeModeName.value = "原文"
            }
        }

        if (applyPreferredModeForChapter(chapter, strategy, force = forcePreferred)) {
            scheduleProgressUpdate(chapter)
            preloadAdjacentChapters(list, index)
            _isLoading.value = false
            return
        }
        val content = bookRepository.getChapterContent(chapter.id)
        android.util.Log.d("ReaderViewModel", "Content loaded, length: ${content.length}")
        
        if (content.isEmpty()) {
             renderContent(chapter.title, "暂无内容 (如果是云端书籍，请检查网络或稍后重试)", strategy)
        } else {
             renderContent(chapter.title, content, strategy)
        }
        scheduleProgressUpdate(chapter)
        preloadAdjacentChapters(list, index)
        _isLoading.value = false
    }

    /**
     * 恢复阅读进度（本地优先，必要时与云端比较）
     * @return Pair(startIndex, promptId)
     */
    private suspend fun restoreReadingState(book: Book, chapters: List<Chapter>): Pair<Int, Int> {
        val local = bookRepository.getLocalReadingHistory(book.id)
        if (book.syncState == 0) {
            return resolveHistoryResult(chapters, local)
        }

        val cloud = bookRepository.getCloudReadingHistory(book)
        val selected = when {
            cloud == null -> local
            local == null -> cloud
            cloud.updatedAt > local.updatedAt -> cloud
            else -> local
        }
        return resolveHistoryResult(chapters, selected)
    }

    private fun resolveHistoryResult(chapters: List<Chapter>, history: com.storytrim.app.core.database.entity.ReadingHistoryEntity?): Pair<Int, Int> {
        if (history == null) return 0 to 0
        val index = chapters.indexOfFirst { it.id == history.lastChapterId }
        val safeIndex = if (index >= 0) index else 0
        return safeIndex to history.lastPromptId
    }

    /**
     * 预加载相邻章节内容（前后各2章）
     */
    private fun preloadAdjacentChapters(chapters: List<Chapter>, currentIndex: Int) {
        val indices = (currentIndex - 2..currentIndex + 2)
            .filter { it in chapters.indices && it != currentIndex }
        val targets = indices.map { chapters[it] }
        viewModelScope.launch {
            bookRepository.preloadChapterContents(targets)
            if (currentModeId > 0) {
                bookRepository.preloadChapterTrims(targets, currentModeId)
            }
        }
    }

    /**
     * 进度确认逻辑：延迟5秒后确认仍停留在当前章节
     */
    private fun scheduleProgressUpdate(chapter: Chapter) {
        if (recordedChapterId == chapter.id) return
        progressJob?.cancel()
        progressJob = viewModelScope.launch {
            delay(5000)
            val activeChapter = _chapters.value?.getOrNull(currentChapterIndex)
            if (activeChapter?.id == chapter.id) {
                val bookId = _book.value?.id ?: return@launch
                val promptId = currentModeId
                bookRepository.updateReadingProgress(bookId, chapter.id, promptId)
                recordedChapterId = chapter.id
            }
        }
    }

    private suspend fun applyPreferredModeIfNeeded(strategy: NavigationStrategy, force: Boolean): Boolean {
        val chapter = _chapters.value?.getOrNull(currentChapterIndex) ?: return false
        val applied = applyPreferredModeForChapter(chapter, strategy, force)
        if (!applied && force) {
            loadChapterContent(currentChapterIndex, strategy, forcePreferred = true)
        }
        return applied
    }

    private suspend fun applyPreferredModeForChapter(
        chapter: Chapter,
        strategy: NavigationStrategy,
        force: Boolean
    ): Boolean {
        if (userPreferredModeId <= 0) {
            currentModeId = 0
            _activeModeName.value = "原文"
            _isMagicActive.value = false
            return false
        }

        if (!force) {
            return false
        }

        val currentBook = book.value ?: return false
        val bookId = currentBook.id
        val preferredModeId = userPreferredModeId

        val promptIds = if (currentBook.syncState == 0 && chapter.md5.isNotBlank()) {
            bookRepository.getChapterTrimStatusByMd5(chapter.md5)
        } else {
            bookRepository.getChapterTrimStatusById(chapter, currentBook.bookMd5)
        }

        if (!promptIds.contains(preferredModeId)) {
            currentModeId = 0
            _activeModeName.value = "原文"
            _isMagicActive.value = false
            return false
        }

        val trimmed = bookRepository.fetchChapterTrim(bookId, chapter, preferredModeId)
        if (!trimmed.isNullOrBlank()) {
            currentModeId = preferredModeId
            _activeModeName.value = _prompts.value?.find { it.id == preferredModeId }?.name ?: "未知模式"
            _isMagicActive.value = true
            renderContent(chapter.title, trimmed, strategy)
            return true
        }

        currentModeId = 0
        _activeModeName.value = "原文"
        _isMagicActive.value = false
        return false
    }

    private fun getPromptName(promptId: Int): String {
        return _prompts.value?.find { it.id == promptId }?.name ?: ""
    }

    private fun calculateTrimRatio(original: String, trimmed: String): Int {
        if (original.isBlank() || trimmed.isBlank()) return 0
        val originalChars = original.replace("\n", "").replace("\r", "").length
        val trimmedChars = trimmed.replace("\n", "").replace("\r", "").length
        if (originalChars == 0) return 0
        val ratio = Math.round((1f - trimmedChars.toFloat() / originalChars.toFloat()) * 100f)
        return maxOf(0, ratio)
    }

    fun setToastSuppressed(suppressed: Boolean) {
        suppressToast = suppressed
    }

    private fun pushToast(message: String) {
        if (message.isBlank() || suppressToast) return
        _toastMessage.postValue(message)
    }

    fun toggleMenu() { _isMenuVisible.value = !(_isMenuVisible.value ?: false) }
    fun loadNextChapter() { _chapters.value?.let { if (currentChapterIndex < it.size - 1) viewModelScope.launch { loadChapterContent(currentChapterIndex + 1, NavigationStrategy.START, forcePreferred = true) } } }
    fun loadPrevChapter() { if (currentChapterIndex > 0) viewModelScope.launch { loadChapterContent(currentChapterIndex - 1, NavigationStrategy.END, forcePreferred = true) } }
    fun jumpToChapter(index: Int) { viewModelScope.launch { loadChapterContent(index, NavigationStrategy.START, forcePreferred = true) } }

    // ==================== TTS 听书功能 ====================

    /**
     * 初始化TTS控制器
     */
    fun initTtsController(context: Context) {
        if (_ttsController.value == null) {
            _ttsController.value = TtsController(context)
        }
    }

    /**
     * 获取书籍名称
     */
    private fun getBookTitle(): String = _book.value?.title ?: "听书"

    /**
     * 准备听书内容
     */
    fun prepareTts(text: String, chapterTitle: String) {
        val bookTitle = getBookTitle()
        _ttsController.value?.prepare(text, chapterTitle)
        _ttsController.value?.setBookTitle(bookTitle)
    }

    /**
     * 开始播放
     */
    fun startTts() {
        val bookTitle = getBookTitle()
        _ttsController.value?.start(bookTitle = bookTitle)
    }

    /**
     * 暂停/继续切换
     */
    fun toggleTts() {
        _ttsController.value?.togglePlayPause()
    }

    /**
     * 停止听书
     */
    fun stopTts() {
        _ttsController.value?.stop()
    }

    /**
     * 上一句
     */
    fun ttsPrevious() {
        _ttsController.value?.previous()
    }

    /**
     * 下一句
     */
    fun ttsNext() {
        _ttsController.value?.next()
    }

    /**
     * 跳转到指定句子
     */
    fun ttsSeekTo(index: Int) {
        _ttsController.value?.seekTo(index)
    }

    /**
     * 设置语速
     */
    fun setTtsSpeechRate(rate: Float) {
        _ttsController.value?.setSpeechRate(rate)
    }

    /**
     * 获取当前播放位置对应的原文位置
     */
    fun getTtsTextPosition(): Int {
        return _ttsController.value?.getCurrentTextPosition() ?: -1
    }

    /**
     * 是否正在播放
     */
    fun isTtsPlaying(): Boolean {
        return _ttsController.value?.isPlaying() ?: false
    }

    /**
     * 显示/隐藏控制面板
     */
    fun toggleTtsPanel() {
        _isTtsPanelVisible.value = !(_isTtsPanelVisible.value ?: false)
    }

    fun showTtsPanel() {
        _isTtsPanelVisible.value = true
    }

    fun hideTtsPanel() {
        _isTtsPanelVisible.value = false
    }

    /**
     * 释放TTS资源
     */
    fun releaseTts() {
        _ttsController.value?.release()
        _ttsController.value = null
    }
}
