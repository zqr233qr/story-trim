<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<template>
  <div class="antialiased text-slate-900 bg-[#F9F7F1]">
    <RouterView />
  </div>
</template>

<style>
/* 全局平滑滚动和字体基础设定 */
html, body, #app {
  height: 100%;
  margin: 0;
  padding: 0;
}

body {
  font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
}

/* 隐藏滚动条但保留功能 */
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 10px;
}
</style>
