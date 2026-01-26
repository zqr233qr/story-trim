package com.storytrim.app.ui.reader

import android.os.Bundle
import android.view.View
import android.view.WindowManager
import androidx.core.content.ContextCompat
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.core.view.ViewCompat
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import androidx.lifecycle.lifecycleScope
import com.storytrim.app.databinding.ActivityReaderBinding
import com.storytrim.app.data.repository.AuthRepository
import com.storytrim.app.ui.common.LoginRequiredDialogFragment
import com.storytrim.app.ui.common.ToastHelper
import com.storytrim.app.ui.reader.core.ReaderView
import com.storytrim.app.ui.reader.core.TextPaginator
import com.storytrim.app.ui.reader.dialog.TtsControlPanelDialog
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.launch
import javax.inject.Inject

@AndroidEntryPoint
class ReaderActivity : AppCompatActivity() {

    private lateinit var binding: ActivityReaderBinding
    private val viewModel: ReaderViewModel by viewModels()

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
    private var isMenuVisible = false
    private var isLoggedIn = false
    private var bookSyncState = 0
    private var topToolbarBasePaddingTop: Int = 0

    private val readerPrefs by lazy {
        getSharedPreferences("reader_prefs", MODE_PRIVATE)
    }

    @Inject
    lateinit var authRepository: AuthRepository

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        WindowCompat.setDecorFitsSystemWindows(window, false)

