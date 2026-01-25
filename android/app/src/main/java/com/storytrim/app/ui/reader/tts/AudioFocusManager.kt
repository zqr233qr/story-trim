package com.storytrim.app.ui.reader.tts

import android.content.Context
import android.media.AudioAttributes
import android.media.AudioManager
import android.util.Log

/**
 * 音频焦点管理器
 * 负责请求和释放音频焦点，支持被打断后恢复
 */
class AudioFocusManager(
    private val context: Context
) {
    companion object {
        private const val TAG = "AudioFocusManager"
    }

    private val audioManager: AudioManager by lazy {
        context.getSystemService(Context.AUDIO_SERVICE) as AudioManager
    }

    private var focusChangeListener: AudioManager.OnAudioFocusChangeListener? = null
    private var isFocusHeld = false
    private var wasPlayingBeforeFocusLoss = false

    /**
     * 焦点状态
     */
    enum class FocusState {
        FOCUSED,      // 获得焦点
        LOST,         // 失去焦点
        LOST_TRANSIENT,  // 暂时失去焦点（可恢复）
        LOST_TRANSIENT_CAN_DUCK  // 暂时失去焦点，可降低音量
    }

    /**
     * 焦点回调
     */
    interface FocusCallback {
        fun onFocusChange(state: FocusState)
    }

    private var focusCallback: FocusCallback? = null

    /**
     * 设置焦点回调
     */
    fun setFocusCallback(callback: FocusCallback?) {
        this.focusCallback = callback
    }

    /**
     * 请求音频焦点
     * @return 是否成功获得焦点
     */
    @Suppress("DEPRECATION")
    fun requestFocus(): Boolean {
        if (isFocusHeld) return true

        focusChangeListener = AudioManager.OnAudioFocusChangeListener { focusChange ->
            Log.d(TAG, "Audio focus changed: $focusChange")
            when (focusChange) {
                AudioManager.AUDIOFOCUS_GAIN -> {
                    isFocusHeld = true
                    focusCallback?.onFocusChange(FocusState.FOCUSED)
                    audioManager.adjustStreamVolume(
                        AudioManager.STREAM_MUSIC,
                        AudioManager.ADJUST_UNMUTE,
                        0
                    )
                }
                AudioManager.AUDIOFOCUS_LOSS -> {
                    isFocusHeld = false
                    focusCallback?.onFocusChange(FocusState.LOST)
                }
                AudioManager.AUDIOFOCUS_LOSS_TRANSIENT -> {
                    wasPlayingBeforeFocusLoss = isFocusHeld
                    focusCallback?.onFocusChange(FocusState.LOST_TRANSIENT)
                }
                AudioManager.AUDIOFOCUS_LOSS_TRANSIENT_CAN_DUCK -> {
                    wasPlayingBeforeFocusLoss = isFocusHeld
                    audioManager.adjustStreamVolume(
                        AudioManager.STREAM_MUSIC,
                        AudioManager.ADJUST_LOWER,
                        0
                    )
                    focusCallback?.onFocusChange(FocusState.LOST_TRANSIENT_CAN_DUCK)
                }
                AudioManager.AUDIOFOCUS_GAIN_TRANSIENT -> {
                    isFocusHeld = true
                    focusCallback?.onFocusChange(FocusState.FOCUSED)
                }
            }
        }

        return try {
            val result = audioManager.requestAudioFocus(
                focusChangeListener,
                AudioManager.STREAM_MUSIC,
                AudioManager.AUDIOFOCUS_GAIN
            )
            result == AudioManager.AUDIOFOCUS_REQUEST_GRANTED
        } catch (e: SecurityException) {
            Log.e(TAG, "requestAudioFocus failed", e)
            false
        }
    }

    /**
     * 释放音频焦点
     */
    fun abandonFocus() {
        focusChangeListener?.let { listener ->
            try {
                audioManager.abandonAudioFocus(listener)
            } catch (e: SecurityException) {
                Log.e(TAG, "abandonAudioFocus failed", e)
            }
        }
        focusChangeListener = null
        isFocusHeld = false
        wasPlayingBeforeFocusLoss = false
    }

    /**
     * 是否持有焦点
     */
    fun hasFocus(): Boolean = isFocusHeld

    /**
     * 获取焦点状态
     */
    fun getFocusState(): FocusState {
        return when {
            !isFocusHeld -> FocusState.LOST
            else -> FocusState.FOCUSED
        }
    }

    /**
     * 获取之前是否在播放
     */
    fun wasPlayingBeforeFocusLoss(): Boolean = wasPlayingBeforeFocusLoss
}
