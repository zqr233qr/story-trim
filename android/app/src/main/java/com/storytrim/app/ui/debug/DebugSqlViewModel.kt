package com.storytrim.app.ui.debug

import android.database.Cursor
import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.storytrim.app.core.database.AppDatabase
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import org.json.JSONArray
import org.json.JSONObject
import javax.inject.Inject

@HiltViewModel
class DebugSqlViewModel @Inject constructor(
    private val appDatabase: AppDatabase
) : ViewModel() {

    private val _result = MutableLiveData<String>()
    val result: LiveData<String> = _result

    private val _isLoading = MutableLiveData<Boolean>()
    val isLoading: LiveData<Boolean> = _isLoading

    fun executeSql(sql: String) {
        if (sql.isBlank()) {
            _result.value = "错误: SQL 语句不能为空"
            return
        }

        _isLoading.value = true
        viewModelScope.launch(Dispatchers.IO) {
            try {
                val db = appDatabase.openHelper.writableDatabase
                val isQuery = sql.trim().split(" ")[0].uppercase() in listOf("SELECT", "PRAGMA", "WITH")

                if (isQuery) {
                    val cursor = db.query(sql)
                    val jsonResult = cursorToJson(cursor)
                    _result.postValue(jsonResult)
                } else {
                    db.execSQL(sql)
                    _result.postValue("执行成功")
                }
            } catch (e: Exception) {
                _result.postValue("执行失败: ${e.message}")
            } finally {
                _isLoading.postValue(false)
            }
        }
    }

    private fun cursorToJson(cursor: Cursor): String {
        val resultSet = JSONArray()
        cursor.use {
            while (it.moveToNext()) {
                val row = JSONObject()
                for (i in 0 until it.columnCount) {
                    val columnName = it.getColumnName(i)
                    when (it.getType(i)) {
                        Cursor.FIELD_TYPE_NULL -> row.put(columnName, JSONObject.NULL)
                        Cursor.FIELD_TYPE_INTEGER -> row.put(columnName, it.getLong(i))
                        Cursor.FIELD_TYPE_FLOAT -> row.put(columnName, it.getDouble(i))
                        Cursor.FIELD_TYPE_STRING -> row.put(columnName, it.getString(i))
                        Cursor.FIELD_TYPE_BLOB -> row.put(columnName, "[BLOB]")
                    }
                }
                resultSet.put(row)
            }
        }
        return if (resultSet.length() > 0) {
            resultSet.toString(2) // 格式化 JSON
        } else {
            "[] (查询结果为空)"
        }
    }
}
