package com.storytrim.app.ui.reader.tts

/**
 * 句子切分器
 * 使用正则表达式进行中文句子边界识别，支持精确到句子的TTS朗读
 */
class SentenceSegmenter {

    companion object {
        private val SENTENCE_END_PUNCTUATION = setOf(
            '。', '！', '？', '】', '）', '」', '』', '〉', '》', '｝', '．', '.', '!', '?'
        )

        private val PARENTHESIS_PAIRS = mapOf(
            '（' to '）', '(' to ')', '【' to '】', '[' to ']',
            '「' to '」', '『' to '』', '〈' to '〉', '《' to '》'
        )

        private const val MIN_SENTENCE_LENGTH = 2
        private const val MAX_SENTENCE_LENGTH = 500
    }

    /**
     * 将文本切分为句子列表
     * @param text 原始文本
     * @return 句子列表
     */
    fun segment(text: String): List<String> {
        if (text.isBlank()) return emptyList()

        val sentences = mutableListOf<String>()
        val sb = StringBuilder()
        var i = 0

        while (i < text.length) {
            val char = text[i]
            sb.append(char)

            // 检查是否是句末标点
            if (char in SENTENCE_END_PUNCTUATION) {
                // 检查是否是省略号（连续的点）
                if (char == '.' || char == '．') {
                    val prevIndex = i - 1
                    if (prevIndex >= 0 && text[prevIndex] == char) {
                        // 可能是省略号，继续收集
                        i++
                        continue
                    }
                }

                // 检查是否是引号/括号闭合
                if (char == '」' || char == '』' || char == ')' || char == '）' ||
                    char == ']' || char == '】' || char == '》' || char == '｝') {
                    // 检查括号匹配，可能需要继续收集
                    val matchingOpen = PARENTHESIS_PAIRS[char]
                    if (matchingOpen != null) {
                        // 查找开括号
                        var openCount = 1
                        var searchIdx = i - 1
                        while (searchIdx >= 0 && openCount > 0) {
                            if (text[searchIdx] == char) openCount++
                            else if (text[searchIdx] == matchingOpen) openCount--
                            searchIdx--
                        }
                        if (openCount > 0) {
                            // 没有匹配的开括号，继续收集
                            i++
                            continue
                        }
                    }
                }

                val sentence = sb.toString().trim()
                if (isValidSentence(sentence)) {
                    sentences.add(sentence)
                }
                sb.clear()
            } else if (char == '\n') {
                // 换行也作为句子分隔符
                val sentence = sb.toString().trim()
                if (isValidSentence(sentence)) {
                    sentences.add(sentence)
                }
                sb.clear()
            }

            i++
        }

        // 处理最后一段
        val lastSentence = sb.toString().trim()
        if (isValidSentence(lastSentence)) {
            sentences.add(lastSentence)
        }

        return sentences
    }

    /**
     * 验证句子是否有效
     */
    private fun isValidSentence(sentence: String): Boolean {
        val length = sentence.length
        // 过滤单字符句子
        if (length < MIN_SENTENCE_LENGTH) return false
        // 过滤过长的句子
        if (length > MAX_SENTENCE_LENGTH) return false
        // 过滤纯标点
        if (sentence.all { it in SENTENCE_END_PUNCTUATION }) return false
        return true
    }

    /**
     * 计算句子在原文中的位置
     * @param sentences 句子列表
     * @param originalText 原文
     * @return 每个句子在原文中的起始位置列表
     */
    fun calculatePositions(sentences: List<String>, originalText: String): List<Int> {
        val positions = mutableListOf<Int>()
        var searchStart = 0

        for (sentence in sentences) {
            val index = originalText.indexOf(sentence, searchStart)
            if (index >= 0) {
                positions.add(index)
                searchStart = index + sentence.length
            } else {
                positions.add(-1)
            }
        }

        return positions
    }
}
