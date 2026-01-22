<script setup lang="ts">
import { onShow } from "@dcloudio/uni-app";
import { ref } from "vue";
import { db } from "@/utils/sqlite";

const sqlText = ref("SELECT * FROM books LIMIT 20");
const resultText = ref("");
const isRunning = ref(false);

// 初始化数据库连接。
const initDatabase = async () => {
  await db.open();
};

// 执行 SQL 并输出结果。
const executeSql = async () => {
  const sql = sqlText.value.trim();
  if (!sql) {
    resultText.value = "SQL 不能为空";
    return;
  }

  isRunning.value = true;
  try {
    const isQuery = /^\s*(select|pragma|with)\b/i.test(sql);
    if (isQuery) {
      const rows = await db.select<any>(sql);
      resultText.value = JSON.stringify(rows, null, 2);
    } else {
      await db.execute(sql);
      resultText.value = "执行成功";
    }
  } catch (e: any) {
    resultText.value = `执行失败: ${e?.message || e}`;
  } finally {
    isRunning.value = false;
  }
};

onShow(async () => {
  await initDatabase();
});
</script>

<template>
  <view class="min-h-screen bg-stone-50 p-4">
    <view class="bg-white rounded-2xl p-4 shadow-[0_4px_20px_rgba(0,0,0,0.06)]">
      <text class="text-sm font-bold text-stone-800">SQL 调试面板</text>
      <textarea
        v-model="sqlText"
        class="mt-3 w-full h-36 border border-stone-200 rounded-xl p-3 text-xs text-stone-700"
        placeholder="输入 SQL 语句"
      ></textarea>
      <button
        class="mt-3 bg-emerald-600 text-white text-sm rounded-xl py-2"
        :disabled="isRunning"
        @click="executeSql"
      >
        {{ isRunning ? "执行中..." : "执行 SQL" }}
      </button>
    </view>

    <view class="mt-4 bg-white rounded-2xl p-4 shadow-[0_4px_20px_rgba(0,0,0,0.06)]">
      <text class="text-sm font-bold text-stone-800">结果输出</text>
      <scroll-view scroll-y class="mt-3 max-h-96">
        <text class="text-xs text-stone-700 whitespace-pre-wrap">{{ resultText }}</text>
      </scroll-view>
    </view>
  </view>
</template>
