package com.storytrim.app.ui.reader.tts

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Intent
import android.media.AudioManager
import android.media.session.PlaybackState
import android.os.Binder
import android.os.Build
import android.os.Handler
import android.os.IBinder
import android.os.Looper
import android.speech.tts.TextToSpeech
import android.speech.tts.UtteranceProgressListener
import android.util.Log
import androidx.core.app.NotificationCompat
import androidx.media.app.NotificationCompat as MediaNotificationCompat
import android.support.v4.media.MediaMetadataCompat
import android.support.v4.media.session.MediaSessionCompat
import android.support.v4.media.session.PlaybackStateCompat
import com.storytrim.app.R
import com.storytrim.app.ui.reader.ReaderActivity
import java.util.concurrent.ConcurrentHashMap

class TtsForegroundService : Service() {

    companion object {
        private const val TAG = "TtsForegroundService"
        private const val CHANNEL_ID = "tts_playback_channel"
        private const val NOTIFICATION_ID = 1001
        private const val UTTERANCE_PREFIX = "tts_sentence_"
        private const val CLEAR_TTS_DELAY = 60000L
    }

    private val binder = TtsBinder()
    private val mainHandler = Handler(Looper.getMainLooper())
    private val clearTtsRunnable = Runnable { clearTts() }

    private var textToSpeech: TextToSpeech? = null
    private var callback: TtsEngine.Callback? = null
    private var onInit = false

    // 使用 Map 存储句子，索引 -> 句子，确保一一对应
    private val sentenceMap = ConcurrentHashMap<Int, String>()
    @Volatile private var currentSentenceIndex = -1
    @Volatile private var totalSentences = 0

    private var speechRate = 1.0f
    private var pitch = 1.0f

    private var audioManager: AudioManager? = null
    private var audioFocusGranted = false

    private var mediaSessionCompat: MediaSessionCompat? = null
    private var isPlaying = false
    private var bookTitle: String = "听书"
    private var chapterTitle: String = ""
    private var currentSentence: String = ""

    private val initListener = TextToSpeech.OnInitListener { status ->
        if (status == TextToSpeech.SUCCESS) {
            Log.d(TAG, "TTS init SUCCESS")
            textToSpeech?.setOnUtteranceProgressListener(utteranceListener)
            onInit = false
            startPlayback()
        } else {
            Log.e(TAG, "TTS init FAILED with status: $status")
            onInit = false
            callback?.onError("TTS引擎初始化失败($status)，请检查系统TTS设置")
        }
    }

    private val utteranceListener = object : UtteranceProgressListener() {
        override fun onStart(utteranceId: String?) {
            Log.d(TAG, "onStart: $utteranceId")
            mainHandler.removeCallbacks(clearTtsRunnable)
            try {
                val index = extractIndex(utteranceId)
                if (index >= 0) {
                    currentSentenceIndex = index
                    // 从Map中获取句子，确保与播放内容一致
                    currentSentence = sentenceMap[index] ?: ""
                    updateMediaMetadata()
                    callback?.onStart(index, currentSentence)
                }
            } catch (e: Exception) {
                Log.e(TAG, "onStart error", e)
            }
        }

        override fun onDone(utteranceId: String?) {
            Log.d(TAG, "onDone: $utteranceId")
            mainHandler.postDelayed(clearTtsRunnable, CLEAR_TTS_DELAY)
            try {
                val index = extractIndex(utteranceId)
                if (index >= 0) {
                    val sentence = sentenceMap[index] ?: ""
                    callback?.onComplete(index, sentence)
                }
                playNextSentence()
            } catch (e: Exception) {
                Log.e(TAG, "onDone error", e)
            }
        }

        @Deprecated("Deprecated in Java")
        override fun onError(utteranceId: String?) {
            Log.e(TAG, "onError deprecated: $utteranceId")
            try {
                callback?.onError("播放失败: $utteranceId")
                playNextSentence()
            } catch (e: Exception) {
                Log.e(TAG, "onError error", e)
            }
        }

        override fun onError(utteranceId: String?, errorCode: Int) {
            Log.e(TAG, "onError with code: $errorCode")
            try {
                callback?.onError("播放错误: $errorCode")
                playNextSentence()
            } catch (e: Exception) {
                Log.e(TAG, "onError with code error", e)
            }
        }
    }

    inner class TtsBinder : Binder() {
        fun getService(): TtsForegroundService = this@TtsForegroundService
    }

    override fun onCreate() {
        super.onCreate()
        Log.d(TAG, "onCreate()")
        createNotificationChannel()
        audioManager = getSystemService(AUDIO_SERVICE) as AudioManager
        initMediaSession()
    }

