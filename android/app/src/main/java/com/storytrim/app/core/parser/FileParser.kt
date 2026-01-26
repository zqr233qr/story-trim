package com.storytrim.app.core.parser

import java.io.BufferedReader
import java.io.File
import java.io.InputStream
import java.io.InputStreamReader
import java.nio.charset.Charset
import java.security.MessageDigest
import java.util.regex.Pattern
import javax.inject.Inject
import javax.inject.Singleton
import kotlin.math.pow
import kotlin.math.sqrt
import java.util.zip.ZipFile
import javax.xml.parsers.DocumentBuilderFactory
import org.w3c.dom.Element
import dagger.hilt.android.qualifiers.ApplicationContext
import android.content.Context
import java.net.URLDecoder

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
class FileParser @Inject constructor(
    @ApplicationContext private val context: Context
) {

    private val rules = listOf(
        ParserRule("Strict_Chinese", """(?:^|\n)第[0-9零一二三四五六七八九十百千万]+[章回节][ \t\f].*""", 100.0),
        ParserRule("Normal_Chinese", """(?:^|\n)第[0-9零一二三四五六七八九十百千万]+[章回节].*""", 90.0),
        ParserRule("Strict_English", """(?:^|\n)Chapter\s+\d+.*""", 80.0),
        ParserRule("Loose_Number", """(?:^|\n)\d+\.\s+.*""", 60.0),
        ParserRule("Loose_Direct", """(?:^|\n)[0-9零一二三四五六七八九十百千万]+\s+.*""", 40.0)
    )

    fun parse(inputStream: InputStream, fileName: String): ParseResult {
        return if (fileName.lowercase().endsWith(".epub")) {
            val tempFile = File(context.cacheDir, "import_${System.currentTimeMillis()}.epub")
            tempFile.outputStream().use { output -> inputStream.copyTo(output) }
            try {
                parseEpub(tempFile, fileName)
            } finally {
                tempFile.delete()
            }
        } else {
            parseTxt(inputStream, fileName)
        }
    }

    private fun parseTxt(inputStream: InputStream, fileName: String): ParseResult {
        val bytes = inputStream.readBytes()
        val charset = if (isUtf8(bytes)) Charsets.UTF_8 else Charset.forName("GBK")
        val content = String(bytes, charset)
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

    private fun isUtf8(bytes: ByteArray): Boolean {
        var i = 0
        while (i < bytes.size) {
            val b = bytes[i].toInt() and 0xFF
            if (b <= 0x7F) {
                i += 1
                continue
            }

            if (b in 0xC2..0xDF) {
                if (i + 1 >= bytes.size) return true
                val b1 = bytes[i + 1].toInt() and 0xFF
                if (b1 !in 0x80..0xBF) return false
                i += 2
                continue
            }

            if (b in 0xE0..0xEF) {
                if (i + 2 >= bytes.size) return true
                val b1 = bytes[i + 1].toInt() and 0xFF
                val b2 = bytes[i + 2].toInt() and 0xFF
                if (b == 0xE0 && b1 !in 0xA0..0xBF) return false
                if (b == 0xED && b1 !in 0x80..0x9F) return false
                if (b != 0xE0 && b != 0xED && b1 !in 0x80..0xBF) return false
                if (b2 !in 0x80..0xBF) return false
                i += 3
                continue
            }

            if (b in 0xF0..0xF4) {
                if (i + 3 >= bytes.size) return true
                val b1 = bytes[i + 1].toInt() and 0xFF
                val b2 = bytes[i + 2].toInt() and 0xFF
                val b3 = bytes[i + 3].toInt() and 0xFF
                if (b == 0xF0 && b1 !in 0x90..0xBF) return false
                if (b == 0xF4 && b1 !in 0x80..0x8F) return false
                if (b != 0xF0 && b != 0xF4 && b1 !in 0x80..0xBF) return false
                if (b2 !in 0x80..0xBF || b3 !in 0x80..0xBF) return false
                i += 4
                continue
            }

            return false
        }
        return true
    }

    private fun parseEpub(epubFile: File, fileName: String): ParseResult {
        ZipFile(epubFile).use { zip ->
            val containerEntry = zip.getEntry("META-INF/container.xml")
                ?: throw IllegalArgumentException("无效的 EPUB 格式 (未找到 OPF)")
            val containerXml = zip.getInputStream(containerEntry).bufferedReader().use { it.readText() }

            val opfPath = Regex("full-path=[\"']([^\"']+)[\"']").find(containerXml)
                ?.groupValues?.get(1)
                ?: throw IllegalArgumentException("无效的 EPUB 格式 (未找到 OPF)")

            val opfEntry = zip.getEntry(opfPath)
                ?: throw IllegalArgumentException("无效的 EPUB 格式 (OPF 不存在)")
            val opfContent = zip.getInputStream(opfEntry).bufferedReader().use { it.readText() }

            val factory = DocumentBuilderFactory.newInstance()
            factory.isNamespaceAware = true
            val doc = factory.newDocumentBuilder().parse(opfContent.byteInputStream())

            val titleNodes = doc.getElementsByTagName("dc:title")
            val title = if (titleNodes.length > 0) {
                titleNodes.item(0).textContent
            } else {
                fileName.substringBeforeLast(".")
            }

            val manifestMap = mutableMapOf<String, String>()
            val manifestItems = doc.getElementsByTagNameNS("*", "item")
            for (i in 0 until manifestItems.length) {
                val node = manifestItems.item(i) as? Element ?: continue
                val id = node.getAttribute("id")
                val href = node.getAttribute("href")
                if (id.isNotBlank() && href.isNotBlank()) {
                    manifestMap[id] = href
                }
            }

            val spine = mutableListOf<String>()
            val spineItems = doc.getElementsByTagNameNS("*", "itemref")
            for (i in 0 until spineItems.length) {
                val node = spineItems.item(i) as? Element ?: continue
                val idref = node.getAttribute("idref")
                if (idref.isNotBlank()) {
                    spine.add(idref)
                }
            }

            val opfDir = opfPath.substringBeforeLast("/", "")
            val chapters = mutableListOf<ParsedChapter>()
            spine.forEachIndexed { index, idref ->
                val href = manifestMap[idref] ?: return@forEachIndexed
                val decodedHref = URLDecoder.decode(href, "UTF-8")
                val fullPath = if (opfDir.isNotBlank()) "$opfDir/$decodedHref" else decodedHref
                val entry = zip.getEntry(fullPath) ?: zip.getEntry(href) ?: return@forEachIndexed
                val html = zip.getInputStream(entry).bufferedReader().use { it.readText() }
                val content = cleanHtmlContent(html)
                if (content.length < 5) return@forEachIndexed
                val chapterTitle = extractHtmlTitle(html, index)
                chapters.add(
                    ParsedChapter(
                        index = chapters.size,
                        title = chapterTitle,
                        content = content,
                        md5 = calculateMD5(content),
                        wordsCount = content.length
                    )
                )
            }

            if (chapters.isEmpty()) {
                val fallback = zip.entries().asSequence()
                    .map { it.name }
                    .filter { it.startsWith("$opfDir/") && it.endsWith(".html") }
                    .filterNot { it.endsWith("cover.html") }
                    .toList()
                fallback.forEachIndexed { index, name ->
                    val entry = zip.getEntry(name) ?: return@forEachIndexed
                    val html = zip.getInputStream(entry).bufferedReader().use { it.readText() }
                    val content = cleanHtmlContent(html)
                    if (content.length < 5) return@forEachIndexed
                    chapters.add(
                        ParsedChapter(
                            index = chapters.size,
                            title = extractHtmlTitle(html, index),
                            content = content,
                            md5 = calculateMD5(content),
                            wordsCount = content.length
                        )
                    )
                }
            }

            if (chapters.isEmpty()) {
                throw IllegalArgumentException("未能提取到有效章节内容")
            }

            val bookMd5 = calculateFileMD5(epubFile)
            return ParseResult(title = title, bookMd5 = bookMd5, chapters = chapters)
        }
    }

    private fun cleanHtmlContent(html: String): String {
        var content = html
        content = content.replace(Regex("<head[^>]*>[\\s\\S]*?</head>", RegexOption.IGNORE_CASE), "")
        content = content.replace(Regex("<style[^>]*>[\\s\\S]*?</style>", RegexOption.IGNORE_CASE), "")
        content = content.replace(Regex("<script[^>]*>[\\s\\S]*?</script>", RegexOption.IGNORE_CASE), "")
        content = content.replace(Regex("</p>", RegexOption.IGNORE_CASE), "\n")
        content = content.replace(Regex("</div>", RegexOption.IGNORE_CASE), "\n")
        content = content.replace(Regex("<br\\s*/?>", RegexOption.IGNORE_CASE), "\n")
        content = content.replace(Regex("</h[1-6]>", RegexOption.IGNORE_CASE), "\n\n")
        content = content.replace(Regex("<[^>]+>"), "")
        content = content.replace("&nbsp;", " ")
            .replace("&lt;", "<")
            .replace("&gt;", ">")
            .replace("&amp;", "&")
            .replace("&quot;", "\"")
            .replace("&apos;", "'")
        return content.split("\n")
            .map { it.trim() }
            .filter { it.isNotBlank() }
            .joinToString("\n")
    }

    private fun extractHtmlTitle(html: String, index: Int): String {
        val titleMatch = Regex("<title[^>]*>(.*?)</title>", RegexOption.IGNORE_CASE).find(html)
        if (titleMatch != null) {
            val title = titleMatch.groupValues[1].replace(Regex("<[^>]+>"), "").trim()
            if (title.isNotBlank()) return title
        }
        val hMatch = Regex("<h[1-2][^>]*>(.*?)</h[1-2]>", RegexOption.IGNORE_CASE).find(html)
        if (hMatch != null) {
            val title = hMatch.groupValues[1].replace(Regex("<[^>]+>"), "").trim()
            if (title.isNotBlank()) return title
        }
        return "第 ${index + 1} 章节"
    }

    private fun calculateFileMD5(file: File): String {
        val md = MessageDigest.getInstance("MD5")
        file.inputStream().use { input ->
            val buffer = ByteArray(8192)
            var read = input.read(buffer)
            while (read > 0) {
                md.update(buffer, 0, read)
                read = input.read(buffer)
            }
        }
        val digest = md.digest()
        return digest.joinToString("") { "%02x".format(it) }
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
