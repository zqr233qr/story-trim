package com.storytrim.app

import android.os.Bundle
import android.util.Log
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity

class SimpleMainActivity : AppCompatActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        Log.d("SimpleMainActivity", "onCreate 开始")
        
        try {
            // 完全独立，不依赖任何其他类
            val textView = TextView(this).apply {
                text = "SimpleMainActivity 启动成功！\n\n点击任意位置跳转到下一个页面"
                textSize = 18f
                gravity = android.view.Gravity.CENTER
                setTextColor(0xFF000000.toInt())
                setBackgroundColor(0xFFFFFFFF.toInt())
            }
            
            textView.setOnClickListener {
                Log.d("SimpleMainActivity", "点击事件触发")
                try {
            // 跳转到最简单的下一个页面
            val intent = android.content.Intent(this, SimpleTestActivity::class.java)
            Log.d("SimpleMainActivity", "准备跳转到 SimpleTestActivity: ${SimpleTestActivity::class.java.name}")
            startActivity(intent)
            Log.d("SimpleMainActivity", "startActivity 调用完成")
            finish()
            Log.d("SimpleMainActivity", "finish 调用完成")
                } catch (e: Exception) {
                    Log.e("SimpleMainActivity", "跳转失败: ${e.message}", e)
                }
            }
            
            setContentView(textView)
            Log.d("SimpleMainActivity", "setContentView 完成")
            
        } catch (e: Exception) {
            Log.e("SimpleMainActivity", "创建页面失败: ${e.message}", e)
            // 即使出错，也显示一个简单的 TextView
            val errorView = TextView(this).apply {
                text = "页面创建错误: ${e.message}"
                textSize = 16f
                gravity = android.view.Gravity.CENTER
                setTextColor(0xFFFF0000.toInt())
                setBackgroundColor(0xFFFFFF00.toInt())
            }
            setContentView(errorView)
        }
    }
}

class SimpleTestActivity : AppCompatActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        Log.d("SimpleTestActivity", "SimpleTestActivity onCreate 开始")
        
        try {
            val textView = TextView(this).apply {
                text = "SimpleTestActivity 启动成功！\n\n基础功能正常！"
                textSize = 20f
                gravity = android.view.Gravity.CENTER
                setTextColor(0xFF000000.toInt())
                setBackgroundColor(0xFFE8F5E8.toInt())
            }
            
            setContentView(textView)
            Log.d("SimpleTestActivity", "SimpleTestActivity setContentView 完成")
            
        } catch (e: Exception) {
            Log.e("SimpleTestActivity", "创建 SimpleTestActivity 失败: ${e.message}", e)
            val errorView = TextView(this).apply {
                text = "SimpleTestActivity 错误: ${e.message}"
                textSize = 16f
                gravity = android.view.Gravity.CENTER
                setTextColor(0xFFFF0000.toInt())
                setBackgroundColor(0xFFFFFF00.toInt())
            }
            setContentView(errorView)
        }
    }
}