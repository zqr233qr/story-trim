<script setup lang="ts">
import type { Chapter } from '../stores/book'

const props = defineProps<{
  show: boolean,
  chapters: Chapter[],
  activeChapterIndex: number,
  isDarkMode?: boolean
}>()

const emit = defineEmits(['close', 'select'])
</script>

<template>
  <div>
    <!-- Backdrop -->
    <transition name="fade">
      <div v-if="show" @click="emit('close')" class="fixed inset-0 bg-black/50 z-40 backdrop-blur-sm"></div>
    </transition>

    <!-- Drawer -->
    <transition name="slide-right">
      <div v-if="show" 
        :class="isDarkMode ? 'bg-stone-900 border-r border-stone-800' : 'bg-[#fafaf9]'"
        class="fixed inset-y-0 left-0 w-4/5 max-w-xs z-50 shadow-2xl flex flex-col transition-colors">
        
        <div class="p-6 border-b shrink-0" :class="isDarkMode ? 'bg-stone-900 border-stone-800' : 'bg-white border-stone-200'">
          <h2 class="text-lg font-bold" :class="isDarkMode ? 'text-stone-100' : 'text-stone-800'">目录</h2>
          <p class="text-xs mt-1" :class="isDarkMode ? 'text-stone-500' : 'text-stone-400'">共 {{ chapters.length }} 章</p>
        </div>

        <div class="flex-1 overflow-y-auto p-2">
          <div v-for="(chap, index) in chapters" :key="chap.id" 
            @click="emit('select', index)"
            :class="[
              activeChapterIndex === index 
                ? (isDarkMode ? 'bg-teal-900/20 text-teal-400 font-bold border-l-4 border-teal-500' : 'bg-teal-50 text-teal-700 font-bold border-l-4 border-teal-500') 
                : (isDarkMode ? 'text-stone-400 hover:bg-stone-800 border-l-4 border-transparent' : 'text-stone-600 hover:bg-stone-100 border-l-4 border-transparent')
            ]"
            class="px-4 py-3.5 text-sm rounded-r-lg cursor-pointer transition-colors mb-1">
            {{ chap.title }}
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.3s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

.slide-right-enter-active, .slide-right-leave-active { transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1); }
.slide-right-enter-from, .slide-right-leave-to { transform: translateX(-100%); }
</style>