    override fun onBind(intent: Intent?): IBinder = binder

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        Log.d(TAG, "onStartCommand(), action=${intent?.action}")
        when (intent?.action) {
            Actions.ACTION_PLAY -> resume()
            Actions.ACTION_PAUSE -> pause()
            Actions.ACTION_STOP -> {
                abandonAudioFocus()
                stop()
                stopForeground(STOP_FOREGROUND_REMOVE)
                stopSelf()
            }
            Actions.ACTION_PREV -> callback?.onPrevChapter()
            Actions.ACTION_NEXT -> callback?.onNextChapter()
        }
        return START_STICKY
    }

    override fun onDestroy() {
        Log.d(TAG, "onDestroy()")
        mainHandler.removeCallbacks(clearTtsRunnable)
        abandonAudioFocus()
        clearTts()
        mediaSessionCompat?.release()
        mediaSessionCompat = null
        super.onDestroy()
    }

    @Suppress("DEPRECATION")
    private fun requestAudioFocus(): Boolean {
        if (audioFocusGranted) return true
        val result = audioManager?.requestAudioFocus(
            AudioManager.OnAudioFocusChangeListener { focusChange ->
                when (focusChange) {
                    AudioManager.AUDIOFOCUS_LOSS -> {
                        Log.d(TAG, "Audio focus LOST")
                        pause()
                        audioFocusGranted = false
                    }
                    AudioManager.AUDIOFOCUS_LOSS_TRANSIENT -> pause()
                    AudioManager.AUDIOFOCUS_LOSS_TRANSIENT_CAN_DUCK -> textToSpeech?.setSpeechRate(0.5f)
                    AudioManager.AUDIOFOCUS_GAIN -> {
                        if (!audioFocusGranted) audioFocusGranted = true
                        textToSpeech?.setSpeechRate(speechRate)
                        resume()
                    }
                }
            },
            AudioManager.STREAM_MUSIC,
            AudioManager.AUDIOFOCUS_GAIN
        )
        audioFocusGranted = result == AudioManager.AUDIOFOCUS_REQUEST_GRANTED
        Log.d(TAG, "requestAudioFocus result: $result, granted: $audioFocusGranted")
        return audioFocusGranted
    }

    @Suppress("DEPRECATION")
    private fun abandonAudioFocus() {
        if (audioFocusGranted) {
            audioManager?.abandonAudioFocus(null)
            audioFocusGranted = false
        }
    }

    fun setCallback(callback: TtsEngine.Callback?) {
        this.callback = callback
    }

    fun speak(sentences: List<String>, startIndex: Int, bookTitle: String = "听书", chapterTitle: String = "") {
        Log.d(TAG, "speak() called, sentences.size=${sentences.size}, startIndex=$startIndex, bookTitle=$bookTitle, chapterTitle=$chapterTitle")

        if (onInit) {
            Log.d(TAG, "TTS is initializing, skip")
            return
        }

        this.bookTitle = bookTitle
        this.chapterTitle = chapterTitle
        
        // 填充句子到 Map（索引 -> 句子）
        sentenceMap.clear()
        sentences.forEachIndexed { index, sentence ->
            sentenceMap[index] = sentence
        }
        totalSentences = sentences.size
        Log.d(TAG, "Filled sentenceMap with ${sentenceMap.size} sentences")

        requestAudioFocus()

        if (textToSpeech == null) {
            onInit = true
            Log.d(TAG, "Creating TTS in Service...")
            var showNotification = true
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                val notificationManager = getSystemService(android.content.Context.NOTIFICATION_SERVICE) as android.app.NotificationManager
                if (!notificationManager.areNotificationsEnabled()) {
                    showNotification = false
                }
            }
            Log.d(TAG, "Initializing TTS...")
            textToSpeech = TextToSpeech(this, initListener)
            if (showNotification) {
                startForeground(NOTIFICATION_ID, createNotification())
            }
        } else {
            startPlayback()
        }
    }

    private fun startPlayback() {
        if (sentenceMap.isEmpty()) return

        mainHandler.removeCallbacks(clearTtsRunnable)

        try {
            val result = textToSpeech?.speak("", TextToSpeech.QUEUE_FLUSH, null, null)
            if (result == TextToSpeech.ERROR) {
                Log.e(TAG, "QUEUE_FLUSH failed, reinitializing...")
                clearTts()
                onInit = true
                textToSpeech = TextToSpeech(this, initListener)
                return
            }

            currentSentenceIndex = 0
            isPlaying = true
            updateMediaMetadata()
            updateNotification()
            playCurrentSentence()

        } catch (e: Exception) {
            Log.e(TAG, "startPlayback error", e)
            callback?.onError("播放异常: ${e.message}")
        }
    }

    private fun playNextSentence() {
        val nextIndex = currentSentenceIndex + 1
        if (nextIndex < totalSentences) {
            currentSentenceIndex = nextIndex
            playCurrentSentence()
        } else {
            callback?.onAllComplete()
        }
    }

    private fun playCurrentSentence() {
        val tts = textToSpeech ?: run {
            Log.e(TAG, "playCurrentSentence failed: TTS is null")
            return
        }

        val index = currentSentenceIndex
        val sentence = sentenceMap[index] ?: run {
            Log.e(TAG, "playCurrentSentence failed: sentence not found at index $index")
            return
        }

        Log.d(TAG, "Playing sentence[$index]: ${sentence.take(20)}...")

        val params = android.os.Bundle()
        params.putString(TextToSpeech.Engine.KEY_PARAM_UTTERANCE_ID, "$UTTERANCE_PREFIX$index")

        tts.setSpeechRate(speechRate)
        tts.setPitch(pitch)

        try {
            val result = tts.speak(sentence, TextToSpeech.QUEUE_FLUSH, params, "$UTTERANCE_PREFIX$index")
            if (result == TextToSpeech.ERROR) {
                Log.e(TAG, "speak returned ERROR")
                callback?.onError("播放失败")
            }
        } catch (e: Exception) {
            Log.e(TAG, "speak exception", e)
            callback?.onError("播放异常: ${e.message}")
        }
    }

    fun pause() {
        Log.d(TAG, "pause()")
        mainHandler.removeCallbacks(clearTtsRunnable)
        textToSpeech?.stop()
        isPlaying = false
        updateMediaSessionState(PlaybackState.STATE_PAUSED)
        updateNotification()
    }

    fun resume() {
        Log.d(TAG, "resume()")
        mainHandler.removeCallbacks(clearTtsRunnable)
        isPlaying = true
        updateMediaSessionState(PlaybackState.STATE_PLAYING)
        playCurrentSentence()
    }

    fun stop() {
        Log.d(TAG, "stop()")
        mainHandler.removeCallbacks(clearTtsRunnable)
        textToSpeech?.stop()
        abandonAudioFocus()
        isPlaying = false
        updateMediaSessionState(PlaybackState.STATE_STOPPED)
        updateNotification()
    }

    fun setSpeechRate(rate: Float) {
        speechRate = rate.coerceIn(0.5f, 2.0f)
        textToSpeech?.setSpeechRate(speechRate)
    }

    fun isInitialized(): Boolean = textToSpeech != null

    private fun extractIndex(utteranceId: String?): Int {
        if (utteranceId == null) return -1
        return try {
            utteranceId.replace(UTTERANCE_PREFIX, "").toInt()
        } catch (e: Exception) {
            -1
        }
    }

    private fun clearTts() {
        Log.d(TAG, "clearTts()")
        sentenceMap.clear()
        currentSentenceIndex = -1
        totalSentences = 0
        currentSentence = ""

        textToSpeech?.let { ttsInstance ->
            try {
                ttsInstance.stop()
                ttsInstance.shutdown()
            } catch (e: Exception) {
                Log.w(TAG, "clearTts exception", e)
            }
        }
        textToSpeech = null
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                getString(R.string.tts_notification_channel_name),
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = getString(R.string.tts_notification_channel_desc)
                setShowBadge(false)
            }
            val manager = getSystemService(NotificationManager::class.java)
            manager.createNotificationChannel(channel)
        }
    }

    private fun createNotification(): Notification {
        val intent = Intent(this, ReaderActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_SINGLE_TOP
        }
        val pendingIntent = PendingIntent.getActivity(
            this, 0, intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )

        val prevAction = NotificationCompat.Action(
            R.drawable.ic_skip_previous,
            getString(R.string.tts_notification_action_prev),
            createActionIntent(Actions.ACTION_PREV)
        )

        val playPauseAction = NotificationCompat.Action(
            if (isPlaying) R.drawable.ic_pause else R.drawable.ic_play,
            if (isPlaying) getString(R.string.tts_notification_action_pause)
            else getString(R.string.tts_notification_action_play),
            createActionIntent(if (isPlaying) Actions.ACTION_PAUSE else Actions.ACTION_PLAY)
        )

        val nextAction = NotificationCompat.Action(
            R.drawable.ic_skip_next,
            getString(R.string.tts_notification_action_next),
            createActionIntent(Actions.ACTION_NEXT)
        )

        val builder = NotificationCompat.Builder(this, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_tts_notification)
            .setContentTitle(bookTitle.ifEmpty { "听书" })
            .setContentText(chapterTitle.ifEmpty { "第${currentSentenceIndex + 1}句" })
            .setContentIntent(pendingIntent)
            .setOngoing(true)
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .setCategory(NotificationCompat.CATEGORY_TRANSPORT)
            .setVisibility(NotificationCompat.VISIBILITY_PUBLIC)
            .setForegroundServiceBehavior(NotificationCompat.FOREGROUND_SERVICE_IMMEDIATE)
            .setOnlyAlertOnce(true)
            .addAction(prevAction)
            .addAction(playPauseAction)
            .addAction(nextAction)
            .setStyle(MediaNotificationCompat.MediaStyle()
                .setShowActionsInCompactView(0, 1)
                .setMediaSession(mediaSessionCompat?.sessionToken))

        return builder.build()
    }

    private fun createActionIntent(action: String): PendingIntent {
        val intent = Intent(this, TtsForegroundService::class.java).apply {
            this.action = action
        }
        return PendingIntent.getService(
            this, action.hashCode(), intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
    }

    private fun updateNotification() {
        val notification = createNotification()
        val manager = getSystemService(NotificationManager::class.java)
        manager.notify(NOTIFICATION_ID, notification)
    }

    private fun initMediaSession() {
        mediaSessionCompat = MediaSessionCompat(this, "storytrim_tts").apply {
            setCallback(object : MediaSessionCompat.Callback() {
                override fun onPlay() = resume()
                override fun onPause() = pause()
                override fun onSkipToNext() {
                    callback?.onNextChapter()
                }
                override fun onSkipToPrevious() {
                    callback?.onPrevChapter()
                }
                override fun onStop() {
                    stop()
                    stopForeground(STOP_FOREGROUND_REMOVE)
                    stopSelf()
                }
            })
            setMediaButtonReceiver(createMediaButtonPendingIntent())
            isActive = true
        }
    }

    private fun updateMediaSessionState(state: Int) {
        val stateCompat = PlaybackStateCompat.Builder()
            // 使用直播模式：排除 ACTION_SEEK_TO，告诉系统隐藏进度条
            .setActions(
                PlaybackStateCompat.ACTION_PLAY or
                PlaybackStateCompat.ACTION_PAUSE or
                PlaybackStateCompat.ACTION_SKIP_TO_NEXT or
                PlaybackStateCompat.ACTION_SKIP_TO_PREVIOUS
                // 注意：故意排除 ACTION_SEEK_TO，启用直播模式隐藏进度条
            )
            .setState(state, PlaybackStateCompat.PLAYBACK_POSITION_UNKNOWN, 0f)
            .build()

        mediaSessionCompat?.setPlaybackState(stateCompat)
    }

    private fun updateMediaMetadata() {
        try {
            val metadata = MediaMetadataCompat.Builder()
                .putString(MediaMetadataCompat.METADATA_KEY_TITLE, bookTitle)
                .putString(MediaMetadataCompat.METADATA_KEY_ARTIST, chapterTitle)
                .putString(MediaMetadataCompat.METADATA_KEY_ALBUM, currentSentence)
                .putLong(MediaMetadataCompat.METADATA_KEY_TRACK_NUMBER, currentSentenceIndex.toLong())
                .putLong(MediaMetadataCompat.METADATA_KEY_NUM_TRACKS, totalSentences.toLong())
                .build()
            mediaSessionCompat?.setMetadata(metadata)
        } catch (e: Exception) {
            Log.e(TAG, "updateMediaMetadata error", e)
        }
    }

    fun setChapterTitle(title: String) {
        this.chapterTitle = title
        updateMediaMetadata()
    }

    fun getPlaybackState(): PlaybackInfo {
        return PlaybackInfo(
            isPlaying = isPlaying,
            bookTitle = bookTitle,
            chapterTitle = chapterTitle,
            currentSentence = currentSentence,
            currentIndex = currentSentenceIndex,
            totalCount = totalSentences
        )
    }

    data class PlaybackInfo(
        val isPlaying: Boolean,
        val bookTitle: String,
        val chapterTitle: String,
        val currentSentence: String,
        val currentIndex: Int,
        val totalCount: Int
    )

    private fun createMediaButtonPendingIntent(): PendingIntent {
        val intent = Intent(Intent.ACTION_MEDIA_BUTTON)
        return PendingIntent.getBroadcast(
            this, 0, intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
    }

    object Actions {
        const val ACTION_PLAY = "com.storytrim.tts.PLAY"
        const val ACTION_PAUSE = "com.storytrim.tts.PAUSE"
        const val ACTION_STOP = "com.storytrim.tts.STOP"
        const val ACTION_PREV = "com.storytrim.tts.PREV"
        const val ACTION_NEXT = "com.storytrim.tts.NEXT"
    }
}
