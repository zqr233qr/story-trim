package com.storytrim.app.ui.shelf

import android.content.Intent
import android.os.Bundle
import android.view.View
import com.storytrim.app.ui.common.ToastHelper
import androidx.activity.viewModels
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import androidx.core.view.ViewCompat
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.recyclerview.widget.LinearLayoutManager
import com.storytrim.app.databinding.ActivityShelfBinding
import com.storytrim.app.ui.shelf.adapter.BookAdapter
import com.storytrim.app.ui.shelf.dialog.BookActionSheetDialogFragment
import com.storytrim.app.ui.shelf.dialog.ProgressOverlayDialogFragment
import com.storytrim.app.ui.login.LoginActivity
import com.storytrim.app.ui.reader.ReaderActivity
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class ShelfActivity : AppCompatActivity() {

    private lateinit var binding: ActivityShelfBinding
    private val viewModel: ShelfViewModel by viewModels()
    private lateinit var adapter: BookAdapter
    private var isLoggedIn: Boolean = false
    private var allBooks: List<com.storytrim.app.data.model.Book> = emptyList()
    private var progressDialog: ProgressOverlayDialogFragment? = null
    private var progressDialogTitle: String = ""
    private var isBusy: Boolean = false

    private val importLauncher = registerForActivityResult(androidx.activity.result.contract.ActivityResultContracts.GetContent()) { uri ->
        uri?.let { viewModel.importBook(it) }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        // 1. 开启边到边模式 (Edge-to-Edge)
        WindowCompat.setDecorFitsSystemWindows(window, false)
        
        binding = ActivityShelfBinding.inflate(layoutInflater)
        setContentView(binding.root)

        // 2. 处理状态栏避让 (Insets)
        // 让主要内容容器避开状态栏高度，但背景依然铺满
        ViewCompat.setOnApplyWindowInsetsListener(binding.root) { view, windowInsets ->
            val insets = windowInsets.getInsets(WindowInsetsCompat.Type.systemBars())
            // 只给顶部添加 padding，保留底部导航栏的处理（如果有的话）
            view.setPadding(0, 0, 0, insets.bottom)
            
            // 给实际的内容容器添加顶部 padding
            binding.containerContent.setPadding(
                binding.containerContent.paddingLeft,
                insets.top + 60.dpToPx(), // 原有的 padding (60dp) + 状态栏高度
                binding.containerContent.paddingRight,
                binding.containerContent.paddingBottom
            )
            
            WindowInsetsCompat.CONSUMED
        }

        setupRecyclerView()
        setupObservers()
        setupListeners()
    }
    
    private fun Int.dpToPx(): Int {
        val density = resources.displayMetrics.density
        return (this * density).toInt()
    }

    private fun setupRecyclerView() {
        adapter = BookAdapter(
            onBookClick = { book ->
                if (isBusy) {
                    ToastHelper.show(this, "任务处理中，请稍候")
                    return@BookAdapter
                }
                val intent = Intent(this, ReaderActivity::class.java)
                intent.putExtra("book_id", book.id)
                startActivity(intent)
            },
            onBookLongClick = { book ->
                if (isBusy) {
                    ToastHelper.show(this, "任务处理中，请稍候")
                    return@BookAdapter
                }
                showActionSheet(book)
            }
        )
        binding.recyclerViewBooks.layoutManager = LinearLayoutManager(this)
        binding.recyclerViewBooks.adapter = adapter
    }

    private fun setupObservers() {
        viewModel.books.observe(this) { books ->
            allBooks = books
            updateBookList()
        }

        viewModel.isLoggedIn.observe(this) { loggedIn ->
            isLoggedIn = loggedIn
            updateBookList()
        }
    }

    private fun updateBookList() {
        val displayBooks = if (isLoggedIn) {
            allBooks
        } else {
            allBooks.filter { it.syncState != 2 }
        }
        adapter.submitList(displayBooks)
        binding.textViewBookCount.text = "${displayBooks.size} 本"

        if (displayBooks.isEmpty()) {
            binding.emptyView.visibility = View.VISIBLE
            binding.recyclerViewBooks.visibility = View.GONE
        } else {
            binding.emptyView.visibility = View.GONE
            binding.recyclerViewBooks.visibility = View.VISIBLE
        }
    }
    
    private fun setupListeners() {
        binding.cardUpload.setOnClickListener {
            if (isBusy) {
                ToastHelper.show(this, "任务处理中，请稍候")
                return@setOnClickListener
            }
            importLauncher.launch("*/*")
        }
        
        binding.avatarContainer.setOnClickListener {
            if (isLoggedIn) {
                ToastHelper.show(this, "已登录")
            } else {
                startActivity(Intent(this, LoginActivity::class.java))
            }
        }

        binding.avatarContainer.setOnLongClickListener {
            startActivity(Intent(this, com.storytrim.app.ui.debug.DebugSqlActivity::class.java))
            true
        }
    }

    private fun showActionSheet(book: com.storytrim.app.data.model.Book) {
        if (isBusy) {
            ToastHelper.show(this, "任务处理中，请稍候")
            return
        }
        val dialog = BookActionSheetDialogFragment.newInstance(
            title = book.title,
            showSync = book.syncState == 0,
            showDownload = book.syncState == 2
        ).setActionListener { action ->
            when (action) {
                BookActionSheetDialogFragment.ActionType.SYNC -> handleSyncBook(book)
                BookActionSheetDialogFragment.ActionType.DOWNLOAD -> handleDownloadBook(book)
                BookActionSheetDialogFragment.ActionType.DELETE -> confirmDeleteBook(book)
            }
        }
        dialog.show(supportFragmentManager, "BookActionSheet")
    }

    @Suppress("UNUSED_PARAMETER")
    private fun handleSyncBook(_book: com.storytrim.app.data.model.Book) {
        if (isBusy) {
            ToastHelper.show(this, "任务处理中，请稍候")
            return
        }
        if (!isLoggedIn) {
            showLoginDialog("同步功能需要登录账号，登录后即可多端同步阅读进度。")
            return
        }
        showProgressDialog("正在同步至云端...")
        viewModel.syncBook(_book) { progress ->
            updateProgressDialog(progress)
        }
    }

    private fun handleDownloadBook(book: com.storytrim.app.data.model.Book) {
        if (isBusy) {
            ToastHelper.show(this, "任务处理中，请稍候")
            return
        }
        if (!isLoggedIn) {
            showLoginDialog("下载功能需要登录账号，登录后即可同步离线内容。")
            return
        }
        if (book.cloudId <= 0) {
            ToastHelper.show(this, "该书籍暂无云端版本")
            return
        }
        showProgressDialog("正在下载到本地...")
        viewModel.downloadBook(book) { progress ->
            updateProgressDialog(progress)
        }
    }

    private fun confirmDeleteBook(book: com.storytrim.app.data.model.Book) {
        if (isBusy) {
            ToastHelper.show(this, "任务处理中，请稍候")
            return
        }
        if (!isLoggedIn) {
            if (book.userId > 0) {
                ToastHelper.show(this, "该书籍属于其他账号，无法删除")
                return
            }
            if (book.syncState == 1) {
                ToastHelper.show(this, "该书籍为云端书籍，未登录状态下无法删除")
                return
            }
        }

        AlertDialog.Builder(this)
            .setTitle("删除书籍")
            .setMessage("确定删除本书吗？此操作不可恢复。")
            .setPositiveButton("删除") { _, _ -> viewModel.deleteBook(book) }
            .setNegativeButton("取消", null)
            .show()
    }

    private fun showLoginDialog(message: String) {
        AlertDialog.Builder(this)
            .setTitle("需要登录")
            .setMessage(message)
            .setPositiveButton("去登录") { _, _ -> startActivity(Intent(this, LoginActivity::class.java)) }
            .setNegativeButton("取消", null)
            .show()
    }

    private fun showProgressDialog(title: String) {
        if (progressDialog == null) {
            progressDialog = ProgressOverlayDialogFragment.newInstance(title)
        }
        progressDialogTitle = title
        setBusy(true)
        if (progressDialog?.isAdded != true) {
            progressDialog?.show(supportFragmentManager, "ShelfProgress")
        }
        updateProgressDialog(0, title)
    }

    private fun updateProgressDialog(progress: Int, titleOverride: String? = null) {
        val dialog = progressDialog ?: return
        val title = titleOverride ?: progressDialogTitle
        dialog.updateProgress(progress, title)

        if (progress <= 0 || progress >= 100) {
            dialog.dismiss()
            setBusy(false)
        }
    }

    private fun setBusy(busy: Boolean) {
        isBusy = busy
        binding.cardUpload.isEnabled = !busy
        binding.recyclerViewBooks.isEnabled = !busy
        binding.cardUpload.alpha = if (busy) 0.6f else 1f
        binding.recyclerViewBooks.alpha = if (busy) 0.6f else 1f
    }
}
