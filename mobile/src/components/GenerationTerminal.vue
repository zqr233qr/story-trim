<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps({
  show: Boolean,
  content: String,
  title: String,
  isDarkMode: Boolean
})

const emit = defineEmits(['close'])

// 只显示最后一段内容，模拟“底部打印机”效果
// 这样可以避免滚动逻辑，利用 flex-col-reverse 实现自动置底
const displayContent = computed(() => {
  const len = props.content.length
  // 保持最后 1000 字符，前面截断
  if (len < 1000) return props.content
  return '... [前文已省略] ...\n' + props.content.slice(-1000)
})
</script>

<template>
  <view class="fixed inset-0 z-[100] pointer-events-none transition-all duration-300"
        :class="show ? 'opacity-100' : 'opacity-0 invisible'">
    
    <!-- Backdrop -->
    <view @click.stop="emit('close')" class="absolute inset-0 bg-black/60 backdrop-blur-sm pointer-events-auto"></view>

    <!-- Terminal Panel -->
    <view :class="[
            isDarkMode ? 'bg-stone-950 border-t border-stone-800 text-stone-300' : 'bg-[#1e1e1e] border-t border-stone-700 text-stone-300',
            show ? 'translate-y-0' : 'translate-y-full'
          ]" 
         class="absolute bottom-0 inset-x-0 w-full max-w-3xl mx-auto h-[60vh] rounded-t-3xl pointer-events-auto flex flex-col overflow-hidden shadow-2xl transition-transform duration-500 cubic-bezier(0.16, 1, 0.3, 1)">
      
      <!-- Header -->
      <view class="h-14 flex items-center justify-between px-6 border-b border-white/10 shrink-0 bg-white/5">
        <view class="flex items-center gap-3">
          <view class="w-2 h-2 rounded-full bg-teal-500 animate-pulse"></view>
          <text class="text-xs font-mono font-bold tracking-widest uppercase text-white/70">AI PROCESSING: {{ title }}</text>
        </view>
        <view @click.stop="emit('close')" class="p-2 -mr-2 active:opacity-50">
          <text class="text-xl text-white/50">×</text>
        </view>
      </view>

      <!-- Content Area (Printer Mode) -->
      <view class="flex-1 p-6 font-mono text-sm leading-loose relative overflow-hidden flex flex-col-reverse">
        
        <!-- Text Container -->
        <view class="whitespace-pre-wrap break-words">
          <text>{{ displayContent }}</text>
          <text class="inline-block w-2.5 h-5 bg-teal-500 align-text-bottom ml-1 animate-pulse">▋</text>
        </view>

        <!-- Top Fade Gradient -->
        <view class="absolute top-0 inset-x-0 h-32 bg-gradient-to-b from-[#1e1e1e] to-transparent pointer-events-none"></view>
      </view>
    </view>
  </view>
</template>

<style scoped>
/* 强制深色模式风格，更像终端 */
</style>