package com.storytrim.app.ui.reader.tts

import android.content.Context
import android.os.Handler
import android.os.Looper
import android.speech.tts.TextToSpeech
import android.speech.tts.UtteranceProgressListener
import android.util.Log
import java.util.Locale
import java.util.Queue
import java.util.concurrent.ConcurrentLinkedQueue

class SystemTtsEngine(
    private val context: Context
) : TtsEngine {

    companion object {
        private const val TAG = "SystemTtsEngine"
        private const val UTTERANCE_PREFIX = "tts_sentence_"
        private const val CLEAR_TTS_DELAY = 60000L // 1分钟无活动自动释放
    }

    private val mainHandler = Handler(Looper.getMainLooper())
    private val clearTtsRunnable = Runnable { clearTts() }

    private var tts: TextToSpeech? = null
    private var callback: TtsEngine.Callback? = null
    private var text: String? = null
    private var onInit = false
    private var pendingSentences: List<String>? = null
    private var pendingStartIndex = 0
    private var sentenceQueue: Queue<String>? = null
    private var currentSentenceIndex = 0

    private var speechRate = 1.0f
    private var pitch = 1.0f

    private val initListener = TextToSpeech.OnInitListener { status ->
        if (status == TextToSpeech.SUCCESS) {
            Log.d(TAG, "TTS init SUCCESS")
            tts?.setOnUtteranceProgressListener(utteranceListener)
            onInit = false
            // 初始化完成后继续播放
            addTextToSpeakList()
        } else {
            Log.e(TAG, "TTS init FAILED with status: $status")
            onInit = false
            callback?.onError("TTS引擎初始化失败($status)")
        }
    }

    private val utteranceListener = object : UtteranceProgressListener() {
        override fun onStart(utteranceId: String?) {
            Log.d(TAG, "onStart: $utteranceId")
            mainHandler.removeCallbacks(clearTtsRunnable)
            try {
                val index = extractIndex(utteranceId)
                val queue = sentenceQueue
                val sentence = if (queue != null && index >= 0) {
                    queue.peek() ?: ""
                } else ""
                callback?.onStart(index, sentence)
            } catch (e: Exception) {
                Log.e(TAG, "onStart error", e)
            }
        }

        override fun onDone(utteranceId: String?) {
            Log.d(TAG, "onDone: $utteranceId")
            mainHandler.postDelayed(clearTtsRunnable, CLEAR_TTS_DELAY)
            try {
                val index = extractIndex(utteranceId)
                val queue = sentenceQueue
                val sentence = if (queue != null && index >= 0) {
                    queue.poll() ?: ""
                } else ""
                callback?.onComplete(index, sentence)
                playNextSentence()
            } catch (e: Exception) {
                Log.e(TAG, "onDone error", e)
            }
        }

        @Deprecated("Deprecated in Java")
        override fun onError(utteranceId: String?) {
            Log.e(TAG, "onError deprecated: $utteranceId")
            try {
                val index = extractIndex(utteranceId)
                callback?.onError("播放失败: $utteranceId")
                playNextSentence()
            } catch (e: Exception) {
                Log.e(TAG, "onError error", e)
            }
        }

        override fun onError(utteranceId: String?, errorCode: Int) {
            Log.e(TAG, "onError with code: $errorCode, utteranceId: $utteranceId")
            try {
                callback?.onError("播放错误: $errorCode")
                playNextSentence()
            } catch (e: Exception) {
                Log.e(TAG, "onError with code error", e)
            }
        }
    }

    private fun extractIndex(utteranceId: String?): Int {
        if (utteranceId == null) return -1
        return try {
            utteranceId.replace(UTTERANCE_PREFIX, "").toInt()
        } catch (e: Exception) {
            -1
        }
    }

    private fun playNextSentence() {
        val queue = sentenceQueue ?: return
        val nextSentence = queue.poll()
        if (nextSentence != null) {
            speakSentence(nextSentence, currentSentenceIndex)
            currentSentenceIndex++
        } else {
            callback?.onAllComplete()
        }
    }

    private fun speakSentence(sentence: String, index: Int) {
        if (tts == null) {
            Log.e(TAG, "speakSentence failed: TTS is null")
            return
        }

        Log.d(TAG, "speakSentence: index=$index")

        val params = android.os.Bundle()
        params.putString(TextToSpeech.Engine.KEY_PARAM_UTTERANCE_ID, "$UTTERANCE_PREFIX$index")

        tts?.setSpeechRate(speechRate)
        tts?.setPitch(pitch)

        try {
            val result = tts?.speak(sentence, TextToSpeech.QUEUE_FLUSH, params, "$UTTERANCE_PREFIX$index")
            if (result == TextToSpeech.ERROR) {
                Log.e(TAG, "speak returned ERROR")
                callback?.onError("播放失败")
            }
        } catch (e: Exception) {
            Log.e(TAG, "speak exception", e)
            callback?.onError("播放异常: ${e.message}")
        }
    }

    private fun addTextToSpeakList() {
        val sentences = pendingSentences ?: return

        mainHandler.removeCallbacks(clearTtsRunnable)

        // 清空当前队列
        val currentTts = tts
        try {
            if (currentTts == null) {
                // TTS未初始化，重新初始化
                onInit = true
                tts = TextToSpeech(context.applicationContext, initListener)
                return
            }

            var result = currentTts.speak("", TextToSpeech.QUEUE_FLUSH, null, null)
            if (result == TextToSpeech.ERROR) {
                Log.e(TAG, "QUEUE_FLUSH failed, reinitializing...")
                clearTts()
                onInit = true
                tts = TextToSpeech(context.applicationContext, initListener)
                return
            }

            // 创建队列并跳过前面的句子
            sentenceQueue = ConcurrentLinkedQueue(sentences)
            repeat(pendingStartIndex) { sentenceQueue?.poll() }
            currentSentenceIndex = pendingStartIndex - 1

            // 播放第一句
            playNextSentence()
        } catch (e: Exception) {
            Log.e(TAG, "addTextToSpeakList error", e)
            callback?.onError("播放异常: ${e.message}")
        }

        pendingSentences = null
    }

    private fun clearTts() {
        Log.d(TAG, "clearTts() called")
        tts?.let { ttsInstance ->
            try {
                ttsInstance.stop()
                ttsInstance.shutdown()
            } catch (e: Exception) {
                Log.w(TAG, "clearTts exception", e)
            }
        }
        tts = null
        sentenceQueue = null
    }

    override fun init(): Boolean = tts != null

    override fun speak(sentences: List<String>, startIndex: Int, title: String) {
        Log.d(TAG, "speak() called, onInit=$onInit, tts=${tts != null}")

        if (onInit) {
            Log.d(TAG, "TTS is initializing, skip")
            return
        }

        text = sentences.joinToString("\n")
        pendingSentences = sentences
        pendingStartIndex = startIndex

        if (tts == null) {
            onInit = true
            Log.d(TAG, "Creating TTS with applicationContext...")
            tts = TextToSpeech(context.applicationContext, initListener)
        } else {
            addTextToSpeakList()
        }
    }

    override fun pause() {
        tts?.stop()
    }

    override fun resume() {
        addTextToSpeakList()
    }

    override fun stop() {
        mainHandler.removeCallbacks(clearTtsRunnable)
        tts?.stop()
    }

    override fun setSpeechRate(rate: Float) {
        speechRate = rate.coerceIn(0.5f, 2.0f)
        tts?.setSpeechRate(speechRate)
    }

    override fun setPitch(pitch: Float) {
        this.pitch = pitch.coerceIn(0.5f, 2.0f)
        tts?.setPitch(this.pitch)
    }

    override fun getState(): TtsEngine.State {
        return when {
            tts == null -> TtsEngine.State.IDLE
            tts?.isSpeaking == true -> TtsEngine.State.PLAYING
            else -> TtsEngine.State.STOPPED
        }
    }

    override fun getCurrentIndex(): Int = currentSentenceIndex

    override fun setCallback(callback: TtsEngine.Callback?) {
        this.callback = callback
    }

    override fun release() {
        Log.d(TAG, "release() called")
        mainHandler.removeCallbacks(clearTtsRunnable)
        clearTts()
        callback = null
    }

    override fun isAvailable(): Boolean = tts != null

    override fun getEngineName(): String = "系统TTS"

    override fun getSupportedLanguages(): List<String> {
        return listOf(
            Locale.CHINESE.displayName,
            Locale.ENGLISH.displayName,
            Locale.getDefault().displayName
        ).distinct()
    }
}
