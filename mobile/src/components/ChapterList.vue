<script setup lang="ts">
import { ref } from 'vue'

interface ModeColors {
  light: { bg: string; text: string };
  dark: { bg: string; text: string };
  sepia: { bg: string; text: string };
}

const props = defineProps<{
  show: boolean,
  chapters: any[],
  activeChapterIndex: number,
  activeModeId?: string,
  readingMode: 'light' | 'dark' | 'sepia',
  modeColors: ModeColors
}>()

const emit = defineEmits(['close', 'select'])

const hasTrimmedContent = (chapter: any) => {
  if (!props.activeModeId || props.activeModeId === 'original') return false
  const targetId = Number(props.activeModeId)
  return chapter.trimmed_prompt_ids?.includes(targetId)
}

const isDarkMode = props.readingMode === 'dark' || props.readingMode === 'sepia'

const panelBg = props.readingMode === 'light' ? '#fafaf9' : '#1c1917'
const panelBorder = props.readingMode === 'light' ? '#e7e5e4' : '#44403c'
const textColor = props.readingMode === 'light' ? '#1c1917' : '#e5e5e5'
const mutedColor = props.readingMode === 'light' ? '#78716c' : '#a8a29e'
const itemBg = props.readingMode === 'light' ? '#f5f5f4' : '#292524'
const activeBg = props.readingMode === 'light' ? '#ccfbf1' : '#134e4a'
const activeText = props.readingMode === 'light' ? '#0f766e' : '#5eead4'
</script>

<template>
  <view class="fixed inset-0 z-[100] flex overflow-hidden pointer-events-none transition-all duration-300"
        :class="show ? 'opacity-100' : 'opacity-0 invisible'">

    <!-- Backdrop -->
    <view @click.stop="emit('close')" class="absolute inset-0 bg-black/40 backdrop-blur-[1px] pointer-events-auto"></view>

    <!-- Drawer -->
    <view :style="{ backgroundColor: panelBg, borderColor: panelBorder }"
          class="relative w-4/5 max-w-[300px] h-full shadow-2xl flex flex-col pointer-events-auto transition-transform duration-500 cubic-bezier border-r"
          :class="show ? 'translate-x-0' : '-translate-x-full'">

      <view class="p-6 border-b shrink-0" :style="{ borderColor: panelBorder }">
        <view class="text-lg font-bold" :style="{ color: textColor }">目录</view>
        <view class="text-[10px] mt-1" :style="{ color: mutedColor }">共 {{ chapters.length }} 章</view>
      </view>

       <scroll-view scroll-y class="flex-1 p-2 min-h-0">
         <view v-for="(chap, index) in chapters" :key="chap.id"
           @click="emit('select', index)"
           :class="[
             activeChapterIndex === index
               ? 'border-l-4 border-teal-500'
               : ''
           ]"
           :style="activeChapterIndex === index
             ? { backgroundColor: activeBg, color: activeText }
             : { backgroundColor: itemBg, color: mutedColor }"
           class="px-4 py-4 text-sm rounded-r-lg mb-1 transition-all flex items-center justify-between">
           <text class="truncate block flex-1">{{ chap.title }}</text>
           <text v-if="hasTrimmedContent(chap)" class="text-[10px] ml-2 opacity-80">🪄</text>
         </view>
       </scroll-view>
    </view>
  </view>
</template>

<style scoped>
.cubic-bezier { transition-timing-function: cubic-bezier(0.16, 1, 0.3, 1); }
</style>