        if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.P) {
            window.attributes.layoutInDisplayCutoutMode =
                android.view.WindowManager.LayoutParams.LAYOUT_IN_DISPLAY_CUTOUT_MODE_SHORT_EDGES
        }

        binding = ActivityReaderBinding.inflate(layoutInflater)
        setContentView(binding.root)

        topToolbarBasePaddingTop = binding.topToolbar.paddingTop

        val bookId = intent.getLongExtra("book_id", -1)
        if (bookId == -1L) {
            ToastHelper.show(this, "书籍参数错误")
            finish()
            return
        }

        window.statusBarColor = android.graphics.Color.TRANSPARENT

        viewModel.initTtsController(this)
        setupSystemBars()
        setupReaderView()
        setupObservers()
        setupListeners()
        observeLoginState()

        viewModel.loadBook(bookId)
    }

    override fun onDestroy() {
        super.onDestroy()
        viewModel.releaseTts()
    }

    private fun setupSystemBars() {
        val windowInsetsController = WindowCompat.getInsetsController(window, window.decorView)
        windowInsetsController.systemBarsBehavior = WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
        windowInsetsController.isAppearanceLightStatusBars = true
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
        val readerHeight = metrics.heightPixels

        val fontSizeSp = currentFontSizeSp
        val paddingHorizontal = (20 * density).toInt()
        val paddingTop = (48 * density).toInt()
        val paddingBottom = 0

        paginator = TextPaginator(
            width = metrics.widthPixels,
            height = readerHeight,
            fontSizePx = fontSizeSp * density,
            paddingHorizontal = paddingHorizontal,
            paddingTop = paddingTop,
            paddingBottom = paddingBottom
        )

        applyReadingMode(currentReadingMode)
        applyBackground()
        binding.readerView.setPageMode(currentPageMode)

        binding.readerView.onPrevClick = {
            if (isMenuVisible) {
                viewModel.toggleMenu()
            } else {
                handlePrevAction()
            }
        }

        binding.readerView.onNextClick = {
            if (isMenuVisible) {
                viewModel.toggleMenu()
            } else {
                handleNextAction()
            }
        }

        binding.readerView.onMenuClick = {
            viewModel.toggleMenu()
        }

        binding.readerView.post {
            rebuildPaginator()
        }
    }

    private fun observeLoginState() {
        lifecycleScope.launch {
            authRepository.isLoggedInFlow().collect { token ->
                isLoggedIn = token.isNotBlank()
            }
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
        binding.readerView.setPageMode(mode)
    }

    override fun dispatchKeyEvent(event: android.view.KeyEvent): Boolean {
        if (event.action == android.view.KeyEvent.ACTION_DOWN && currentPageMode == "click") {
            when (event.keyCode) {
                android.view.KeyEvent.KEYCODE_VOLUME_UP -> {
                    if (isMenuVisible) {
                        viewModel.toggleMenu()
                    } else {
                        handlePrevAction()
                    }
                    return true
                }
                android.view.KeyEvent.KEYCODE_VOLUME_DOWN -> {
                    if (isMenuVisible) {
                        viewModel.toggleMenu()
                    } else {
                        handleNextAction()
                    }
                    return true
                }
            }
        }
        return super.dispatchKeyEvent(event)
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
        val readerHeight = metrics.heightPixels
        paginator = TextPaginator(
            width = metrics.widthPixels,
            height = readerHeight,
            fontSizePx = currentFontSizeSp * density,
            paddingHorizontal = (20 * density).toInt(),
            paddingTop = (48 * density).toInt(),
            paddingBottom = 0
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

            paginator?.let {
                currentPages = it.paginate(state.text)

                when (state.scrollTo) {
                    ReaderViewModel.NavigationStrategy.START -> currentPageIndex = 0
                    ReaderViewModel.NavigationStrategy.END -> currentPageIndex = (currentPages.size - 1).coerceAtLeast(0)
                    ReaderViewModel.NavigationStrategy.KEEP -> {}
                }

                displayPage()

                viewModel.prepareTts(state.text, state.title)
            }
        }

        viewModel.isLoading.observe(this) { isLoading ->
            binding.progressBar.visibility = if (isLoading) View.VISIBLE else View.GONE
        }

        viewModel.book.observe(this) { book ->
            binding.tvBookTitle.text = book.title
            bookSyncState = book.syncState
        }

        viewModel.activeModeName.observe(this) { modeName ->
            binding.readerView.setModeName(modeName)
        }

        viewModel.toastMessage.observe(this) { message ->
            if (!message.isNullOrBlank()) {
                ToastHelper.show(this, message)
            }
        }

        viewModel.isMenuVisible.observe(this) { isVisible ->
            isMenuVisible = isVisible
            if (isVisible) {
                showSystemUI()
                window.clearFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
                rebuildPaginator()
                applyTopToolbarInsets()
                binding.topToolbar.visibility = View.VISIBLE
                binding.bottomMenu.visibility = View.VISIBLE
                binding.btnMagic.visibility = View.VISIBLE
            } else {
                hideSystemUI()
                window.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
                rebuildPaginator()
                resetTopToolbarPadding()
                binding.topToolbar.visibility = View.GONE
                binding.bottomMenu.visibility = View.GONE
                binding.btnMagic.visibility = View.GONE
            }
        }

        viewModel.isMagicActive.observe(this) { isActive ->
            if (isActive) {
                binding.btnMagic.backgroundTintList = android.content.res.ColorStateList.valueOf(
                    ContextCompat.getColor(this, com.storytrim.app.R.color.storytrim_accent)
                )
                binding.btnMagic.imageTintList = android.content.res.ColorStateList.valueOf(
                    ContextCompat.getColor(this, com.storytrim.app.R.color.white)
                )
            } else {
                binding.btnMagic.backgroundTintList = android.content.res.ColorStateList.valueOf(
                    ContextCompat.getColor(this, com.storytrim.app.R.color.storytrim_surface_muted)
                )
                binding.btnMagic.imageTintList = android.content.res.ColorStateList.valueOf(
                    ContextCompat.getColor(this, com.storytrim.app.R.color.storytrim_text_primary)
                )
            }
        }

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

        viewModel.isTtsPanelVisible.observe(this) { isVisible ->
            if (isVisible) {
                showTtsControlPanel()
            }
        }

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
            if (!isLoggedIn) {
                showLoginDialog("精简功能需要登录账号，登录后即可使用智能精简。")
                return@setOnClickListener
            }
            viewModel.toggleMagic()
        }

        binding.btnMagic.setOnLongClickListener {
            if (!isLoggedIn) {
                showLoginDialog("精简功能需要登录账号，登录后即可使用智能精简。")
                return@setOnLongClickListener true
            }
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
            if (!isLoggedIn) {
                showLoginDialog("精简功能需要登录账号，登录后即可使用智能精简。")
                return@setOnClickListener
            }
            if (bookSyncState == 0) {
                ToastHelper.show(this, "本地书籍暂不支持指定精简")
                return@setOnClickListener
            }
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

    private fun handlePrevAction() {
        if (currentPageMode != "click") {
            if (currentPageIndex > 0) {
                currentPageIndex--
                displayPage()
            } else {
                viewModel.loadPrevChapter()
            }
            return
        }

        if (currentPageIndex > 0) {
            currentPageIndex--
            displayPage()
        } else {
            viewModel.loadPrevChapter()
        }
    }

    private fun handleNextAction() {
        if (currentPageMode != "click") {
            turnToNextPage()
            return
        }

        turnToNextPage()
    }

    private fun showLoginDialog(message: String) {
        LoginRequiredDialogFragment.newInstance(message)
            .show(supportFragmentManager, "LoginRequiredDialog")
    }

    private fun applyTopToolbarInsets() {
        ViewCompat.setOnApplyWindowInsetsListener(binding.topToolbar) { view, windowInsets ->
            val topInset = windowInsets.getInsets(WindowInsetsCompat.Type.statusBars()).top
            view.setPadding(
                view.paddingLeft,
                topToolbarBasePaddingTop + topInset,
                view.paddingRight,
                view.paddingBottom
            )
            WindowInsetsCompat.CONSUMED
        }
        ViewCompat.requestApplyInsets(binding.topToolbar)
    }

    private fun resetTopToolbarPadding() {
        binding.topToolbar.setPadding(
            binding.topToolbar.paddingLeft,
            topToolbarBasePaddingTop,
            binding.topToolbar.paddingRight,
            binding.topToolbar.paddingBottom
        )
    }

    private fun syncHighlightWithTts(sentenceIndex: Int, sentence: String) {
        if (sentence.isEmpty()) return

        binding.readerView.setHighlightByIndex(sentenceIndex, sentence)

        if (currentPages.isNotEmpty() && currentPageIndex < currentPages.size) {
            val pageContent = currentPages[currentPageIndex]
            if (!pageContent.contains(sentence)) {
                findAndGotoSentencePage(sentence, sentenceIndex)
            }
        }
    }

    private fun scrollToHighlight(start: Int) {
        val paginator = paginator ?: return
        val paint = paginator.getPaint()
        val lineHeight = paint.textSize * 1.5f

        val layout = binding.readerView.getLayout() ?: return
        val startLine = layout.getLineForOffset(start)

        val contentTop = paginator.paddingTop.toFloat()
        val viewHeight = binding.readerView.height
        val targetY = contentTop + (startLine * lineHeight) - (viewHeight / 2) + (lineHeight / 2)

        binding.readerView.post {
            binding.readerView.scrollTo(0, targetY.toInt().coerceAtLeast(0))
        }
    }

    private fun findAndGotoSentencePage(sentence: String, sentenceIndex: Int) {
        for ((pageIdx, pageContent) in currentPages.withIndex()) {
            if (pageContent.contains(sentence)) {
                if (pageIdx != currentPageIndex) {
                    currentPageIndex = pageIdx
                    displayPage()
                }

                binding.readerView.post {
                    val sentenceStart = pageContent.indexOf(sentence)
                    val sentenceEnd = sentenceStart + sentence.length
                    binding.readerView.setHighlightRangeInternal(sentenceStart, sentenceEnd)
                    scrollToHighlight(sentenceStart)
                }
                return
            }
        }

        binding.readerView.post {
            binding.readerView.setHighlightByIndex(sentenceIndex, sentence)
        }
    }
}
