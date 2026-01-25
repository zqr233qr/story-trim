package com.storytrim.app

import android.content.Intent
import android.os.Bundle
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import com.storytrim.app.core.network.AuthInterceptor
import com.storytrim.app.ui.login.LoginActivity
import com.storytrim.app.ui.shelf.ShelfActivity
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import javax.inject.Inject

@AndroidEntryPoint
class MainActivity : AppCompatActivity() {

    @Inject
    lateinit var authInterceptor: AuthInterceptor

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        // 这里可以设置一个简单的 Splash 布局，或者保持空白（因为跳转很快）
        // setContentView(R.layout.activity_splash) 
        
        checkLoginStatus()
    }

    private fun checkLoginStatus() {
        lifecycleScope.launch {
            // 给一点点缓冲时间，避免画面一闪而过，同时也确保 DataStore 初始化
            // delay(200) 
            
            val token = authInterceptor.getToken()
            
            if (token.isEmpty()) {
                startActivity(Intent(this@MainActivity, LoginActivity::class.java))
            } else {
                startActivity(Intent(this@MainActivity, ShelfActivity::class.java))
            }
            finish()
        }
    }
}
