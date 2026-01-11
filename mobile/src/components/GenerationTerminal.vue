<script setup lang="ts">
import { computed } from 'vue'

interface ModeColors {
  light: { bg: string; text: string };
  dark: { bg: string; text: string };
  sepia: { bg: string; text: string };
}

const props = defineProps({
  show: Boolean,
  content: String,
  title: String,
  readingMode: String as () => 'light' | 'dark' | 'sepia',
  modeColors: Object as () => ModeColors
})

const emit = defineEmits(['close'])

// 只显示最后一段内容，模拟"底部打印机"效果
const displayContent = computed(() => {
  const len = props.content?.length || 0
  if (len < 1000) return props.content
  return '... [前文已省略] ...\n' + props.content?.slice(-1000)
})

// 根据阅读模式调整背景色和边框颜色
const panelBg = computed(() => {
  switch (props.readingMode) {
    case 'sepia': return '#2d2416'
    default: return '#1e1e1e'
  }
})

const borderColor = computed(() => {
  switch (props.readingMode) {
    case 'dark': return '#44403c'
    case 'sepia': return '#d97706'
    default: return '#57534e'
  }
})

const textColor = computed(() => {
  switch (props.readingMode) {
    case 'sepia': return '#d4c4a8'
    default: return '#e5e5e5'
  }
})

const gradientColor = computed(() => {
  switch (props.readingMode) {
    case 'sepia': return 'from-[#2d2416]'
    default: return 'from-[#1e1e1e]'
  }
})
</script>

<template>
  <view class="fixed inset-0 z-[100] pointer-events-none transition-all duration-300"
        :class="show ? 'opacity-100' : 'opacity-0 invisible'">

    <!-- Backdrop -->
    <view @click.stop="emit('close')" class="absolute inset-0 bg-black/60 backdrop-blur-sm pointer-events-auto"></view>

    <!-- Terminal Panel -->
    <view :style="{ backgroundColor: panelBg, borderColor: borderColor }"
         class="absolute bottom-0 inset-x-0 w-full max-w-3xl mx-auto h-[60vh] rounded-t-3xl pointer-events-auto flex flex-col overflow-hidden shadow-2xl transition-transform duration-500 cubic-bezier(0.16, 1, 0.3, 1) border-t border-x"
         :class="show ? 'translate-y-0' : 'translate-y-full'">

      <!-- Header -->
      <view class="h-14 flex items-center justify-between px-6 border-b shrink-0" :style="{ borderColor: borderColor }">
        <view class="flex items-center gap-3">
          <view class="w-2 h-2 rounded-full bg-teal-500 animate-pulse"></view>
          <text class="text-xs font-mono font-bold tracking-widest uppercase" :style="{ color: textColor }">AI PROCESSING: {{ title }}</text>
        </view>
        <view @click.stop="emit('close')" class="p-2 -mr-2 active:opacity-50">
          <text class="text-xl" :style="{ color: textColor }">×</text>
        </view>
      </view>

      <!-- Content Area (Printer Mode) -->
      <view class="flex-1 p-6 font-mono text-sm leading-loose relative overflow-hidden flex flex-col-reverse">

        <!-- Text Container -->
        <view class="whitespace-pre-wrap break-words" :style="{ color: textColor }">
          <text>{{ displayContent }}</text>
          <text class="inline-block w-2.5 h-5 bg-teal-500 align-text-bottom ml-1 animate-pulse">▋</text>
        </view>

        <!-- Top Fade Gradient -->
        <view class="absolute top-0 inset-x-0 h-32 pointer-events-none" :class="`bg-gradient-to-b ${gradientColor} to-transparent`"></view>
      </view>
    </view>
  </view>
</template>

<style scoped>
/* 强制深色模式风格，更像终端 */
</style>
