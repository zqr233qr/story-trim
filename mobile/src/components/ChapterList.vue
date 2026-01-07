<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  show: boolean,
  chapters: any[],
  activeChapterIndex: number,
  isDarkMode?: boolean
}>()

const emit = defineEmits(['close', 'select'])
</script>

<template>
  <view class="fixed inset-0 z-[100] flex overflow-hidden pointer-events-none transition-all duration-300"
        :class="show ? 'opacity-100' : 'opacity-0 invisible'">
    
    <!-- Backdrop -->
    <view @click.stop="emit('close')" class="absolute inset-0 bg-black/40 backdrop-blur-[1px] pointer-events-auto"></view>

    <!-- Drawer -->
    <view :class="[
            isDarkMode ? 'bg-stone-900 border-r border-stone-800' : 'bg-[#fafaf9]',
            show ? 'translate-x-0' : '-translate-x-full'
          ]"
          class="relative w-4/5 max-w-[300px] h-full shadow-2xl flex flex-col pointer-events-auto transition-transform duration-500 cubic-bezier">
      
      <view class="p-6 border-b shrink-0" :class="isDarkMode ? 'border-stone-800' : 'border-stone-200'">
        <view class="text-lg font-bold" :class="isDarkMode ? 'text-stone-100' : 'text-stone-800'">目录</view>
        <view class="text-[10px] mt-1 text-stone-400">共 {{ chapters.length }} 章</view>
      </view>

      <scroll-view scroll-y class="flex-1 p-2">
        <view v-for="(chap, index) in chapters" :key="chap.id" 
          @click="emit('select', index)"
          :class="[
            activeChapterIndex === index 
              ? (isDarkMode ? 'bg-teal-900/20 text-teal-400 border-l-4 border-teal-500' : 'bg-teal-50 text-teal-700 border-l-4 border-teal-500') 
              : (isDarkMode ? 'text-stone-400' : 'text-stone-600')
          ]"
          class="px-4 py-4 text-sm rounded-r-lg mb-1 transition-all">
          <text class="truncate block">{{ chap.title }}</text>
        </view>
      </scroll-view>
    </view>
  </view>
</template>

<style scoped>
.cubic-bezier { transition-timing-function: cubic-bezier(0.16, 1, 0.3, 1); }
</style>
