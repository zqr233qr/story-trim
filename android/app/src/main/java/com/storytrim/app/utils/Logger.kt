package com.storytrim.app.utils

import android.util.Log

object Logger {
    private const val TAG = "StoryTrim"

    fun d(message: String, tag: String = TAG) = Log.d(tag, message)
    fun e(message: String, tag: String = TAG) = Log.e(tag, message)
    fun i(message: String, tag: String = TAG) = Log.i(tag, message)
    fun w(message: String, tag: String = TAG) = Log.w(tag, message)
}