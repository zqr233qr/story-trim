package com.storytrim.app.ui.shelf

import android.content.Intent
import android.os.Bundle
import android.view.View
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.core.view.ViewCompat
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.recyclerview.widget.LinearLayoutManager
import com.storytrim.app.databinding.ActivityShelfBinding
import com.storytrim.app.ui.shelf.adapter.BookAdapter
import com.storytrim.app.ui.login.LoginActivity
import com.storytrim.app.ui.reader.ReaderActivity
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class ShelfActivity : AppCompatActivity() {

    private lateinit var binding: ActivityShelfBinding
    private val viewModel: ShelfViewModel by viewModels()
    private lateinit var adapter: BookAdapter

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
        adapter = BookAdapter { book ->
            val intent = Intent(this, ReaderActivity::class.java)
            intent.putExtra("book_id", book.id)
            startActivity(intent)
        }
        binding.recyclerViewBooks.layoutManager = LinearLayoutManager(this)
        binding.recyclerViewBooks.adapter = adapter
    }

    private fun setupObservers() {
        viewModel.books.observe(this) { books ->
            adapter.submitList(books)
            binding.textViewBookCount.text = "${books.size} 本"
            
            if (books.isEmpty()) {
                binding.emptyView.visibility = View.VISIBLE
                binding.recyclerViewBooks.visibility = View.GONE
            } else {
                binding.emptyView.visibility = View.GONE
                binding.recyclerViewBooks.visibility = View.VISIBLE
            }
        }
    }
    
    private fun setupListeners() {
        binding.cardUpload.setOnClickListener {
            importLauncher.launch("text/plain")
        }
        
        binding.avatarContainer.setOnClickListener {
             startActivity(Intent(this, LoginActivity::class.java))
        }

        binding.avatarContainer.setOnLongClickListener {
            startActivity(Intent(this, com.storytrim.app.ui.debug.DebugSqlActivity::class.java))
            true
        }
    }
}