package com.storytrim.app.ui.reader.core

import android.graphics.Paint
import android.text.Layout
import android.text.StaticLayout
import android.text.TextPaint
import java.util.ArrayList

class TextPaginator(
    private val width: Int,
    private val height: Int,
    private val fontSizePx: Float,
    private val lineSpacingMult: Float = 1.5f,
    val paddingHorizontal: Int,
    val paddingTop: Int,
    val paddingBottom: Int
) {
    private val paint = TextPaint(Paint.ANTI_ALIAS_FLAG).apply {
        textSize = fontSizePx
        color = 0xFF1C1917.toInt() // Stone-900
    }

    private val contentWidth = width - paddingHorizontal * 2
    private val contentHeight = height - paddingTop - paddingBottom

    fun paginate(content: String): List<String> {
        val pages = ArrayList<String>()
        
        // Using StaticLayout to measure text flow
        val layout = StaticLayout.Builder.obtain(content, 0, content.length, paint, contentWidth)
            .setAlignment(Layout.Alignment.ALIGN_NORMAL)
            .setLineSpacing(0f, lineSpacingMult)
            .setIncludePad(false)
            .build()

        var currentLine = 0
        while (currentLine < layout.lineCount) {
            val pageStartLine = currentLine
            var pageHeight = 0
            
            // Calculate how many lines fit in one page
            while (currentLine < layout.lineCount) {
                val lineHeight = layout.getLineBottom(currentLine) - layout.getLineTop(currentLine)
                if (pageHeight + lineHeight > contentHeight) {
                    break
                }
                pageHeight += lineHeight
                currentLine++
            }
            
            // Extract text for this page
            val startOffset = layout.getLineStart(pageStartLine)
            val endOffset = layout.getLineEnd(currentLine - 1)
            val pageText = content.substring(startOffset, endOffset)
            pages.add(pageText)
        }
        
        // Handle empty content or if logic fails
        if (pages.isEmpty() && content.isNotEmpty()) {
            pages.add(content)
        }
        
        return pages
    }
    
    fun getPaint(): TextPaint = paint

    fun setTextColor(color: Int) {
        paint.color = color
    }
}
