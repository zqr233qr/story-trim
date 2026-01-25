package com.storytrim.app.ui.reader.tts

import android.annotation.SuppressLint
import android.content.Context
import android.graphics.Canvas
import android.graphics.Paint
import android.graphics.Typeface
import android.util.AttributeSet
import android.view.GestureDetector
import android.view.MotionEvent
import android.view.View
import android.view.WindowManager
import android.widget.FrameLayout
import android.widget.ImageButton
import android.widget.LinearLayout
import android.widget.TextView
import androidx.core.content.ContextCompat
import com.storytrim.app.R

/**
 * 悬浮窗歌词视图
 * 支持拖动、字体大小调整、颜色选择、锁定穿透
 */
class FloatingLyricsView @JvmOverloads constructor(
    context: Context,
    attrs: AttributeSet? = null,
    defStyleAttr: Int = 0
) : LinearLayout(context, attrs, defStyleAttr) {

    companion object {
        private const val TAG = "FloatingLyricsView"

        const val DEFAULT_TEXT_SIZE = 28f
        const val MIN_TEXT_SIZE = 18f
        const val MAX_TEXT_SIZE = 48f
        const val TEXT_SIZE_STEP = 2f

        val PRESET_COLORS = listOf(
            0xFFFFFFFF.toInt(),  // 白色
            0xFF14B8A6.toInt(),  // 青色
            0xFF22C55E.toInt(),  // 绿色
            0xFFFACC15.toInt(),  // 黄色
            0xFFF472B6.toInt(),  // 粉色
            0xFFFB923C.toInt()   // 橙色
        )
    }

    private var lyricsText: String = ""
    private var isPlaying: Boolean = false

    private var textSize: Float = DEFAULT_TEXT_SIZE
    private var textColor: Int = 0xFFFFFFFF.toInt()
    private var isLocked: Boolean = false
    private var isColorPanelVisible: Boolean = false

    private val lyricsTextView: TextView
    private val controlContainer: LinearLayout
    private val colorPanel: LinearLayout
    private val playPauseButton: ImageButton
    private val lockButton: ImageButton

    private val textPaint = Paint(Paint.ANTI_ALIAS_FLAG).apply {
        textAlign = Paint.Align.CENTER
        typeface = Typeface.DEFAULT_BOLD
    }

    private val gestureDetector: GestureDetector

    private var _playPauseCallback: (() -> Unit)? = null
    private var _closeCallback: (() -> Unit)? = null
    private var _textSizeCallback: ((Float) -> Unit)? = null
    private var _textColorCallback: ((Int) -> Unit)? = null
    private var _lockCallback: ((Boolean) -> Unit)? = null

    private var windowLayoutParams: WindowManager.LayoutParams? = null
    private var windowManager: WindowManager? = null

    private var lastTouchX = 0f
    private var lastTouchY = 0f

    init {
        orientation = VERTICAL
        setBackgroundResource(R.drawable.bg_floating_window)

        inflate(context, R.layout.view_floating_lyrics, this)

        lyricsTextView = findViewById(R.id.lyrics_text)
        controlContainer = findViewById(R.id.control_container)
        colorPanel = findViewById(R.id.color_panel)
        playPauseButton = findViewById(R.id.btn_play_pause)
        lockButton = findViewById(R.id.btn_lock)

        windowManager = context.getSystemService(Context.WINDOW_SERVICE) as WindowManager

        setTextSize(textSize)
        setTextColor(textColor)
        setupClickListeners()
        setupColorPanel()

        gestureDetector = GestureDetector(context, GestureListener())

        setBackgroundResource(R.drawable.bg_floating_window)
        setPadding(24, 16, 24, 16)

        isClickable = true
        isFocusable = true
    }

    private fun setupClickListeners() {
        setupButtonWithFeedback(findViewById(R.id.btn_decrease)) {
            decreaseTextSize()
        }
        setupButtonWithFeedback(findViewById(R.id.btn_increase)) {
            increaseTextSize()
        }
        setupButtonWithFeedback(findViewById(R.id.btn_color)) {
            toggleColorPanel()
        }
        setupButtonWithFeedback(lockButton) {
            toggleLock()
        }
        setupButtonWithFeedback(playPauseButton) {
            _playPauseCallback?.invoke()
        }
        setupButtonWithFeedback(findViewById(R.id.btn_close)) {
            _closeCallback?.invoke()
        }
    }

    private fun setupButtonWithFeedback(button: android.widget.ImageButton, onClick: () -> Unit) {
        button.setOnClickListener {
            it.alpha = 0.5f
            it.postDelayed({ it.alpha = 1.0f }, 100)
            onClick()
        }
    }

    private fun setupColorPanel() {
        val colorViews = listOf(
            R.id.color_white to PRESET_COLORS[0],
            R.id.color_cyan to PRESET_COLORS[1],
            R.id.color_green to PRESET_COLORS[2],
            R.id.color_yellow to PRESET_COLORS[3],
            R.id.color_pink to PRESET_COLORS[4],
            R.id.color_orange to PRESET_COLORS[5]
        )

        colorViews.forEach { (viewId, color) ->
            val view = findViewById<View>(viewId)
            view.background = createColorDrawable(color)
            view.setOnClickListener { selectColor(color) }
        }
    }

    private fun createColorDrawable(color: Int): android.graphics.drawable.ShapeDrawable {
        val drawable = android.graphics.drawable.ShapeDrawable()
        drawable.shape = android.graphics.drawable.shapes.OvalShape()
        drawable.paint.color = color
        drawable.paint.style = Paint.Style.FILL
        drawable.setIntrinsicWidth(32)
        drawable.setIntrinsicHeight(32)
        return drawable
    }

    fun updateLyrics(sentence: String, playing: Boolean) {
        this.lyricsText = sentence
        this.isPlaying = playing

        lyricsTextView.text = sentence.ifEmpty { "等待播放..." }

        playPauseButton.setImageResource(
            if (playing) R.drawable.ic_pause else R.drawable.ic_play
        )

        invalidate()
    }

    fun setTextSize(size: Float) {
        textSize = size.coerceIn(MIN_TEXT_SIZE, MAX_TEXT_SIZE)
        lyricsTextView.textSize = textSize / resources.displayMetrics.density
        invalidate()
    }

    fun getTextSize(): Float = textSize

    fun decreaseTextSize() {
        setTextSize(textSize - TEXT_SIZE_STEP)
        _textSizeCallback?.invoke(textSize)
    }

    fun increaseTextSize() {
        setTextSize(textSize + TEXT_SIZE_STEP)
        _textSizeCallback?.invoke(textSize)
    }

    fun setTextColor(color: Int) {
        textColor = color
        lyricsTextView.setTextColor(color)
        invalidate()
    }

    fun getTextColor(): Int = textColor

    private fun selectColor(color: Int) {
        setTextColor(color)
        _textColorCallback?.invoke(color)
        hideColorPanel()
    }

    private fun toggleColorPanel() {
        if (isColorPanelVisible) {
            hideColorPanel()
        } else {
            showColorPanel()
        }
    }

    private fun showColorPanel() {
        isColorPanelVisible = true
        colorPanel.visibility = VISIBLE
    }

    private fun hideColorPanel() {
        isColorPanelVisible = false
        colorPanel.visibility = GONE
    }

    fun isColorPanelVisible(): Boolean = isColorPanelVisible

    fun toggleLock() {
        isLocked = !isLocked
        updateLockState()
        _lockCallback?.invoke(isLocked)
    }

    private fun updateLockState() {
        if (isLocked) {
            lockButton.setImageResource(R.drawable.ic_unlock)
            controlContainer.visibility = GONE
            hideColorPanel()
            background = null
            setPadding(0, 0, 0, 0)
            isClickable = false
            isFocusable = false
        } else {
            lockButton.setImageResource(R.drawable.ic_lock)
            controlContainer.visibility = VISIBLE
            setBackgroundResource(R.drawable.bg_floating_window)
            setPadding(24, 16, 24, 16)
            isClickable = true
            isFocusable = true
        }
    }

    fun isLocked(): Boolean = isLocked

    fun setPlayPauseCallback(callback: () -> Unit) {
        this._playPauseCallback = callback
    }

    fun setCloseCallback(callback: () -> Unit) {
        this._closeCallback = callback
    }

    fun setTextSizeCallback(callback: (Float) -> Unit) {
        this._textSizeCallback = callback
    }

    fun setTextColorCallback(callback: (Int) -> Unit) {
        this._textColorCallback = callback
    }

    fun setLockCallback(callback: (Boolean) -> Unit) {
        this._lockCallback = callback
    }

    fun getCurrentLayoutParams(): WindowManager.LayoutParams? {
        return windowLayoutParams
    }

    fun saveWindowLayoutParams(params: WindowManager.LayoutParams) {
        this.windowLayoutParams = params
    }

    fun updatePosition(dx: Int, dy: Int) {
        windowLayoutParams?.let { params ->
            params.x += dx
            params.y += dy
            windowManager?.updateViewLayout(this, params)
        }
    }

    @SuppressLint("ClickableViewAccessibility")
    override fun onTouchEvent(event: MotionEvent): Boolean {
        if (isLocked) {
            return false
        }

        when (event.action) {
            MotionEvent.ACTION_DOWN -> {
                lastTouchX = event.rawX
                lastTouchY = event.rawY
                return true
            }
            MotionEvent.ACTION_MOVE -> {
                val dx = (event.rawX - lastTouchX).toInt()
                val dy = (event.rawY - lastTouchY).toInt()
                if (dx != 0 || dy != 0) {
                    updatePosition(dx, dy)
                    lastTouchX = event.rawX
                    lastTouchY = event.rawY
                }
                return true
            }
        }
        gestureDetector.onTouchEvent(event)
        return true
    }

    private inner class GestureListener : GestureDetector.SimpleOnGestureListener() {
        override fun onSingleTapConfirmed(e: MotionEvent): Boolean {
            if (!isLocked && controlContainer.visibility == GONE) {
                controlContainer.visibility = VISIBLE
            }
            return true
        }

        override fun onDoubleTap(e: MotionEvent): Boolean {
            if (!isLocked) {
                _playPauseCallback?.invoke()
            }
            return true
        }
    }
}
