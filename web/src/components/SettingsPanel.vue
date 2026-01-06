<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  show: boolean,
  modes: string[],
  activeMode: string,
  fontSize: number,
  isDarkMode: boolean,
  pageMode: 'scroll' | 'click'
}>()

const emit = defineEmits(['close', 'update:activeMode', 'update:fontSize', 'update:isDarkMode', 'update:pageMode', 'addMode'])

const modeNames: Record<string, string> = {
  'original': '原文',
  'dewater': '标准沉浸',
  'summary': '轻度精简',
  'speed': '极简速读'
}
</script>

<template>
  <transition name="slide-up">
    <div v-if="show" class="fixed inset-0 z-50 flex items-end justify-center">
      <!-- Backdrop -->
      <div @click="emit('close')" class="absolute inset-0 bg-black/20 backdrop-blur-[1px] transition-opacity"></div>

      <!-- Panel -->
      <div class="relative bg-white w-full max-w-lg rounded-t-2xl shadow-xl p-6 pb-8 z-10 safe-bottom">
        <div class="w-12 h-1 bg-stone-200 rounded-full mx-auto mb-6"></div>

        <!-- Mode Switcher -->
        <div class="mb-8">
          <div class="text-xs font-bold text-stone-400 uppercase tracking-wider mb-3 flex justify-between">
            <span>AI 阅读层</span>
            <span class="text-teal-600 font-normal cursor-pointer text-[10px] hover:underline" @click="emit('addMode')">+ 新增处理</span>
          </div>

          <div v-if="modes.length > 0" class="flex gap-3 overflow-x-auto no-scrollbar pb-2">
            <button v-for="modeKey in modes" :key="modeKey"
              @click="emit('update:activeMode', modeKey)"
              :class="activeMode === modeKey ? 'bg-teal-500 text-white border-transparent shadow-md shadow-teal-200' : 'bg-white text-stone-600 border-stone-200 hover:border-teal-300'"
              class="px-4 py-2 rounded-full text-xs font-medium border flex-shrink-0 transition-all">
              {{ modeNames[modeKey] || modeKey }}
            </button>
          </div>
          <div v-else class="text-sm text-stone-400 italic bg-stone-50 p-3 rounded-lg text-center">当前书籍尚未进行 AI 处理</div>
        </div>

        <!-- Font Size -->
        <div class="mb-8">
          <div class="text-xs font-bold text-stone-400 uppercase tracking-wider mb-3">字号</div>
          <div class="flex items-center justify-between bg-stone-50 rounded-xl p-3">
            <span class="text-sm font-serif px-2 text-stone-500">A</span>
            <input 
              :value="fontSize" 
              @input="emit('update:fontSize', parseInt(($event.target as HTMLInputElement).value))"
              type="range" min="14" max="24" step="1" 
              class="w-full mx-4 h-1 bg-stone-200 rounded-lg appearance-none cursor-pointer accent-teal-600"
            >
            <span class="text-xl font-serif px-2 text-stone-800">A</span>
          </div>
        </div>

        <!-- Page Mode -->
        <div class="mb-8">
           <div class="text-xs font-bold text-stone-400 uppercase tracking-wider mb-3">翻页模式</div>
           <div class="bg-stone-50 rounded-xl p-1 flex">
             <button @click="emit('update:pageMode', 'scroll')" 
               :class="pageMode === 'scroll' ? 'bg-white shadow-sm text-stone-800' : 'text-stone-400 hover:text-stone-600'"
               class="flex-1 py-2 rounded-lg text-xs font-bold transition-all">
               滚动
             </button>
             <button @click="emit('update:pageMode', 'click')" 
               :class="pageMode === 'click' ? 'bg-white shadow-sm text-stone-800' : 'text-stone-400 hover:text-stone-600'"
               class="flex-1 py-2 rounded-lg text-xs font-bold transition-all">
               点击
             </button>
           </div>
        </div>

        <!-- Dark Mode -->
        <div class="mb-8 flex items-center justify-between">
           <span class="text-xs font-bold text-stone-400 uppercase tracking-wider">夜间模式</span>
           <button 
             @click="emit('update:isDarkMode', !isDarkMode)"
             :class="isDarkMode ? 'bg-stone-800' : 'bg-stone-200'"
             class="w-12 h-6 rounded-full relative transition-colors duration-300 focus:outline-none">
             <div 
               :class="isDarkMode ? 'translate-x-6 bg-stone-400' : 'translate-x-1 bg-white'"
               class="w-4 h-4 rounded-full absolute top-1 transition-transform duration-300 shadow-sm"></div>
           </button>
        </div>
      </div>
    </div>
  </transition>
</template>

<style scoped>
.slide-up-enter-active, .slide-up-leave-active { transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1); }
.slide-up-enter-from, .slide-up-leave-to { transform: translateY(100%); }
.safe-bottom { padding-bottom: env(safe-area-inset-bottom); }
</style>
