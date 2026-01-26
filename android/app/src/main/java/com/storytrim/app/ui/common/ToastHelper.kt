package com.storytrim.app.ui.common

import android.content.Context
import android.view.Gravity
import android.view.LayoutInflater
import android.widget.Toast
import com.storytrim.app.databinding.ViewToastBinding

object ToastHelper {
    fun show(context: Context, message: String, long: Boolean = false) {
        if (message.isBlank()) return
        val binding = ViewToastBinding.inflate(LayoutInflater.from(context))
        binding.tvToast.text = message

        @Suppress("DEPRECATION")
        Toast(context.applicationContext).apply {
            view = binding.root
            duration = if (long) Toast.LENGTH_LONG else Toast.LENGTH_SHORT
            setGravity(Gravity.BOTTOM or Gravity.CENTER_HORIZONTAL, 0, dpToPx(context, 96))
            show()
        }
    }

    private fun dpToPx(context: Context, dp: Int): Int {
        return (dp * context.resources.displayMetrics.density).toInt()
    }
}
