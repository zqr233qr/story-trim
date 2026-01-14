<script setup lang="ts">
import { computed } from 'vue'
import type { Book } from '../stores/book'

const props = defineProps<{ 
  book: Book & { full_trim_status?: string; full_trim_progress?: number }
  deleting?: boolean
}>()
const emit = defineEmits(['click', 'sync', 'delete', 'longpress'])

const statusText = computed(() => {
  const map: Record<string, string> = { 'new': '新书籍', 'processing': '处理中', 'ready': '已就绪' }
  return map[props.book.status] || props.book.status
})

const trimProgressText = computed(() => {
  if (props.book.full_trim_status === 'running' && props.book.full_trim_progress !== undefined) {
    return `⚡ ${props.book.full_trim_progress}%`
  }
  return null
})

const isCloud = computed(() => {
  // sync_state 1 or 2 means synced to cloud
  return props.book.sync_state === 1 || props.book.sync_state === 2;
})

const handleLongPress = () => {
  emit('longpress', props.book)
}
</script>

<template>
  <view 
    @click="emit('click')" 
    @longpress="handleLongPress"
    class="bg-white p-4 rounded-2xl shadow-sm border border-stone-100 flex gap-4 active:scale-[0.98] transition-transform mb-3 relative overflow-hidden"
  >
    <!-- Processing Overlay (Subtle) -->
    <view v-if="book.status === 'processing'" class="absolute top-0 right-0 p-2">
       <view class="w-2 h-2 rounded-full bg-yellow-400 animate-pulse"></view>
    </view>

    <!-- Cover Placeholder -->
    <view class="w-16 h-20 bg-stone-50 rounded-lg flex items-center justify-center border border-stone-100 shrink-0 shadow-inner">
      <text class="text-[10px] text-stone-300 font-serif">BOOK</text>
    </view>

    <view class="flex-1 min-w-0 flex flex-col justify-between py-0.5">
      <!-- Top Info -->
      <view>
        <text class="font-bold text-stone-800 truncate text-base block">{{ book.title }}</text>
        <text class="text-xs text-stone-400 mt-1 block truncate">共 {{ book.total_chapters || 0 }} 章节</text>
        
        <!-- Trim Progress (if running) -->
        <view v-if="trimProgressText" class="mt-1.5">
          <text class="text-xs font-medium text-amber-600 bg-amber-50 px-2 py-0.5 rounded inline-block">{{ trimProgressText }}</text>
        </view>
      </view>
      
      <!-- Bottom Tags Row -->
      <view class="flex items-center flex-wrap gap-2 mt-2">
        <!-- AI Trimmed Badge -->
        <view v-if="book.book_trimmed_ids?.length" class="flex items-center gap-0.5 bg-teal-50 px-2 py-1 rounded-lg border border-teal-100">
          <text class="text-[10px] text-teal-600 font-bold">🪄 AI 精简</text>
        </view>

        <!-- Sync Status Badges -->
        <view v-if="isCloud" class="flex items-center gap-1 bg-blue-50 px-2 py-1 rounded-lg border border-blue-100">
          <text class="text-[10px]">☁️</text>
          <text class="text-[10px] text-blue-600 font-bold">云端</text>
        </view>

        <view 
          v-else 
          @click.stop="emit('sync')"
          class="flex items-center gap-1 bg-stone-100 hover:bg-stone-200 active:bg-stone-300 px-2 py-1 rounded-lg transition-colors border border-stone-200"
        >
          <text class="text-[10px]">⬆️</text>
          <text class="text-[10px] text-stone-600 font-bold">同步</text>
        </view>
      </view>
    </view>
  </view>
</template>