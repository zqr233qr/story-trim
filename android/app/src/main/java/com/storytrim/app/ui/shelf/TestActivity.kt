package com.storytrim.app.ui.shelf

import android.os.Bundle
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity

class TestActivity : AppCompatActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        // 极简测试，不依赖任何布局文件
        val textView = TextView(this).apply {
            text = "TestActivity 启动成功！\n\n这是一个极简的测试页面。"
            textSize = 20f
            gravity = android.view.Gravity.CENTER
        }
        
        setContentView(textView)
    }
}
