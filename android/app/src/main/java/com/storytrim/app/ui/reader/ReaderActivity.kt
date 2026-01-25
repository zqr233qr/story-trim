package com.storytrim.app.ui.reader

import android.os.Bundle
import android.util.DisplayMetrics
import android.view.View
import android.widget.Toast
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.core.view.ViewCompat
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import com.storytrim.app.databinding.ActivityReaderBinding
import com.storytrim.app.ui.reader.core.ReaderView
import com.storytrim.app.ui.reader.core.TextPaginator
import com.storytrim.app.ui.reader.dialog.TtsControlPanelDialog
import com.storytrim.app.ui.reader.tts.TtsController
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class ReaderActivity : AppCompatActivity() {

    private lateinit var binding: ActivityReaderBinding
    private val viewModel: ReaderViewModel by viewModels()
    
    // Pagination state
    private var currentPageIndex = 0
    private var currentPages: List<String> = emptyList()
    private var currentTitle: String = ""
    private var paginator: TextPaginator? = null
    private var currentContent: String = ""
    private var currentFontSizeSp = 18f
    private var currentReadingMode = "light"
    private var hideStatusBar = false
    private var currentPageMode = "click"
    private var currentBackgroundPath: String? = null

    private val readerPrefs by lazy {
        getSharedPreferences("reader_prefs", MODE_PRIVATE)
    }
    
    // TTS
    private var ttsForegroundService: com.storytrim.app.ui.reader.tts.TtsForegroundService? = null
    private var ttsServiceBound = false

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        WindowCompat.setDecorFitsSystemWindows(window, false)
        
        // 适配刘海屏，允许内容延伸到刘海区域
        if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.P) {
            window.attributes.layoutInDisplayCutoutMode = 
                android.view.WindowManager.LayoutParams.LAYOUT_IN_DISPLAY_CUTOUT_MODE_SHORT_EDGES
        }
        
        binding = ActivityReaderBinding.inflate(layoutInflater)
        setContentView(binding.root)

        val bookId = intent.getLongExtra("book_id", -1)
        if (bookId == -1L) {
            Toast.makeText(this, "书籍参数错误", Toast.LENGTH_SHORT).show()
            finish()
            return
        }
        
        window.statusBarColor = android.graphics.Color.TRANSPARENT

        viewModel.initTtsController(this)
        setupSystemBars()
        setupReaderView()
        setupObservers()
        setupListeners()

        viewModel.loadBook(bookId)
    }
    
    override fun onDestroy() {
        super.onDestroy()
        viewModel.releaseTts()
    }
    
    private fun setupSystemBars() {
        val windowInsetsController = WindowCompat.getInsetsController(window, window.decorView)
        // 允许通过滑动唤出系统栏
        windowInsetsController.systemBarsBehavior = WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
        
        // 重要：设置状态栏图标为深色（因为阅读背景是浅色）
        windowInsetsController.isAppearanceLightStatusBars = true
        
        // 初始状态：隐藏系统栏 (全屏沉浸)
        hideSystemUI()
    }

    private fun hideSystemUI() {
        WindowCompat.getInsetsController(window, window.decorView).hide(WindowInsetsCompat.Type.systemBars())
    }

    private fun showSystemUI() {
        WindowCompat.getInsetsController(window, window.decorView).show(WindowInsetsCompat.Type.systemBars())
    }

    private fun setupReaderView() {
        loadReaderPrefs()
        val metrics = resources.displayMetrics
        val density = metrics.density
        
        val fontSizeSp = currentFontSizeSp
        val paddingHorizontal = (20 * density).toInt()
        val paddingTop = (48 * density).toInt()
        val paddingBottom = (16 * density).toInt()
        
        paginator = TextPaginator(
            width = metrics.widthPixels,
            height = metrics.heightPixels,
            fontSizePx = fontSizeSp * density,
            paddingHorizontal = paddingHorizontal,
            paddingTop = paddingTop,
            paddingBottom = paddingBottom
        )

        applyReadingMode(currentReadingMode)
        applyBackground()

        binding.readerView.onPrevClick = {
            if (currentPageIndex > 0) {
                currentPageIndex--
                displayPage()
            } else {
                viewModel.loadPrevChapter()
            }
        }

        binding.readerView.onNextClick = {
            turnToNextPage()
        }

        binding.readerView.onMenuClick = {
            viewModel.toggleMenu()
        }
    }

    private fun loadReaderPrefs() {
        currentFontSizeSp = readerPrefs.getFloat("reader_font_size_sp", 18f)
        currentReadingMode = readerPrefs.getString("reader_reading_mode", "light") ?: "light"
        hideStatusBar = true
        currentPageMode = readerPrefs.getString("reader_page_mode", "click") ?: "click"
        currentBackgroundPath = readerPrefs.getString("reader_background_path", null)
    }

    private fun applyReadingMode(mode: String) {
        val (bgColor, textColor, statusColor) = when (mode) {
            "dark" -> Triple(0xFF0C0A09.toInt(), 0xFFE5E5E5.toInt(), 0xFF9CA3AF.toInt())
            "sepia" -> Triple(0xFFF5E6D3.toInt(), 0xFF5D4E37.toInt(), 0xFF7C6F5A.toInt())
            else -> Triple(0xFFFAF9F6.toInt(), 0xFF1C1917.toInt(), 0xFF9CA3AF.toInt())
        }

        if (hasCustomBackground()) {
            binding.rootLayout.setBackgroundColor(android.graphics.Color.TRANSPARENT)
            binding.readerView.setBackgroundColor(android.graphics.Color.TRANSPARENT)
        } else {
            binding.rootLayout.setBackgroundColor(bgColor)
            binding.readerView.setBackgroundColor(bgColor)
        }
        paginator?.setTextColor(textColor)
        binding.readerView.updateStatusTextColor(statusColor)
        binding.tvBookTitle.setTextColor(textColor)
        binding.tvCurrentMode.setTextColor(statusColor)

        WindowCompat.getInsetsController(window, window.decorView).isAppearanceLightStatusBars = mode != "dark"
    }

    fun updateFontSize(sizeSp: Float) {
        currentFontSizeSp = sizeSp
        readerPrefs.edit().putFloat("reader_font_size_sp", sizeSp).apply()
        rebuildPaginator()
    }

    fun updateReadingMode(mode: String) {
        currentReadingMode = mode
        readerPrefs.edit().putString("reader_reading_mode", mode).apply()
        applyReadingMode(mode)
        if (hideStatusBar) {
            hideSystemUI()
        }
    }

    fun updateHideStatusBar(enabled: Boolean) {
        hideStatusBar = enabled
        readerPrefs.edit().putBoolean("reader_hide_status_bar", enabled).apply()
        if (enabled) {
            hideSystemUI()
        } else {
            showSystemUI()
        }
    }

    fun updatePageMode(mode: String) {
        currentPageMode = mode
        readerPrefs.edit().putString("reader_page_mode", mode).apply()
    }

    fun getCurrentBackgroundPath(): String? = currentBackgroundPath

    fun updateBackgroundImage(path: String?) {
        currentBackgroundPath = path
        readerPrefs.edit().putString("reader_background_path", path).apply()
        applyBackground()
        applyReadingMode(currentReadingMode)
    }

    fun getCurrentFontSizeSp(): Float = currentFontSizeSp

    fun getCurrentReadingMode(): String = currentReadingMode

    fun getCurrentPageMode(): String = currentPageMode

    fun getHideStatusBarSetting(): Boolean = hideStatusBar

    private fun rebuildPaginator() {
        val metrics = resources.displayMetrics
        val density = metrics.density
        paginator = TextPaginator(
            width = metrics.widthPixels,
            height = metrics.heightPixels,
            fontSizePx = currentFontSizeSp * density,
            paddingHorizontal = (20 * density).toInt(),
            paddingTop = (48 * density).toInt(),
            paddingBottom = (16 * density).toInt()
        )
        applyReadingMode(currentReadingMode)
        if (currentContent.isNotBlank()) {
            currentPages = paginator?.paginate(currentContent) ?: emptyList()
            currentPageIndex = currentPageIndex.coerceAtMost((currentPages.size - 1).coerceAtLeast(0))
            displayPage()
        }
    }

    private fun hasCustomBackground(): Boolean {
        return !currentBackgroundPath.isNullOrBlank()
    }

    private fun applyBackground() {
        val path = currentBackgroundPath
        if (path.isNullOrBlank()) {
            binding.ivReaderBackground.visibility = View.GONE
            binding.viewBackgroundOverlay.visibility = View.GONE
            return
        }

        val file = java.io.File(path)
        if (!file.exists()) {
            binding.ivReaderBackground.visibility = View.GONE
            binding.viewBackgroundOverlay.visibility = View.GONE
            return
        }

        val bitmap = android.graphics.BitmapFactory.decodeFile(file.absolutePath)
        if (bitmap == null) {
            binding.ivReaderBackground.visibility = View.GONE
            binding.viewBackgroundOverlay.visibility = View.GONE
            return
        }

        binding.ivReaderBackground.setImageBitmap(bitmap)
        binding.ivReaderBackground.visibility = View.VISIBLE
        binding.viewBackgroundOverlay.visibility = View.VISIBLE
    }
    
    private fun turnToNextPage() {
        if (currentPageIndex < currentPages.size - 1) {
            currentPageIndex++
            displayPage()
        } else {
            viewModel.loadNextChapter()
        }
    }

    private fun displayPage() {
        if (currentPages.isNotEmpty() && currentPageIndex >= 0 && currentPageIndex < currentPages.size) {
            paginator?.let {
                val pageContent = currentPages[currentPageIndex]
                binding.readerView.setPageData(
                    pageContent, 
                    currentTitle, 
                    currentPageIndex, 
                    currentPages.size, 
                    it
                )
            }
        }
    }
    
    private fun setupObservers() {
        viewModel.contentState.observe(this) { state ->
            currentTitle = state.title
            currentContent = state.text
            
            // Calculate pages
            paginator?.let {
                currentPages = it.paginate(state.text)
                
                when (state.scrollTo) {
                    ReaderViewModel.NavigationStrategy.START -> currentPageIndex = 0
                    ReaderViewModel.NavigationStrategy.END -> currentPageIndex = (currentPages.size - 1).coerceAtLeast(0)
                    ReaderViewModel.NavigationStrategy.KEEP -> {}
                }
                
                displayPage()
                
                // 准备TTS内容
                viewModel.prepareTts(state.text, state.title)
            }
        }

        viewModel.isLoading.observe(this) { isLoading ->
            binding.progressBar.visibility = if (isLoading) View.VISIBLE else View.GONE
        }
        
        viewModel.book.observe(this) { book ->
            binding.tvBookTitle.text = book.title
        }

        viewModel.activeModeName.observe(this) { modeName ->
            binding.tvCurrentMode.text = modeName
            binding.readerView.setModeName(modeName)
        }

        viewModel.toastMessage.observe(this) { message ->
            if (!message.isNullOrBlank()) {
                Toast.makeText(this, message, Toast.LENGTH_SHORT).show()
            }
        }

        viewModel.isMenuVisible.observe(this) { isVisible ->
            if (isVisible) {
                if (!hideStatusBar) {
                    showSystemUI()
                } else {
                    hideSystemUI()
                }
                binding.topToolbar.visibility = View.VISIBLE
                binding.bottomMenu.visibility = View.VISIBLE
                binding.btnMagic.visibility = View.VISIBLE
            } else {
                hideSystemUI()
                binding.topToolbar.visibility = View.GONE
                binding.bottomMenu.visibility = View.GONE
                binding.btnMagic.visibility = View.GONE
            }
        }

        // Observe magic button state
        viewModel.isMagicActive.observe(this) { isActive ->
            if (isActive) {
                binding.btnMagic.backgroundTintList = android.content.res.ColorStateList.valueOf(android.graphics.Color.parseColor("#0D9488")) // Teal
                binding.btnMagic.imageTintList = android.content.res.ColorStateList.valueOf(android.graphics.Color.WHITE)
            } else {
                binding.btnMagic.backgroundTintList = android.content.res.ColorStateList.valueOf(android.graphics.Color.parseColor("#E5E5E4")) // Stone-200
                binding.btnMagic.imageTintList = android.content.res.ColorStateList.valueOf(android.graphics.Color.parseColor("#1C1917"))
            }
        }

        // Observe terminal event
        viewModel.showTerminalEvent.observe(this) { show ->
            if (show) {
                val dialog = com.storytrim.app.ui.reader.dialog.GenerationTerminalDialogFragment()
                dialog.show(supportFragmentManager, "GenerationTerminalDialog")
            }
        }

        viewModel.showModeConfigEvent.observe(this) {
            val dialog = com.storytrim.app.ui.reader.dialog.AiTrimConfigDialogFragment()
            dialog.show(supportFragmentManager, com.storytrim.app.ui.reader.dialog.AiTrimConfigDialogFragment.TAG)
        }

        // Observe TTS panel visibility
        viewModel.isTtsPanelVisible.observe(this) { isVisible ->
            if (isVisible) {
                showTtsControlPanel()
            } else {
                // Panel will be dismissed by user action
            }
        }

        // Observe TTS sentence changes for highlight sync
        viewModel.ttsController.observe(this) { controller ->
            controller?.currentSentenceWithIndex?.observe(this) { ttsSentence ->
                if (ttsSentence != null && ttsSentence.index >= 0) {
                    syncHighlightWithTts(ttsSentence.index, ttsSentence.sentence)
                } else {
                    binding.readerView.clearHighlight()
                }
            }
        }
    }
    
    private fun setupListeners() {
        binding.btnMagic.setOnClickListener {
            viewModel.toggleMagic()
        }

        binding.btnMagic.setOnLongClickListener {
            val dialog = com.storytrim.app.ui.reader.dialog.AiTrimConfigDialogFragment()
            dialog.show(supportFragmentManager, com.storytrim.app.ui.reader.dialog.AiTrimConfigDialogFragment.TAG)
            true
        }

        binding.btnBack.setOnClickListener {
            finish()
        }
        
        binding.btnCatalog.setOnClickListener {
            val dialog = com.storytrim.app.ui.reader.dialog.ChapterListDialogFragment()
            dialog.show(supportFragmentManager, "ChapterListDialog")
        }
        
        binding.btnSettings.setOnClickListener {
            val dialog = com.storytrim.app.ui.reader.dialog.ReaderSettingsDialogFragment.newInstance()
            dialog.show(supportFragmentManager, com.storytrim.app.ui.reader.dialog.ReaderSettingsDialogFragment.TAG)
        }

        binding.btnBatchTrim.setOnClickListener {
            val dialog = com.storytrim.app.ui.reader.dialog.ChapterTrimDialogFragment()
            dialog.show(supportFragmentManager, "ChapterTrimDialog")
        }

        binding.btnPrevChapter.setOnClickListener {
            viewModel.loadPrevChapter()
        }

        binding.btnNextChapter.setOnClickListener {
            viewModel.loadNextChapter()
        }

        binding.btnTtsTop.setOnClickListener {
            viewModel.showTtsPanel()
        }
    }

    private fun showTtsControlPanel() {
        val dialog = TtsControlPanelDialog.newInstance()
        dialog.show(supportFragmentManager, TtsControlPanelDialog.TAG)
    }

    /**
     * 同步TTS高亮到阅读页
     * 使用索引匹配方式设置高亮
     * 翻页模式下不需要滚动
     */
    private fun syncHighlightWithTts(sentenceIndex: Int, sentence: String) {
        if (sentence.isEmpty()) return

        // 使用索引匹配方式设置高亮
        binding.readerView.setHighlightByIndex(sentenceIndex, sentence)

        // 检查当前页是否包含该句子
        if (currentPages.isNotEmpty() && currentPageIndex < currentPages.size) {
            val pageContent = currentPages[currentPageIndex]
            if (!pageContent.contains(sentence)) {
                // 句子不在当前页，需要翻页
                findAndGotoSentencePage(sentence, sentenceIndex)
            }
            // 句子在当前页，只设置高亮（翻页模式下不需要滚动）
        }
    }

    /**
     * 滚动到高亮位置（居中显示）
     */
    private fun scrollToHighlight(start: Int, end: Int) {
        val paginator = paginator ?: return
        val paint = paginator.getPaint()
        val lineHeight = paint.textSize * 1.5f
        
        // 计算高亮在第几行
        val layout = binding.readerView.getLayout() ?: return
        val startLine = layout.getLineForOffset(start)
        
        // 计算目标Y位置（让高亮行居中）
        val contentTop = paginator.paddingTop.toFloat()
        val viewHeight = binding.readerView.height
        val targetY = contentTop + (startLine * lineHeight) - (viewHeight / 2) + (lineHeight / 2)
        
        // 使用post实现平滑滚动
        binding.readerView.post {
            binding.readerView.scrollTo(0, targetY.toInt().coerceAtLeast(0))
        }
    }

    /**
     * 查找句子所在的页面并跳转
     */
    private fun findAndGotoSentencePage(sentence: String, sentenceIndex: Int) {
        val paginator = paginator ?: return

        // 在所有页面中查找
        for ((pageIdx, pageContent) in currentPages.withIndex()) {
            if (pageContent.contains(sentence)) {
                if (pageIdx != currentPageIndex) {
                    currentPageIndex = pageIdx
                    displayPage()
                }

                // 延迟设置高亮和滚动，确保页面已渲染
                binding.readerView.post {
                    val sentenceStart = pageContent.indexOf(sentence)
                    val sentenceEnd = sentenceStart + sentence.length
                    binding.readerView.setHighlightRangeInternal(sentenceStart, sentenceEnd)
                    scrollToHighlight(sentenceStart, sentenceEnd)
                }
                return
            }
        }

        // 如果没找到，使用索引匹配尝试
        binding.readerView.post {
            binding.readerView.setHighlightByIndex(sentenceIndex, sentence)
        }
    }

    // Auto-flip for TTS
}
