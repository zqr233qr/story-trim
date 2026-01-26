package com.storytrim.app.ui.points

import android.os.Bundle
import android.view.View
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.core.view.ViewCompat
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.recyclerview.widget.LinearLayoutManager
import com.storytrim.app.databinding.ActivityPointsBinding
import com.storytrim.app.ui.points.adapter.PointsLedgerAdapter
import com.storytrim.app.ui.common.ToastHelper
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class PointsActivity : AppCompatActivity() {

    private lateinit var binding: ActivityPointsBinding
    private val viewModel: PointsViewModel by viewModels()
    private lateinit var adapter: PointsLedgerAdapter

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        WindowCompat.setDecorFitsSystemWindows(window, false)
        binding = ActivityPointsBinding.inflate(layoutInflater)
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

        adapter = PointsLedgerAdapter()
        binding.recyclerRecords.layoutManager = LinearLayoutManager(this)
        binding.recyclerRecords.adapter = adapter

        viewModel.balance.observe(this) { balance ->
            binding.tvPointsBalance.text = balance.toString()
        }

        viewModel.records.observe(this) { records ->
            adapter.submitList(records)
            binding.emptyView.visibility = if (records.isEmpty()) View.VISIBLE else View.GONE
        }

        viewModel.isLoading.observe(this) { loading ->
            binding.loadingView.visibility = if (loading) View.VISIBLE else View.GONE
        }

        viewModel.error.observe(this) { message ->
            if (!message.isNullOrBlank()) {
                ToastHelper.show(this, message)
            }
        }

        viewModel.load()
    }
}
