package com.storytrim.app.ui.login

import android.os.Bundle
import android.widget.Toast
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import com.storytrim.app.databinding.ActivityRegisterBinding
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class RegisterActivity : AppCompatActivity() {

    private lateinit var binding: ActivityRegisterBinding
    private val viewModel: LoginViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityRegisterBinding.inflate(layoutInflater)
        setContentView(binding.root)

        setupViews()
        setupObservers()
    }

    private fun setupViews() {
        binding.btnRegister.setOnClickListener {
            val username = binding.etUsername.text.toString().trim()
            val password = binding.etPassword.text.toString().trim()
            val confirmPassword = binding.etConfirmPassword.text.toString().trim()

            if (username.isEmpty()) {
                Toast.makeText(this, "请输入用户名", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            if (password.isEmpty()) {
                Toast.makeText(this, "请输入密码", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            if (confirmPassword.isEmpty()) {
                Toast.makeText(this, "请确认密码", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            if (password != confirmPassword) {
                Toast.makeText(this, "两次密码不一致", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }

            viewModel.register(username, password)
        }

        binding.tvGoToLogin.setOnClickListener {
            finish()
        }
    }

    private fun setupObservers() {
        viewModel.isLoading.observe(this) { isLoading ->
            binding.btnRegister.isEnabled = !isLoading
            binding.btnRegister.text = if (isLoading) "注册中..." else "注册"
        }

        viewModel.registerResult.observe(this) { result ->
            result.onSuccess {
                Toast.makeText(this, "注册成功，请登录", Toast.LENGTH_SHORT).show()
                finish()
            }.onFailure { e ->
                Toast.makeText(this, "注册失败：${e.message}", Toast.LENGTH_SHORT).show()
            }
        }
    }
}
