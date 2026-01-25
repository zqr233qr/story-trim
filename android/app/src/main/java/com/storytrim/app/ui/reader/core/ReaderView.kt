package com.storytrim.app.ui.reader.core

import android.content.Context
import android.graphics.Canvas
import android.graphics.Color
import android.graphics.Paint
import android.text.Layout
import android.text.StaticLayout
import android.text.TextPaint
import android.util.AttributeSet
import android.util.Log
import android.view.MotionEvent
import android.view.View
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

class ReaderView @JvmOverloads constructor(
    context: Context, attrs: AttributeSet? = null, defStyleAttr: Int = 0
) : View(context, attrs, defStyleAttr) {

    private var content: String = ""
    private var chapterTitle: String = ""
    private var modeName: String = ""
    private var currentPage: Int = 0
    private var totalPages: Int = 0
    private var batteryLevel: Int = 100
    
    private var paginator: TextPaginator? = null
    
    // TTS高亮相关
    private var highlightStart: Int = -1
    private var highlightEnd: Int = -1
    private var highlightColor: Int = Color.parseColor("#14B8A6") // Teal-500
    
    // Paints
    private val statusPaint = TextPaint(Paint.ANTI_ALIAS_FLAG).apply {
        textSize = 36f // ~12sp
        color = Color.GRAY
    }
    
    private val highlightPaint = Paint(Paint.ANTI_ALIAS_FLAG).apply {
        color = Color.parseColor("#14B8A6")
        alpha = 80 // 半透明
    }
    
    // Callbacks
    var onPrevClick: (() -> Unit)? = null
    var onNextClick: (() -> Unit)? = null
    var onMenuClick: (() -> Unit)? = null

    fun setPageData(pageContent: String, title: String, current: Int, total: Int, textPaginator: TextPaginator) {
        this.content = pageContent
        this.chapterTitle = title
        this.currentPage = current
        this.totalPages = total
        this.paginator = textPaginator
        invalidate()
    }

    fun setModeName(name: String) {
        modeName = name
        invalidate()
    }

    /**
     * 通过句子索引设置高亮
     * 使用全文匹配方式查找句子位置
     * @param sentenceIndex 句子索引
     * @param sentence 句子内容
     */
    fun setHighlightByIndex(sentenceIndex: Int, sentence: String) {
        if (content.isEmpty()) {
            clearHighlight()
            return
        }

        val start = content.indexOf(sentence)
        if (start >= 0) {
            val end = start + sentence.length
            setHighlightRangeInternal(start, end)
            Log.d("ReaderView", "Highlight by index: $sentenceIndex -> $start-$end")
        } else {
            clearHighlight()
        }
    }

    /**
     * 内部方法：直接设置高亮范围
     */
    fun setHighlightRangeInternal(start: Int, end: Int) {
        this.highlightStart = start
        this.highlightEnd = end
        invalidate()
    }

    /**
     * 清除高亮
     */
    fun clearHighlight() {
        highlightStart = -1
        highlightEnd = -1
        invalidate()
    }

    fun updateStatusTextColor(color: Int) {
        statusPaint.color = color
        invalidate()
    }

    override fun onDraw(canvas: Canvas) {
        super.onDraw(canvas)
        if (content.isEmpty() || paginator == null) return

        // 1. Draw Content
        drawContent(canvas)
        
        // 2. Draw Header (Chapter Title)
        drawHeader(canvas)
        
        // 3. Draw Footer (Page num, Time, Battery)
        drawFooter(canvas)
    }
    
    private fun drawContent(canvas: Canvas) {
        val paint = paginator!!.getPaint()
        canvas.save()
        canvas.translate(paginator!!.paddingHorizontal.toFloat(), paginator!!.paddingTop.toFloat())
        
        val layout = StaticLayout.Builder.obtain(content, 0, content.length, paint, width - paginator!!.paddingHorizontal * 2)
            .setAlignment(Layout.Alignment.ALIGN_NORMAL)
            .setLineSpacing(0f, 1.5f)
            .setIncludePad(false)
            .build()
        
        // 如果有高亮范围，先绘制背景
        if (highlightStart >= 0 && highlightEnd > highlightStart) {
            drawHighlightBackground(canvas, layout)
        }
        
        layout.draw(canvas)
        canvas.restore()
    }
    
    private fun drawHighlightBackground(canvas: Canvas, layout: StaticLayout) {
        val startLine = layout.getLineForOffset(highlightStart)
        val endLine = layout.getLineForOffset(highlightEnd)
        
        for (line in startLine..endLine) {
            val lineStart = maxOf(highlightStart, layout.getLineStart(line))
            val lineEnd = minOf(highlightEnd, layout.getLineEnd(line))
            
            if (lineStart < lineEnd) {
                val left = layout.getLineLeft(line)
                val top = layout.getLineTop(line)
                val right = layout.getLineRight(line)
                val bottom = layout.getLineBottom(line)
                
                // 调整矩形以匹配文本实际位置
                val paint = paginator!!.getPaint()
                val startX = left + paint.measureText(content.substring(layout.getLineStart(line), lineStart))
                val endX = left + paint.measureText(content.substring(layout.getLineStart(line), lineEnd))
                
                canvas.drawRect(
                    startX,
                    top.toFloat(),
                    endX,
                    bottom.toFloat(),
                    highlightPaint
                )
            }
        }
    }
    
    private fun drawHeader(canvas: Canvas) {
        val x = paginator!!.paddingHorizontal.toFloat()
        val y = paginator!!.paddingTop.toFloat() / 2 + 10
        val modeText = modeName
        val modeWidth = if (modeText.isBlank()) 0f else statusPaint.measureText(modeText)
        val gap = dpToPx(12).toFloat()
        val availableTitleWidth = if (modeWidth > 0f) {
            width.toFloat() - paginator!!.paddingHorizontal * 2 - modeWidth - gap
        } else {
            width.toFloat() - paginator!!.paddingHorizontal * 2
        }

        val titleText = ellipsizeText(chapterTitle, statusPaint, availableTitleWidth)
        canvas.drawText(titleText, x, y, statusPaint)

        if (modeWidth > 0f) {
            val modeX = width - paginator!!.paddingHorizontal.toFloat() - modeWidth
            canvas.drawText(modeText, modeX, y, statusPaint)
        }
    }

    private fun ellipsizeText(text: String, paint: TextPaint, maxWidth: Float): String {
        if (maxWidth <= 0f || paint.measureText(text) <= maxWidth) return text
        val ellipsis = "..."
        val ellipsisWidth = paint.measureText(ellipsis)
        if (ellipsisWidth >= maxWidth) return ellipsis
        var end = text.length
        while (end > 0 && paint.measureText(text, 0, end) + ellipsisWidth > maxWidth) {
            end--
        }
        return if (end <= 0) ellipsis else text.substring(0, end) + ellipsis
    }

    private fun dpToPx(dp: Int): Int {
        return (dp * resources.displayMetrics.density).toInt()
    }
    
    private fun drawFooter(canvas: Canvas) {
        val y = height - (paginator!!.paddingBottom.toFloat() / 2)
        
        // Battery (Left)
        val batX = paginator!!.paddingHorizontal.toFloat()
        canvas.drawText("电量 $batteryLevel%", batX, y, statusPaint)
        
        // Time (Center)
        val time = SimpleDateFormat("HH:mm", Locale.getDefault()).format(Date())
        val timeWidth = statusPaint.measureText(time)
        val timeX = (width - timeWidth) / 2
        canvas.drawText(time, timeX, y, statusPaint)
        
        // Page Number (Right)
        val pageStr = "${currentPage + 1}/$totalPages"
        val pageX = width - paginator!!.paddingHorizontal.toFloat() - statusPaint.measureText(pageStr)
        canvas.drawText(pageStr, pageX, y, statusPaint)
    }

    override fun onTouchEvent(event: MotionEvent): Boolean {
        if (event.action == MotionEvent.ACTION_UP) {
            val x = event.x
            val w = width
            
            when {
                x < w * 0.33 -> onPrevClick?.invoke()
                x > w * 0.66 -> onNextClick?.invoke()
                else -> onMenuClick?.invoke()
            }
            return true
        }
        return true
    }
    
    // Add text accessibility for TextPaginator access
    fun getPaginator() = paginator
    
    /**
     * 获取当前内容的Layout对象
     */
    fun getLayout(): StaticLayout? {
        if (content.isEmpty() || paginator == null) return null
        val paint = paginator!!.getPaint()
        return StaticLayout.Builder.obtain(content, 0, content.length, paint, width - paginator!!.paddingHorizontal * 2)
            .setAlignment(Layout.Alignment.ALIGN_NORMAL)
            .setLineSpacing(0f, 1.5f)
            .setIncludePad(false)
            .build()
    }
}
