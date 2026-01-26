package com.storytrim.app.ui.tasks

import android.os.Bundle
import android.view.View
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.core.view.ViewCompat
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.recyclerview.widget.LinearLayoutManager
import com.storytrim.app.databinding.ActivityTaskCenterBinding
import com.storytrim.app.ui.tasks.adapter.TaskItemAdapter
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class TaskCenterActivity : AppCompatActivity() {

    private lateinit var binding: ActivityTaskCenterBinding
    private lateinit var adapter: TaskItemAdapter
    private val viewModel: TaskCenterViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        WindowCompat.setDecorFitsSystemWindows(window, false)
        binding = ActivityTaskCenterBinding.inflate(layoutInflater)
        setContentView(binding.root)

        ViewCompat.setOnApplyWindowInsetsListener(binding.root) { view, windowInsets ->
            val insets = windowInsets.getInsets(WindowInsetsCompat.Type.systemBars())
            binding.container.setPadding(
                binding.container.paddingLeft,
                insets.top + (16 * resources.displayMetrics.density).toInt(),
                binding.container.paddingRight,
                binding.container.paddingBottom
            )
            view.setPadding(0, 0, 0, insets.bottom)
            WindowInsetsCompat.CONSUMED
        }

        binding.btnBack.setOnClickListener { finish() }

        adapter = TaskItemAdapter()
        binding.recyclerTasks.layoutManager = LinearLayoutManager(this)
        binding.recyclerTasks.adapter = adapter

        viewModel.tasks.observe(this) { tasks ->
            adapter.submitList(tasks)
            binding.emptyView.visibility = if (tasks.isEmpty()) View.VISIBLE else View.GONE
        }

        viewModel.isLoading.observe(this) { loading ->
            binding.loadingView.visibility = if (loading) View.VISIBLE else View.GONE
        }
    }

    override fun onStart() {
        super.onStart()
        viewModel.startPolling()
    }

    override fun onStop() {
        super.onStop()
        viewModel.stopPolling()
    }
}
