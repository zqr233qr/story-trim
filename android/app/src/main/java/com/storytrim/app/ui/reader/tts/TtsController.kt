package com.storytrim.app.ui.reader.tts

import android.Manifest
import android.content.ComponentName
import android.content.Context
import android.content.Intent
import android.content.ServiceConnection
import android.content.pm.PackageManager
import android.os.Build
import android.os.IBinder
import android.util.Log
import androidx.activity.result.ActivityResultLauncher
import androidx.activity.result.contract.ActivityResultContracts
import androidx.core.content.ContextCompat
import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch

class TtsController(
    private val context: Context
) {
    companion object {
        private const val TAG = "TtsController"
        const val ACTION_PLAY_PAUSE = "com.storytrim.widget.PLAY_PAUSE"
        const val ACTION_NEXT = "com.storytrim.widget.NEXT"
        const val ACTION_PREV = "com.storytrim.widget.PREV"
    }

    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Main)
    private val appContext = context.applicationContext

    private var ttsService: TtsForegroundService? = null
    private var serviceBound = false
    private var pendingStartAfterPermission: Boolean = false
    private val segmenter = SentenceSegmenter()

    // 悬浮窗管理器
    val floatingLyricsManager = FloatingLyricsManager(context)

    // 通知权限请求
    private var notificationPermissionLauncher: ActivityResultLauncher<String>? = null

    private var sentenceList: List<String> = emptyList()
    private var sentencePositions: List<Int> = emptyList()
    private var isPreparing = false

    // 书籍标题
    private var bookTitle: String = ""
    private var chapterTitle: String = ""

    private val _ttsState = MutableLiveData<TtsState>(TtsState.IDLE)
    val ttsState: LiveData<TtsState> = _ttsState

    private val _currentSentenceIndex = MutableLiveData(-1)
    val currentSentenceIndex: LiveData<Int> = _currentSentenceIndex

    private val _currentSentence = MutableLiveData<String?>()
    val currentSentence: LiveData<String?> = _currentSentence

    private val _currentSentenceWithIndex = MutableLiveData<TtsSentence?>()
    val currentSentenceWithIndex: LiveData<TtsSentence?> = _currentSentenceWithIndex

    private val _sentences = MutableLiveData<List<String>>()
    val sentences: LiveData<List<String>> = _sentences

    private val _progress = MutableLiveData<TtsProgress>()
    val progress: LiveData<TtsProgress> = _progress

    private val _isTtsReady = MutableLiveData(false)
    val isTtsReady: LiveData<Boolean> = _isTtsReady

    enum class TtsState {
        IDLE, PLAYING, PAUSED, STOPPED, ERROR
    }

    data class TtsSentence(
        val index: Int,
        val sentence: String
    )

    data class TtsProgress(
        val currentIndex: Int,
        val totalCount: Int,
        val progressPercent: Float,
        val currentSentence: String
    )

    private val serviceCallback = object : TtsEngine.Callback {
        override fun onStart(sentenceIndex: Int, sentence: String) {
            Log.d(TAG, "onStart: $sentenceIndex, sentence: ${sentence.take(20)}...")
            _currentSentenceIndex.postValue(sentenceIndex)
            _currentSentence.postValue(sentence)
            _currentSentenceWithIndex.postValue(TtsSentence(sentenceIndex, sentence))
            updateProgress(sentenceIndex, sentence)
            _ttsState.postValue(TtsState.PLAYING)
            updateFloatingLyrics(sentence, true)
        }

        override fun onComplete(sentenceIndex: Int, sentence: String) {
            Log.d(TAG, "onComplete: $sentenceIndex")
            updateProgress(sentenceIndex, sentence)
            updateFloatingLyrics(sentence, true)
        }

        override fun onError(error: String) {
            Log.e(TAG, "onError: $error")
            _ttsState.postValue(TtsState.ERROR)
            updateFloatingLyrics("播放出错: $error", false)
        }

        override fun onAllComplete() {
            Log.d(TAG, "onAllComplete")
            _ttsState.postValue(TtsState.STOPPED)
            _currentSentenceIndex.postValue(-1)
            _currentSentence.postValue(null)
            updateFloatingLyrics("播放完成", false)
        }

        override fun onPrevChapter() {
            Log.d(TAG, "onPrevChapter")
            val sentence = _currentSentence.value ?: ""
            updateFloatingLyrics(sentence, true)
            previous()
        }

        override fun onNextChapter() {
            Log.d(TAG, "onNextChapter")
            val sentence = _currentSentence.value ?: ""
            updateFloatingLyrics(sentence, true)
            next()
        }
    }

    private val serviceConnection = object : ServiceConnection {
        override fun onServiceConnected(name: ComponentName?, service: IBinder?) {
            Log.d(TAG, "Service connected")
            val binder = service as TtsForegroundService.TtsBinder
            ttsService = binder.getService()
            ttsService?.setCallback(serviceCallback)
            serviceBound = true
            _isTtsReady.postValue(true)
        }

        override fun onServiceDisconnected(name: ComponentName?) {
            Log.d(TAG, "Service disconnected")
            ttsService = null
            serviceBound = false
            _isTtsReady.postValue(false)
        }
    }

    init {
        bindService()
        setupPermissionLauncher()
    }

    /**
     * 设置权限请求回调
     */
    private fun setupPermissionLauncher() {
        try {
            if (context is androidx.activity.ComponentActivity) {
                notificationPermissionLauncher = context.registerForActivityResult(
                    ActivityResultContracts.RequestPermission()
                ) { isGranted ->
                    Log.d(TAG, "Notification permission granted: $isGranted")
                    if (isGranted && pendingStartAfterPermission) {
                        pendingStartAfterPermission = false
                        start()
                    }
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to setup permission launcher", e)
        }
    }

    /**
     * 检查通知权限
     */
    private fun checkNotificationPermission(): Boolean {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            ContextCompat.checkSelfPermission(
                appContext,
                Manifest.permission.POST_NOTIFICATIONS
            ) == PackageManager.PERMISSION_GRANTED
        } else {
            true
        }
    }

    /**
     * 请求通知权限
     */
    private fun requestNotificationPermission(): Boolean {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            val launcher = notificationPermissionLauncher ?: return false
            if (context is androidx.activity.ComponentActivity) {
                launcher.launch(Manifest.permission.POST_NOTIFICATIONS)
                return true
            }
        }
        return false
    }

    private fun bindService() {
        val intent = Intent(appContext, TtsForegroundService::class.java)
        appContext.bindService(intent, serviceConnection, Context.BIND_AUTO_CREATE)
    }

    fun prepare(text: String, chapterTitle: String = "") {
        Log.d(TAG, "prepare() called with text length: ${text.length}")
        if (text.isBlank()) {
            Log.d(TAG, "prepare() text is blank, skipping")
            return
        }

        this.chapterTitle = chapterTitle
        isPreparing = true
        scope.launch(Dispatchers.Default) {
            sentenceList = segmenter.segment(text)
            Log.d(TAG, "segmented into ${sentenceList.size} sentences")
            _sentences.postValue(sentenceList)
            sentencePositions = segmenter.calculatePositions(sentenceList, text)
            _progress.postValue(TtsProgress(0, sentenceList.size, 0f, ""))
            _ttsState.postValue(TtsState.IDLE)
            isPreparing = false
            Log.d(TAG, "prepare() completed")
        }
    }

    fun start(startIndex: Int = 0, bookTitle: String = "") {
        Log.d(TAG, "start() called, serviceBound=$serviceBound")

        // Android 13+ 需要通知权限才能显示前台服务通知
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            if (!checkNotificationPermission()) {
                Log.w(TAG, "Notification permission not granted, requesting...")
                pendingStartAfterPermission = true
                val requested = requestNotificationPermission()
                if (!requested) {
                    Log.w(TAG, "Failed to request notification permission")
                    _ttsState.value = TtsState.ERROR
                }
                return
            }
        }

        if (!serviceBound) {
            Log.w(TAG, "Service not bound, waiting...")
            _ttsState.value = TtsState.ERROR
            return
        }

        if (sentenceList.isEmpty() && !isPreparing) {
            Log.e(TAG, "start() failed: sentenceList is empty")
            _ttsState.value = TtsState.ERROR
            return
        }

        val service = ttsService
        if (service != null) {
            Log.d(TAG, "start() speaking from index: $startIndex, bookTitle: $bookTitle, chapterTitle: $chapterTitle")
            service.speak(sentenceList, startIndex, bookTitle, chapterTitle)
        } else {
            Log.e(TAG, "start() failed: ttsService is null")
            _ttsState.value = TtsState.ERROR
        }
    }

    fun togglePlayPause() {
        val service = ttsService ?: run {
            start(bookTitle = bookTitle)
            return
        }
        // 切换通过状态判断
        when (_ttsState.value) {
            TtsState.PLAYING -> pause()
            TtsState.PAUSED -> resume()
            TtsState.IDLE, TtsState.STOPPED, TtsState.ERROR -> start(bookTitle = bookTitle)
            else -> start(bookTitle = bookTitle)
        }
    }

    fun pause() {
        ttsService?.stop()
        _ttsState.value = TtsState.PAUSED
    }

    fun resume() {
        start()
    }

    fun stop() {
        ttsService?.stop()
        _currentSentenceIndex.value = -1
        _currentSentence.value = null
        _ttsState.value = TtsState.STOPPED
    }

    fun seekTo(index: Int) {
        if (sentenceList.isEmpty()) return
        val targetIndex = index.coerceIn(0, sentenceList.size - 1)
        _currentSentenceIndex.value = targetIndex
        val sentence = if (targetIndex in sentenceList.indices) {
            sentenceList[targetIndex]
        } else ""
        _currentSentence.value = sentence
        updateProgress(targetIndex, sentence)
        ttsService?.speak(sentenceList, targetIndex, "")
    }

    fun previous() {
        val currentIndex = _currentSentenceIndex.value ?: 0
        if (currentIndex > 0) {
            seekTo(currentIndex - 1)
        }
    }

    fun next() {
        val currentIndex = _currentSentenceIndex.value ?: 0
        if (currentIndex < sentenceList.size - 1) {
            seekTo(currentIndex + 1)
        }
    }

    fun setSpeechRate(rate: Float) {
        ttsService?.setSpeechRate(rate)
    }

    fun getSpeechRate(): Float = 1.0f

    fun isPlaying(): Boolean {
        return _ttsState.value == TtsState.PLAYING
    }

    fun getCurrentTextPosition(): Int {
        val index = _currentSentenceIndex.value ?: return -1
        return if (index in sentencePositions.indices) sentencePositions[index] else -1
    }

    private fun updateProgress(sentenceIndex: Int, sentence: String) {
        if (sentenceList.isEmpty()) return
        val percent = (sentenceIndex + 1).toFloat() / sentenceList.size * 100
        _progress.postValue(TtsProgress(sentenceIndex, sentenceList.size, percent, sentence))
    }

    /**
      * 更新悬浮窗歌词
      */
     private fun updateFloatingLyrics(sentence: String, playing: Boolean) {
         try {
             floatingLyricsManager.updateLyrics(sentence, playing)
         } catch (e: Exception) {
             Log.e(TAG, "updateFloatingLyrics error", e)
         }
     }

    /**
     * 设置书籍标题
     */
    fun setBookTitle(title: String) {
        this.bookTitle = title
    }

    /**
     * 处理来自小组件的播放/暂停操作
     */
    fun handleWidgetPlayPause() {
        when (_ttsState.value) {
            TtsState.PLAYING -> pause()
            TtsState.PAUSED, TtsState.IDLE, TtsState.STOPPED, TtsState.ERROR -> start()
            else -> start()
        }
    }

    /**
     * 处理来自小组件的上一句操作
     */
    fun handleWidgetPrev() {
        previous()
    }

    /**
     * 处理来自小组件的下一句操作
     */
    fun handleWidgetNext() {
        next()
    }

    fun release() {
        Log.d(TAG, "release()")
        try {
            if (serviceBound) {
                ttsService?.setCallback(null)
                appContext.unbindService(serviceConnection)
                serviceBound = false
            }
            // 停止Service
            val intent = Intent(appContext, TtsForegroundService::class.java)
            appContext.stopService(intent)
        } catch (e: Exception) {
            Log.e(TAG, "release() error", e)
        }
        ttsService = null
        _isTtsReady.postValue(false)
        scope.cancel()
    }
}
