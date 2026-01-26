package com.storytrim.app

import android.content.Intent
import android.os.Bundle
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import com.storytrim.app.ui.home.HomeActivity
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import javax.inject.Inject

@AndroidEntryPoint
class MainActivity : AppCompatActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        // 这里可以设置一个简单的 Splash 布局，或者保持空白（因为跳转很快）
        // setContentView(R.layout.activity_splash) 
        
        launchHome()
    }

    private fun launchHome() {
        lifecycleScope.launch {
            startActivity(Intent(this@MainActivity, HomeActivity::class.java))
            finish()
        }
    }
}
