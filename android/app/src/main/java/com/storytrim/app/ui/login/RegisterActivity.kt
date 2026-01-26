package com.storytrim.app.ui.login

import android.os.Bundle
import com.storytrim.app.ui.common.ToastHelper
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
                ToastHelper.show(this, "请输入用户名")
                return@setOnClickListener
            }
            if (password.isEmpty()) {
                ToastHelper.show(this, "请输入密码")
                return@setOnClickListener
            }
            if (confirmPassword.isEmpty()) {
                ToastHelper.show(this, "请确认密码")
                return@setOnClickListener
            }
            if (password != confirmPassword) {
                ToastHelper.show(this, "两次密码不一致")
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
                ToastHelper.show(this, "注册成功，请登录")
                finish()
            }.onFailure { e ->
                ToastHelper.show(this, "注册失败：${e.message}")
            }
        }
    }
}
