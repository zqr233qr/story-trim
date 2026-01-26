package com.storytrim.app.ui.reader.core

import android.content.Context
import android.speech.tts.TextToSpeech
import android.speech.tts.UtteranceProgressListener
import android.util.Log
import java.util.Locale

class TtsManager(context: Context) {

    private var tts: TextToSpeech? = null
    private var isInitialized = false
    
    var onPageFinish: (() -> Unit)? = null
    
    init {
        tts = TextToSpeech(context) { status ->
            if (status == TextToSpeech.SUCCESS) {
                tts?.language = Locale.getDefault()
                tts?.setOnUtteranceProgressListener(object : UtteranceProgressListener() {
                    override fun onStart(utteranceId: String?) {}
                    override fun onDone(utteranceId: String?) {
                        onPageFinish?.invoke()
                    }
                    @Deprecated("Deprecated in Java")
                    override fun onError(utteranceId: String?) {}
                })
                isInitialized = true
            } else {
                Log.e("TTS", "Initialization failed")
            }
        }
    }

    fun speak(text: String) {
        if (!isInitialized) return
        val params = android.os.Bundle()
        params.putString(TextToSpeech.Engine.KEY_PARAM_UTTERANCE_ID, "page_utterance")
        tts?.speak(text, TextToSpeech.QUEUE_FLUSH, params, "page_utterance")
    }

    fun stop() {
        tts?.stop()
    }

    fun shutdown() {
        tts?.shutdown()
    }
    
    fun isSpeaking(): Boolean {
        return tts?.isSpeaking ?: false
    }
}
