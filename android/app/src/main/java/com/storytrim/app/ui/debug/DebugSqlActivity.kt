package com.storytrim.app.ui.debug

import android.os.Bundle
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import com.storytrim.app.databinding.ActivityDebugSqlBinding
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class DebugSqlActivity : AppCompatActivity() {

    private lateinit var binding: ActivityDebugSqlBinding
    private val viewModel: DebugSqlViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityDebugSqlBinding.inflate(layoutInflater)
        setContentView(binding.root)

        binding.btnExecute.setOnClickListener {
            val sql = binding.etSql.text.toString()
            viewModel.executeSql(sql)
        }

        viewModel.result.observe(this) { result ->
            binding.tvResult.text = result
        }

        viewModel.isLoading.observe(this) { isLoading ->
            binding.btnExecute.isEnabled = !isLoading
            binding.btnExecute.text = if (isLoading) "正在执行..." else "执行 SQL"
        }
    }
}
