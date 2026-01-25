package com.storytrim.app.ui.reader.tts

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.os.Build
import androidx.core.app.NotificationCompat
import com.storytrim.app.R
import com.storytrim.app.ui.reader.ReaderActivity

/**
 * 听书通知栏帮助类
 */
class TtsNotificationHelper(private val context: Context) {

    companion object {
        private const val CHANNEL_ID = "tts_playback_channel"
        private const val NOTIFICATION_ID = 1001
    }

    private val notificationManager: NotificationManager =
        context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager

    init {
        createNotificationChannel()
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                context.getString(R.string.tts_notification_channel_name),
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = context.getString(R.string.tts_notification_channel_desc)
                setShowBadge(false)
            }
            notificationManager.createNotificationChannel(channel)
        }
    }

    /**
     * 创建听书通知
     * @param title 书籍/章节标题
     * @param isPlaying 当前是否正在播放
     * @param currentSentence 当前播放的句子
     */
    fun createNotification(
        title: String,
        isPlaying: Boolean,
        currentSentence: String = ""
    ): Notification {
        val intent = Intent(context, ReaderActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_SINGLE_TOP
        }
        val pendingIntent = PendingIntent.getActivity(
            context,
            0,
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )

        val contentText = if (currentSentence.isNotEmpty()) {
            context.getString(R.string.tts_notification_content, currentSentence.take(50))
        } else {
            context.getString(R.string.tts_notification_content, title)
        }

        return NotificationCompat.Builder(context, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_tts_notification)
            .setContentTitle(context.getString(R.string.tts_notification_title))
            .setContentText(contentText)
            .setContentIntent(pendingIntent)
            .setOngoing(isPlaying)
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .setCategory(NotificationCompat.CATEGORY_SERVICE)
            .setVisibility(NotificationCompat.VISIBILITY_PUBLIC)
            .addAction(
                R.drawable.ic_play,
                context.getString(R.string.tts_notification_action_play),
                createActionPendingIntent(Actions.ACTION_PLAY)
            )
            .addAction(
                R.drawable.ic_pause,
                context.getString(R.string.tts_notification_action_pause),
                createActionPendingIntent(Actions.ACTION_PAUSE)
            )
            .addAction(
                R.drawable.ic_stop,
                context.getString(R.string.tts_notification_action_stop),
                createActionPendingIntent(Actions.ACTION_STOP)
            )
            .build()
    }

    private fun createActionPendingIntent(action: String): PendingIntent {
        val intent = Intent(context, TtsForegroundService::class.java).apply {
            this.action = action
        }
        return PendingIntent.getService(
            context,
            action.hashCode(),
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
    }

    fun showNotification(notification: Notification) {
        notificationManager.notify(NOTIFICATION_ID, notification)
    }

    fun cancelNotification() {
        notificationManager.cancel(NOTIFICATION_ID)
    }

    object Actions {
        const val ACTION_PLAY = "com.storytrim.tts.PLAY"
        const val ACTION_PAUSE = "com.storytrim.tts.PAUSE"
        const val ACTION_STOP = "com.storytrim.tts.STOP"
    }
}
