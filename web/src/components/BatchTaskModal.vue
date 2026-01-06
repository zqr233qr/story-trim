<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{ 
  show: boolean, 
  bookTitle: string,
  isDarkMode?: boolean
}>()
const emit = defineEmits(['close', 'confirm'])

const modes = [
  { id: 'dewater', name: '标准沉浸', icon: '💧', desc: '大幅删减无意义的重复描写、心理独白。' },
  { id: 'summary', name: '轻度精简', icon: '🍃', desc: '仅优化语感、合并琐碎短句。' },
  { id: 'speed', name: '极简速读', icon: '⚡', desc: '剧情优先。大胆删除所有环境与心理描写。' }
]

const selectedId = ref('dewater')
</script>

<template>
  <transition name="fade">
    <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
      <div :class="isDarkMode ? 'bg-stone-900 border border-stone-800' : 'bg-white shadow-xl'" class="w-full max-w-sm rounded-2xl p-6 relative">
        <h3 class="text-lg font-bold mb-2" :class="isDarkMode ? 'text-stone-100' : 'text-stone-800'">全书后台处理</h3>
        <p class="text-xs mb-4" :class="isDarkMode ? 'text-stone-500' : 'text-stone-500'">将对《{{ bookTitle }}》剩余章节进行批量精简。处理将在后台进行，您可以继续阅读。</p>

        <div class="space-y-2 mb-6">
          <div v-for="mode in modes" :key="mode.id"
            @click="selectedId = mode.id"
            :class="selectedId === mode.id 
              ? (isDarkMode ? 'border-teal-600 bg-teal-900/20 text-teal-500' : 'border-teal-500 bg-teal-50 text-teal-700') 
              : (isDarkMode ? 'border-stone-800 bg-stone-900/50 text-stone-500 hover:border-stone-700' : 'border-stone-200 text-stone-600')"
            class="flex items-center gap-3 p-3 border rounded-lg cursor-pointer transition-colors">
            <span class="text-lg">{{ mode.icon }}</span>
            <span class="text-sm font-bold">{{ mode.name }}</span>
          </div>
        </div>

        <div class="flex gap-3">
          <button @click="emit('close')" :class="isDarkMode ? 'bg-stone-800 text-stone-400 hover:bg-stone-700' : 'bg-stone-100 text-stone-500 hover:bg-stone-200'" class="flex-1 py-2.5 font-medium text-sm rounded-lg transition-colors">取消</button>
          <button @click="emit('confirm', selectedId)" class="flex-1 py-2.5 bg-stone-900 text-white font-medium text-sm rounded-lg shadow-md hover:bg-teal-600 flex items-center justify-center gap-2 active:scale-[0.98]">
            <span>启动任务</span>
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
          </button>
        </div>
      </div>
    </div>
  </transition>
</template>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
