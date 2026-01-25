package com.storytrim.app.ui.reader.tts

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.graphics.PixelFormat
import android.os.Build
import android.os.IBinder
import android.util.Log
import android.view.Gravity
import android.view.WindowManager
import androidx.core.app.NotificationCompat
import com.storytrim.app.R
import com.storytrim.app.ui.reader.ReaderActivity

/**
 * 悬浮窗歌词服务
 * 在屏幕顶层显示歌词，支持拖动、缩放、暂停控制
 */
class FloatingLyricsService : Service() {

    companion object {
        private const val TAG = "FloatingLyricsService"
        private const val CHANNEL_ID = "floating_lyrics_channel"
        private const val NOTIFICATION_ID = 1002

        const val ACTION_SHOW = "com.storytrim.floating_lyrics.SHOW"
        const val ACTION_HIDE = "com.storytrim.floating_lyrics.HIDE"
        const val ACTION_UPDATE = "com.storytrim.floating_lyrics.UPDATE"
        const val ACTION_TOGGLE_PLAY_PAUSE = "com.storytrim.floating_lyrics.TOGGLE_PLAY_PAUSE"
        const val ACTION_CLOSE = "com.storytrim.floating_lyrics.CLOSE"

        const val EXTRA_SENTENCE = "sentence"
        const val EXTRA_IS_PLAYING = "is_playing"

        private var instance: FloatingLyricsService? = null

        fun getInstance(): FloatingLyricsService? = instance

        fun isShowing(): Boolean = instance != null
    }

    private var windowManager: WindowManager? = null
    private var floatingView: FloatingLyricsView? = null
    private var currentSentence: String = ""
    private var currentIsPlaying: Boolean = false

    override fun onCreate() {
        super.onCreate()
        Log.d(TAG, "onCreate()")
        instance = this

        windowManager = getSystemService(WINDOW_SERVICE) as WindowManager

        createNotificationChannel()
        startForeground(NOTIFICATION_ID, createNotification())
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        Log.d(TAG, "onStartCommand: ${intent?.action}")

        when (intent?.action) {
            ACTION_SHOW -> showFloatingWindow()
            ACTION_HIDE -> hideFloatingWindow()
            ACTION_UPDATE -> updateLyrics(
                intent.getStringExtra(EXTRA_SENTENCE) ?: "",
                intent.getBooleanExtra(EXTRA_IS_PLAYING, false)
            )
            ACTION_TOGGLE_PLAY_PAUSE -> {
                // 发送广播通知 TtsController
                sendBroadcast(Intent(TtsController.ACTION_PLAY_PAUSE))
            }
            ACTION_CLOSE -> {
                hideFloatingWindow()
                stopSelf()
            }
        }

        return START_STICKY
    }

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onDestroy() {
        Log.d(TAG, "onDestroy()")
        instance = null
        hideFloatingWindow()
        super.onDestroy()
    }

    /**
     * 显示悬浮窗
     */
    private fun showFloatingWindow() {
        if (floatingView != null) return

        Log.d(TAG, "showFloatingWindow()")

        // 创建悬浮窗视图
        floatingView = FloatingLyricsView(this).apply {
            setPlayPauseCallback {
                sendBroadcast(Intent(TtsController.ACTION_PLAY_PAUSE))
            }
            setCloseCallback {
                hideFloatingWindow()
                stopSelf()
            }
        }

        // 设置窗口参数
        val params = createLayoutParams()
        floatingView?.saveWindowLayoutParams(params)

        try {
            windowManager?.addView(floatingView, params)
            Log.d(TAG, "Floating window added")

            // 初始显示当前内容
            if (currentSentence.isNotEmpty()) {
                floatingView?.updateLyrics(currentSentence, currentIsPlaying)
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to show floating window", e)
        }
    }

    /**
     * 隐藏悬浮窗
     */
    private fun hideFloatingWindow() {
        floatingView?.let { view ->
            try {
                windowManager?.removeView(view)
            } catch (e: Exception) {
                Log.w(TAG, "Failed to remove floating window", e)
            }
        }
        floatingView = null
    }

    /**
     * 更新歌词内容
     */
    private fun updateLyrics(sentence: String, isPlaying: Boolean) {
        currentSentence = sentence
        currentIsPlaying = isPlaying

        floatingView?.updateLyrics(sentence, isPlaying)
    }

    /**
     * 创建窗口布局参数
     */
    private fun createLayoutParams(): WindowManager.LayoutParams {
        return WindowManager.LayoutParams().apply {
            type = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY
            } else {
                @Suppress("DEPRECATION")
                WindowManager.LayoutParams.TYPE_PHONE
            }

            format = PixelFormat.TRANSLUCENT
            flags = WindowManager.LayoutParams.FLAG_NOT_FOCUSABLE or
                    WindowManager.LayoutParams.FLAG_LAYOUT_IN_SCREEN or
                    WindowManager.LayoutParams.FLAG_LAYOUT_NO_LIMITS

            width = WindowManager.LayoutParams.WRAP_CONTENT
            height = WindowManager.LayoutParams.WRAP_CONTENT

            gravity = Gravity.CENTER

            x = 0
            y = 0
        }
    }

    /**
     * 创建通知渠道
     */
    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                getString(R.string.floating_lyrics_channel_name),
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = getString(R.string.floating_lyrics_channel_desc)
                setShowBadge(false)
            }

            val manager = getSystemService(NotificationManager::class.java)
            manager.createNotificationChannel(channel)
        }
    }

    /**
     * 创建通知
     */
    private fun createNotification(): Notification {
        val intent = Intent(this, ReaderActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_SINGLE_TOP
        }

        val pendingIntent = PendingIntent.getActivity(
            this, 0, intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )

        return NotificationCompat.Builder(this, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_tts_notification)
            .setContentTitle(getString(R.string.floating_lyrics_title))
            .setContentText(currentSentence.ifEmpty { "桌面歌词已开启" })
            .setContentIntent(pendingIntent)
            .setOngoing(true)
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .setCategory(NotificationCompat.CATEGORY_SERVICE)
            .build()
    }

    /**
     * 更新通知
     */
    private fun updateNotification() {
        val notification = createNotification()
        val manager = getSystemService(NotificationManager::class.java)
        manager.notify(NOTIFICATION_ID, notification)
    }
}
