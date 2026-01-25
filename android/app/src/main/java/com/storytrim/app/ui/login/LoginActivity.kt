package com.storytrim.app.ui.login

import android.content.Intent
import android.os.Bundle
import android.widget.Toast
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.core.view.ViewCompat
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import com.storytrim.app.databinding.ActivityLoginBinding
import com.storytrim.app.ui.shelf.ShelfActivity
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class LoginActivity : AppCompatActivity() {

    private lateinit var binding: ActivityLoginBinding
    private val viewModel: LoginViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        // 开启 Edge-to-Edge
        WindowCompat.setDecorFitsSystemWindows(window, false)
        
        binding = ActivityLoginBinding.inflate(layoutInflater)
        setContentView(binding.root)

        // 处理状态栏遮挡
        ViewCompat.setOnApplyWindowInsetsListener(binding.root) { view, windowInsets ->
            val insets = windowInsets.getInsets(WindowInsetsCompat.Type.systemBars())
            view.setPadding(0, 0, 0, insets.bottom)
            // 给内容容器加顶部 padding
            binding.contentContainer.setPadding(
                binding.contentContainer.paddingLeft,
                insets.top + 20.dpToPx(), 
                binding.contentContainer.paddingRight,
                binding.contentContainer.paddingBottom
            )
            WindowInsetsCompat.CONSUMED
        }

        setupViews()
        setupObservers()
    }
    
    private fun Int.dpToPx(): Int {
        val density = resources.displayMetrics.density
        return (this * density).toInt()
    }

    private fun setupViews() {
        binding.btnLogin.setOnClickListener {
            val username = binding.etUsername.text.toString().trim()
            val password = binding.etPassword.text.toString().trim()

            if (username.isEmpty()) {
                Toast.makeText(this, "请输入用户名", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            if (password.isEmpty()) {
                Toast.makeText(this, "请输入密码", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }

            viewModel.login(username, password)
        }

        binding.tvGoToRegister.setOnClickListener {
            startActivity(Intent(this, RegisterActivity::class.java))
        }
    }

    private fun setupObservers() {
        viewModel.isLoading.observe(this) { isLoading ->
            binding.btnLogin.isEnabled = !isLoading
            binding.btnLogin.text = if (isLoading) "登录中..." else "登录"
        }

        viewModel.loginResult.observe(this) { result ->
            result.onSuccess {
                // Toast.makeText(this, "登录成功", Toast.LENGTH_SHORT).show()
                startActivity(Intent(this, ShelfActivity::class.java))
                finish()
            }.onFailure { e ->
                Toast.makeText(this, "登录失败：${e.message}", Toast.LENGTH_SHORT).show()
            }
        }
    }
}