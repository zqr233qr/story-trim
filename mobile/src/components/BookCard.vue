<script setup lang="ts">
import { computed } from 'vue'
import type { Book } from '../stores/book'

const props = defineProps<{ book: Book }>()
const emit = defineEmits(['click'])

const statusText = computed(() => {
  const map: Record<string, string> = { 'new': '未处理', 'processing': '处理中', 'ready': '已精简' }
  return map[props.book.status] || props.book.status
})
</script>

<template>
  <div @click="emit('click')" class="bg-white p-4 rounded-xl shadow-sm border border-stone-100 flex items-center gap-4 active:scale-[0.98] transition-transform cursor-pointer hover:shadow-md">
    <!-- Cover Placeholder -->
    <div class="w-12 h-16 bg-stone-100 rounded flex items-center justify-center text-[10px] text-stone-300 font-serif shrink-0">
      封面
    </div>

    <div class="flex-1 min-w-0">
      <div class="flex justify-between items-start">
        <h4 class="font-bold text-stone-800 truncate text-sm sm:text-base">{{ book.title }}</h4>
        
        <!-- Status Badge -->
        <div class="flex items-center gap-1.5 bg-stone-50 px-2 py-1 rounded-full border border-stone-100 shrink-0">
          <div :class="{
            'bg-teal-500': book.status === 'ready',
            'bg-yellow-400 animate-pulse': book.status === 'processing',
            'bg-stone-300': book.status === 'new'
          }" class="w-1.5 h-1.5 rounded-full"></div>
          <span class="text-[10px] text-stone-500 font-medium">{{ statusText }}</span>
        </div>
      </div>
      
      <p class="text-xs text-stone-400 mt-1 truncate">{{ book.lastChapter }}</p>
      
      <!-- Progress Bar -->
      <div class="w-full bg-stone-100 h-1 mt-3 rounded-full overflow-hidden">
        <div class="bg-stone-300 h-full rounded-full transition-all duration-500" :style="{ width: book.progress + '%' }"></div>
      </div>
    </div>
  </div>
</template>