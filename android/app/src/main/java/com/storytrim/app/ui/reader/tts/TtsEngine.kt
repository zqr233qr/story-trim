package com.storytrim.app.ui.reader.tts

/**
 * TTS引擎接口
 * 预留多引擎架构，支持扩展不同的TTS实现
 */
interface TtsEngine {

    /**
     * 播放状态
     */
    enum class State {
        IDLE,      // 空闲
        PLAYING,   // 正在播放
        PAUSED,    // 已暂停
        STOPPED    // 已停止
    }

    /**
     * 播放回调
     * 使用索引+句子参数，确保朗读与显示一致
     */
    interface Callback {
        fun onStart(sentenceIndex: Int, sentence: String)
        fun onComplete(sentenceIndex: Int, sentence: String)
        fun onError(error: String)
        fun onAllComplete()
        fun onPrevChapter()
        fun onNextChapter()
    }

    /**
     * 初始化引擎
     * @return 是否初始化成功
     */
    fun init(): Boolean

    /**
     * 播放句子列表
     * @param sentences 句子列表
     * @param startIndex 起始句子索引
     * @param title 书籍标题（用于通知栏显示）
     */
    fun speak(sentences: List<String>, startIndex: Int, title: String = "听书")

    /**
     * 暂停播放
     */
    fun pause()

    /**
     * 恢复播放
     */
    fun resume()

    /**
     * 停止播放
     */
    fun stop()

    /**
     * 设置语速
     * @param rate 语速倍率 (0.5 - 2.0)
     */
    fun setSpeechRate(rate: Float)

    /**
     * 设置音调
     * @param pitch 音调 (0.5 - 2.0)
     */
    fun setPitch(pitch: Float)

    /**
     * 获取当前播放状态
     */
    fun getState(): State

    /**
     * 获取当前播放的句子索引
     */
    fun getCurrentIndex(): Int

    /**
     * 设置回调
     */
    fun setCallback(callback: Callback?)

    /**
     * 释放资源
     */
    fun release()

    /**
     * 检查引擎是否可用
     */
    fun isAvailable(): Boolean

    /**
     * 获取引擎名称
     */
    fun getEngineName(): String

    /**
     * 获取支持的语言列表
     */
    fun getSupportedLanguages(): List<String>
}
