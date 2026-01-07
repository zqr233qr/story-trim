<script setup lang="ts">
import { computed } from 'vue'
import type { Book } from '../stores/book'

const props = defineProps<{ book: Book }>()
const emit = defineEmits(['click'])

const statusText = computed(() => {
  const map: Record<string, string> = { 'new': '新书籍', 'processing': '处理中', 'ready': '已就绪' }
  return map[props.book.status] || props.book.status
})
</script>

<template>
  <view @click="emit('click')" class="bg-white p-4 rounded-2xl shadow-sm border border-stone-100 flex gap-4 active:scale-[0.98] transition-transform mb-3">
    <!-- Cover Placeholder -->
    <view class="w-16 h-20 bg-stone-50 rounded-lg flex items-center justify-center border border-stone-100 shrink-0 shadow-inner">
      <text class="text-[10px] text-stone-300 font-serif">BOOK</text>
    </view>

    <view class="flex-1 min-w-0 flex flex-col justify-between py-0.5">
      <view>
        <view class="flex justify-between items-start gap-2">
          <text class="font-bold text-stone-800 truncate text-base">{{ book.title }}</text>
          <view v-if="book.book_trimmed_ids?.length" class="flex items-center gap-0.5 bg-teal-50 px-1.5 py-0.5 rounded border border-teal-100 shrink-0">
            <text class="text-[10px] text-teal-600 font-bold">🪄 AI</text>
          </view>
        </view>
        <text class="text-xs text-stone-400 mt-1 block truncate">共 {{ book.total_chapters || 0 }} 章节</text>
      </view>
      
      <view class="flex items-center justify-between">
        <view class="flex items-center gap-1.5">
          <view :class="{
            'bg-teal-500': book.status === 'ready',
            'bg-yellow-400 animate-pulse': book.status === 'processing',
            'bg-stone-300': book.status === 'new'
          }" class="w-1.5 h-1.5 rounded-full"></view>
          <text class="text-[10px] text-stone-500 font-medium">{{ statusText }}</text>
        </view>
        <!-- Optional: Reading Progress Info -->
      </view>
    </view>
  </view>
</template>
