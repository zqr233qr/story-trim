<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{ 
  show: boolean, 
  bookTitle: string,
  isDarkMode?: boolean
}>()
const emit = defineEmits(['close', 'start'])

const modes = [
  { id: 'dewater', name: '标准沉浸', icon: '💧', desc: '大幅删减无意义的重复描写、心理独白和环境堆砌。保留核心对话与伏笔。' },
  { id: 'summary', name: '轻度精简', icon: '🍃', desc: '仅优化语感、合并琐碎短句。全量保留对话和环境渲染，适合细读。' },
  { id: 'speed', name: '极简速读', icon: '⚡', desc: '剧情优先。大胆删除所有环境与心理描写，紧凑叙事，快速通关。' }
]

const selectedId = ref('dewater')
</script>

<template>
  <transition name="slide-up">
    <div v-if="show" class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/30 backdrop-blur-sm p-4">
      <div @click.stop :class="isDarkMode ? 'bg-stone-900 border border-stone-800 shadow-none' : 'bg-white shadow-2xl'" class="w-full max-w-md rounded-2xl overflow-hidden mb-4 sm:mb-0 flex flex-col max-h-[90vh]">
        <div class="p-6 pb-2 shrink-0">
          <h3 class="text-xl font-bold" :class="isDarkMode ? 'text-stone-100' : 'text-stone-800'">AI 精简设置</h3>
          <p class="text-sm mt-1" :class="isDarkMode ? 'text-stone-500' : 'text-stone-500'">为《{{ bookTitle }}》选择精简策略</p>
        </div>

        <div class="p-6 space-y-3 overflow-y-auto">
          <div v-for="mode in modes" :key="mode.id"
            @click="selectedId = mode.id"
            :class="selectedId === mode.id 
              ? (isDarkMode ? 'border-teal-600 bg-teal-900/20 ring-1 ring-teal-600' : 'border-teal-500 bg-teal-50 ring-1 ring-teal-500') 
              : (isDarkMode ? 'border-stone-800 bg-stone-900/50 hover:border-stone-700' : 'border-stone-200 hover:border-teal-200')"
            class="flex items-start gap-3 p-3 border rounded-xl cursor-pointer transition-all">
            
            <div :class="selectedId === mode.id ? 'bg-teal-100 text-teal-600' : 'bg-stone-100 text-stone-400'" class="p-2 rounded-lg shrink-0 text-xl flex items-center justify-center w-10 h-10">
              {{ mode.icon }}
            </div>

            <div>
              <div class="font-bold text-sm" :class="isDarkMode ? 'text-stone-200' : 'text-stone-800'">{{ mode.name }}</div>
              <div class="text-xs mt-1 leading-relaxed" :class="isDarkMode ? 'text-stone-500' : 'text-stone-500'">{{ mode.desc }}</div>
            </div>
          </div>
        </div>

        <div :class="isDarkMode ? 'bg-stone-900 border-stone-800' : 'bg-stone-50 border-stone-100'" class="p-4 border-t flex gap-3 shrink-0">
          <button @click="emit('close')" :class="isDarkMode ? 'text-stone-500 hover:bg-stone-800' : 'text-stone-500 hover:bg-stone-200'" class="flex-1 py-3 font-medium text-sm rounded-xl transition-colors">稍后</button>
          <button @click="emit('start', selectedId)" class="flex-1 py-3 bg-stone-900 text-white font-medium text-sm rounded-xl shadow-lg hover:bg-teal-600 transition-colors flex items-center justify-center gap-2 active:scale-[0.98]">
            <span>开始精简</span>
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
          </button>
        </div>
      </div>
    </div>
  </transition>
</template>

<style scoped>
.slide-up-enter-active, .slide-up-leave-active { transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.3s ease; }
.slide-up-enter-from, .slide-up-leave-to { transform: translateY(100%); opacity: 0; }
</style>
