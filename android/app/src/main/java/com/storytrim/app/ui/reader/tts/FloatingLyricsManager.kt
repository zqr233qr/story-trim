package com.storytrim.app.ui.reader.tts

import android.app.Activity
import android.content.Context
import android.content.Intent
import android.net.Uri
import android.os.Build
import android.provider.Settings
import android.util.Log
import androidx.activity.result.ActivityResultLauncher

/**
 * 悬浮窗歌词管理器
 * 负责检查权限、启动/关闭悬浮窗
 */
class FloatingLyricsManager(
    private val context: Context
) {
    companion object {
        private const val TAG = "FloatingLyricsManager"
        private const val OVERLAY_PERMISSION_REQUEST_CODE = 1001
    }

    private var permissionLauncher: ActivityResultLauncher<Intent>? = null

    /**
     * 检查是否有悬浮窗权限
     */
    fun canDrawOverlays(): Boolean {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            Settings.canDrawOverlays(context)
        } else {
            true
        }
    }

    /**
     * 请求悬浮窗权限
     */
    fun requestOverlayPermission(activity: Activity) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            val intent = Intent(
                Settings.ACTION_MANAGE_OVERLAY_PERMISSION,
                Uri.parse("package:${activity.packageName}")
            )
            activity.startActivity(intent)
        }
    }

    /**
     * 检查并请求权限，返回是否已授权
     */
    fun checkAndRequestPermission(activity: Activity): Boolean {
        return if (canDrawOverlays()) {
            true
        } else {
            requestOverlayPermission(activity)
            false
        }
    }

    /**
     * 显示悬浮窗
     */
    fun show() {
        Log.d(TAG, "show()")

        if (!canDrawOverlays()) {
            Log.w(TAG, "No overlay permission")
            return
        }

        val intent = Intent(context, FloatingLyricsService::class.java).apply {
            action = FloatingLyricsService.ACTION_SHOW
        }

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            context.startForegroundService(intent)
        } else {
            context.startService(intent)
        }
    }

    /**
     * 隐藏悬浮窗
     */
    fun hide() {
        Log.d(TAG, "hide()")

        val intent = Intent(context, FloatingLyricsService::class.java).apply {
            action = FloatingLyricsService.ACTION_HIDE
        }
        context.startService(intent)
    }

    /**
     * 关闭悬浮窗
     */
    fun close() {
        Log.d(TAG, "close()")

        val intent = Intent(context, FloatingLyricsService::class.java).apply {
            action = FloatingLyricsService.ACTION_CLOSE
        }
        context.startService(intent)
    }

    /**
     * 更新歌词
     */
    fun updateLyrics(sentence: String, isPlaying: Boolean) {
        if (!FloatingLyricsService.isShowing()) return

        val intent = Intent(context, FloatingLyricsService::class.java).apply {
            action = FloatingLyricsService.ACTION_UPDATE
            putExtra(FloatingLyricsService.EXTRA_SENTENCE, sentence)
            putExtra(FloatingLyricsService.EXTRA_IS_PLAYING, isPlaying)
        }
        context.startService(intent)
    }

    /**
     * 切换播放/暂停
     */
    fun togglePlayPause() {
        val intent = Intent(context, FloatingLyricsService::class.java).apply {
            action = FloatingLyricsService.ACTION_TOGGLE_PLAY_PAUSE
        }
        context.startService(intent)
    }

    /**
     * 悬浮窗是否正在显示
     */
    fun isShowing(): Boolean = FloatingLyricsService.isShowing()

    /**
     * 设置权限请求回调
     */
    fun setPermissionLauncher(launcher: ActivityResultLauncher<Intent>) {
        this.permissionLauncher = launcher
    }
}
