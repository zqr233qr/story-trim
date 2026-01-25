package com.storytrim.app.core.parser

import java.io.BufferedReader
import java.io.InputStream
import java.io.InputStreamReader
import java.nio.charset.Charset
import java.security.MessageDigest
import java.util.regex.Pattern
import javax.inject.Inject
import javax.inject.Singleton
import kotlin.math.pow
import kotlin.math.sqrt

data class ParsedChapter(
    val index: Int,
    val title: String,
    val content: String,
    val md5: String,
    val wordsCount: Int
)

data class ParseResult(
    val title: String,
    val bookMd5: String,
    val chapters: List<ParsedChapter>
)

data class ParserRule(
    val name: String,
    val pattern: String,
    val weight: Double
)

@Singleton
class FileParser @Inject constructor() {

    private val rules = listOf(
        ParserRule("Strict_Chinese", """(?:^|\n)第[0-9零一二三四五六七八九十百千万]+[章回节][ \t\f].*""", 100.0),
        ParserRule("Normal_Chinese", """(?:^|\n)第[0-9零一二三四五六七八九十百千万]+[章回节].*""", 90.0),
        ParserRule("Strict_English", """(?:^|\n)Chapter\s+\d+.*""", 80.0),
        ParserRule("Loose_Number", """(?:^|\n)\d+\.\s+.*""", 60.0),
        ParserRule("Loose_Direct", """(?:^|\n)[0-9零一二三四五六七八九十百千万]+\s+.*""", 40.0)
    )

    fun parse(inputStream: InputStream, fileName: String): ParseResult {
        val content = inputStream.bufferedReader().use { it.readText() }
        val bookMd5 = calculateMD5(content)
        val bestMatch = findBestRule(content)
        val chapters = extractChapters(content, bestMatch)
        
        if (chapters.isEmpty()) {
            val title = fileName.substringBeforeLast(".")
            return ParseResult(
                title = title,
                bookMd5 = bookMd5,
                chapters = listOf(
                    ParsedChapter(
                        index = 0,
                        title = title,
                        content = content.trim(),
                        md5 = calculateMD5(content.trim()),
                        wordsCount = content.length
                    )
                )
            )
        }

        return ParseResult(
            title = fileName.substringBeforeLast("."),
            bookMd5 = bookMd5,
            chapters = chapters
        )
    }

    private fun findBestRule(content: String): List<Int> {
        var bestIndices: List<Int> = emptyList()
        var maxScore = Double.NEGATIVE_INFINITY
        val totalLen = content.length

        for (rule in rules) {
            val indices = mutableListOf<Int>()
            val matcher = Pattern.compile(rule.pattern).matcher(content)
            while (matcher.find()) {
                indices.add(matcher.start())
            }

            if (indices.isNotEmpty()) {
                val score = calculateScore(totalLen, indices, rule.weight)
                if (score > maxScore) {
                    maxScore = score
                    bestIndices = indices
                }
            }
        }
        return bestIndices
    }

    private fun calculateScore(totalLen: Int, indices: List<Int>, weight: Double): Double {
        val count = indices.size
        if (count == 0) return -10000.0

        val lengths = mutableListOf<Int>()
        for (i in 0 until count) {
            val start = indices[i]
            val next = if (i == count - 1) totalLen else indices[i + 1]
            lengths.add(next - start)
        }

        val sum = lengths.sum()
        val avg = sum.toDouble() / count
        if (avg < 200) return -20000.0

        val varianceSum = lengths.sumOf { (it.toDouble() - avg).pow(2) }
        val stdDev = sqrt(varianceSum / count)
        val cv = if (avg != 0.0) stdDev / avg else 0.0

        return weight + (count * 0.1).coerceAtMost(50.0) - (cv * 50)
    }

    private fun extractChapters(content: String, indices: List<Int>): List<ParsedChapter> {
        val chapters = mutableListOf<ParsedChapter>()
        val totalLen = content.length

        for (i in indices.indices) {
            val start = indices[i]
            val end = if (i == indices.size - 1) totalLen else indices[i + 1]

            val sliceLimit = (start + 100).coerceAtMost(end)
            val headerSlice = content.substring(start, sliceLimit)
            val trimmedHeader = headerSlice.trimStart()
            val prefixLen = headerSlice.length - trimmedHeader.length
            
            val lineEnd = trimmedHeader.indexOf('\n')
            val titleStr = if (lineEnd != -1) trimmedHeader.substring(0, lineEnd) else trimmedHeader
            val cleanTitle = titleStr.trim()

            val titleLengthFull = prefixLen + (if (lineEnd != -1) lineEnd + 1 else trimmedHeader.length)
            val contentStart = start + titleLengthFull
            
            if (contentStart >= end) continue

            val chapterContent = content.substring(contentStart, end).trim()
            if (chapterContent.length < 5) continue

            chapters.add(
                ParsedChapter(
                    index = i,
                    title = cleanTitle,
                    content = chapterContent,
                    md5 = calculateMD5(chapterContent),
                    wordsCount = chapterContent.length
                )
            )
        }
        return chapters
    }

    private fun calculateMD5(input: String): String {
        val md = MessageDigest.getInstance("MD5")
        val digest = md.digest(input.toByteArray())
        return digest.fold("") { str, it -> str + "%02x".format(it) }
    }
}