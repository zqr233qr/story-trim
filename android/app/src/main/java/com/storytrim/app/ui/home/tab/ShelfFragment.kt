package com.storytrim.app.ui.home.tab

import android.content.Intent
import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import com.storytrim.app.ui.common.ToastHelper
import androidx.core.view.ViewCompat
import androidx.core.view.WindowInsetsCompat
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.recyclerview.widget.LinearLayoutManager
import com.storytrim.app.databinding.ActivityShelfBinding
import com.storytrim.app.ui.login.LoginActivity
import com.storytrim.app.ui.reader.ReaderActivity
import com.storytrim.app.ui.shelf.ShelfViewModel
import com.storytrim.app.ui.shelf.adapter.BookAdapter
import com.storytrim.app.ui.shelf.dialog.BookActionSheetDialogFragment
import com.storytrim.app.ui.shelf.dialog.ProgressOverlayDialogFragment
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class ShelfFragment : Fragment() {

    private var _binding: ActivityShelfBinding? = null
    private val binding get() = _binding!!
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

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = ActivityShelfBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        setupInsets()
        setupRecyclerView()
        setupObservers()
        setupListeners()
    }

    private fun setupInsets() {
        ViewCompat.setOnApplyWindowInsetsListener(binding.root) { _, windowInsets ->
            val insets = windowInsets.getInsets(WindowInsetsCompat.Type.systemBars())
            binding.containerContent.setPadding(
                binding.containerContent.paddingLeft,
                insets.top + 60.dpToPx(),
                binding.containerContent.paddingRight,
                binding.containerContent.paddingBottom
            )
            WindowInsetsCompat.CONSUMED
        }
    }

    private fun Int.dpToPx(): Int {
        val density = resources.displayMetrics.density
        return (this * density).toInt()
    }

    private fun setupRecyclerView() {
        adapter = BookAdapter(
            onBookClick = { book ->
                if (isBusy) {
                    ToastHelper.show(requireContext(), "任务处理中，请稍候")
                    return@BookAdapter
                }
                val intent = Intent(requireContext(), ReaderActivity::class.java)
                intent.putExtra("book_id", book.id)
                startActivity(intent)
            },
            onBookLongClick = { book ->
                if (isBusy) {
                    ToastHelper.show(requireContext(), "任务处理中，请稍候")
                    return@BookAdapter
                }
                showActionSheet(book)
            }
        )
        binding.recyclerViewBooks.layoutManager = LinearLayoutManager(requireContext())
        binding.recyclerViewBooks.adapter = adapter
    }

    private fun setupObservers() {
        viewModel.books.observe(viewLifecycleOwner) { books ->
            allBooks = books
            updateBookList()
        }

        viewModel.isLoggedIn.observe(viewLifecycleOwner) { loggedIn ->
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
                ToastHelper.show(requireContext(), "任务处理中，请稍候")
                return@setOnClickListener
            }
            importLauncher.launch("*/*")
        }

        binding.avatarContainer.setOnClickListener {
            if (!isLoggedIn) {
                startActivity(Intent(requireContext(), LoginActivity::class.java))
            }
        }

        binding.avatarContainer.setOnLongClickListener {
            startActivity(Intent(requireContext(), com.storytrim.app.ui.debug.DebugSqlActivity::class.java))
            true
        }
    }

    private fun showActionSheet(book: com.storytrim.app.data.model.Book) {
        if (isBusy) {
            ToastHelper.show(requireContext(), "任务处理中，请稍候")
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
        dialog.show(parentFragmentManager, "BookActionSheet")
    }

    private fun handleSyncBook(book: com.storytrim.app.data.model.Book) {
        if (isBusy) {
            ToastHelper.show(requireContext(), "任务处理中，请稍候")
            return
        }
        if (!isLoggedIn) {
            showLoginDialog("同步功能需要登录账号，登录后即可多端同步阅读进度。")
            return
        }
        showProgressDialog("正在同步至云端...")
        viewModel.syncBook(book) { progress ->
            updateProgressDialog(progress)
        }
    }

    private fun handleDownloadBook(book: com.storytrim.app.data.model.Book) {
        if (isBusy) {
            ToastHelper.show(requireContext(), "任务处理中，请稍候")
            return
        }
        if (!isLoggedIn) {
            showLoginDialog("下载功能需要登录账号，登录后即可同步离线内容。")
            return
        }
        if (book.cloudId <= 0) {
            ToastHelper.show(requireContext(), "该书籍暂无云端版本")
            return
        }
        showProgressDialog("正在下载到本地...")
        viewModel.downloadBook(book) { progress ->
            updateProgressDialog(progress)
        }
    }

    private fun confirmDeleteBook(book: com.storytrim.app.data.model.Book) {
        if (isBusy) {
            ToastHelper.show(requireContext(), "任务处理中，请稍候")
            return
        }
        if (!isLoggedIn) {
            if (book.userId > 0) {
                ToastHelper.show(requireContext(), "该书籍属于其他账号，无法删除")
                return
            }
            if (book.syncState == 1) {
                ToastHelper.show(requireContext(), "该书籍为云端书籍，未登录状态下无法删除")
                return
            }
        }

        androidx.appcompat.app.AlertDialog.Builder(requireContext())
            .setTitle("删除书籍")
            .setMessage("确定删除本书吗？此操作不可恢复。")
            .setPositiveButton("删除") { _, _ -> viewModel.deleteBook(book) }
            .setNegativeButton("取消", null)
            .show()
    }

    private fun showLoginDialog(message: String) {
        androidx.appcompat.app.AlertDialog.Builder(requireContext())
            .setTitle("需要登录")
            .setMessage(message)
            .setPositiveButton("去登录") { _, _ -> startActivity(Intent(requireContext(), LoginActivity::class.java)) }
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
            progressDialog?.show(parentFragmentManager, "ShelfProgress")
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

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
