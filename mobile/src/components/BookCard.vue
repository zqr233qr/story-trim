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

const isCloud = computed(() => {
  // sync_state 1 or 2 means synced to cloud
  return props.book.sync_state === 1 || props.book.sync_state === 2;
})

const handleLongPress = () => {
  emit('longpress', props.book)
}

// 封面相关逻辑
import { ref } from 'vue'
const coverError = ref(false)
const localCoverUrl = computed(() => {
  if (!props.book.book_md5) return null
  return `_doc/covers/${props.book.book_md5}.jpg`
})

const onCoverError = () => {
  coverError.value = true
}
</script>

<template>
  <view 
    @click="emit('click')" 
    @longpress="handleLongPress"
    class="bg-white p-4 rounded-3xl shadow-[0_4px_20px_-4px_rgba(0,0,0,0.05)] border border-stone-50 flex gap-5 active:scale-[0.98] transition-all duration-200 mb-4 relative overflow-hidden"
  >
    <!-- Processing Overlay (Subtle) -->
    <view v-if="book.status === 'processing'" class="absolute top-0 right-0 p-3">
       <view class="w-1.5 h-1.5 rounded-full bg-stone-300 animate-pulse"></view>
    </view>

    <!-- Cover Area -->
    <view class="w-16 h-22 shrink-0 relative overflow-hidden rounded-xl">
      <!-- Real Cover (if exists) -->
      <image 
        v-if="!coverError && localCoverUrl"
        :src="localCoverUrl"
        mode="aspectFill"
        class="w-full h-full bg-stone-50"
        @error="onCoverError"
      />
      
      <!-- Placeholder: Minimalist Modern (FallBack) -->
      <view 
        v-else
        class="w-full h-full bg-stone-100 flex flex-col items-center justify-end pb-3 shadow-inner"
      >
        <image src="/static/icons/book-open.svg" class="w-6 h-6 opacity-20 mb-1" />
        <view class="w-8 h-0.5 bg-stone-200 mb-1 rounded-full"></view>
        <view class="w-5 h-0.5 bg-stone-200 rounded-full"></view>
      </view>
    </view>

    <view class="flex-1 min-w-0 flex flex-col justify-between py-1">
      <!-- Top Info -->
      <view>
        <text class="font-bold text-stone-900 truncate text-lg tracking-tight block">{{ book.title }}</text>
        <view class="flex items-center gap-1.5 mt-1">
          <text class="text-[11px] text-stone-400 font-medium tracking-wide">{{ book.total_chapters || 0 }} 章节</text>
        </view>
      </view>
      
      <!-- Bottom Tags Row -->
      <view class="flex items-center flex-wrap gap-2 mt-3">
        <!-- AI Trimmed Badge -->
        <view v-if="book.book_trimmed_ids?.length" class="flex items-center gap-1 bg-stone-100 px-2.5 py-1 rounded-lg">
          <image src="/static/icons/sparkles.svg" class="w-3 h-3 opacity-60" />
          <text class="text-[10px] text-stone-600 font-bold tracking-wide">已精简</text>
        </view>

        <!-- Sync Status Badges -->
        <view v-if="isCloud" class="flex items-center gap-1 bg-stone-50 border border-stone-100 px-2.5 py-1 rounded-lg">
          <image src="/static/icons/cloud.svg" class="w-3 h-3 opacity-60" />
          <text class="text-[10px] text-stone-500 font-bold tracking-wide">云端</text>
        </view>

        <view 
          v-else 
          @click.stop="emit('sync')"
          class="flex items-center gap-1 bg-white border border-stone-200 px-2.5 py-1 rounded-lg active:bg-stone-50 transition-colors"
        >
          <image src="/static/icons/sync.svg" class="w-3 h-3 opacity-60" />
          <text class="text-[10px] text-stone-600 font-bold tracking-wide">同步</text>
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
.h-22 {
  height: 88px;
}
</style>