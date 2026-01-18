<script setup lang="ts">
import { ref, computed } from 'vue'

interface Prompt {
  id: number;
  name: string;
}

interface ModeColors {
  light: { bg: string; text: string };
  dark: { bg: string; text: string };
  sepia: { bg: string; text: string };
}

const props = defineProps<{
  show: boolean,
  modes: string[],
  prompts: Prompt[],
  activeMode: string,
  fontSize: number,
  readingMode: 'light' | 'dark' | 'sepia',
  modeColors: ModeColors,
  pageMode: 'scroll' | 'click',
  hideStatusBar: boolean
}>()

const emit = defineEmits(['close', 'update:activeMode', 'update:fontSize', 'update:readingMode', 'update:pageMode', 'update:hideStatusBar', 'addMode'])

const getModeName = (id: string) => {
  if (id === 'original') return '原文'
  const prompt = props.prompts.find(p => p.id.toString() === id)
  return prompt ? prompt.name : id
}

const isDarkMode = computed(() => props.readingMode === 'dark' || props.readingMode === 'sepia')

const panelBg = computed(() => isDarkMode.value ? '#1c1917' : '#ffffff')
const panelBorder = computed(() => isDarkMode.value ? '#44403c' : '#e7e5e4')
const textColor = computed(() => isDarkMode.value ? '#e5e5e5' : '#1c1917')
const mutedColor = computed(() => isDarkMode.value ? '#a8a29e' : '#78716c')
</script>

<template>
  <view class="fixed inset-0 z-[100] flex items-end justify-center pointer-events-none transition-all duration-300"
        :class="show ? 'opacity-100' : 'opacity-0 invisible'">

    <!-- Backdrop -->
    <view @click.stop="emit('close')" class="absolute inset-0 bg-black/40 backdrop-blur-[1px] pointer-events-auto"></view>

    <!-- Panel -->
    <view :style="{ backgroundColor: panelBg, borderColor: panelBorder }"
          class="relative w-full max-w-lg rounded-t-3xl shadow-2xl p-6 pb-10 z-10 pointer-events-auto transition-transform duration-500 cubic-bezier border-t"
          :class="show ? 'translate-y-0' : 'translate-y-full'">

      <view class="w-12 h-1 rounded-full mx-auto mb-6" :style="{ backgroundColor: mutedColor }"></view>

      <!-- Font Size -->
      <view class="mb-8">
        <view class="text-[10px] font-bold uppercase tracking-widest mb-3" :style="{ color: mutedColor }">字号</view>
        <view class="flex items-center justify-between rounded-xl p-3" :style="{ backgroundColor: isDarkMode ? '#44403c' : '#fafaf9' }">
          <text class="text-sm px-2" :style="{ color: mutedColor }">A</text>
          <slider
            :value="fontSize"
            @change="emit('update:fontSize', $event.detail.value)"
            min="14" max="30" step="1"
            class="flex-1 mx-4"
            activeColor="#0d9488"
            block-size="20"
            backgroundColor="#d6d3d1"
          />
          <text class="text-xl px-2 font-bold" :style="{ color: textColor }">A</text>
        </view>
      </view>

      <!-- Page Mode (Temporarily Hidden due to bugs) -->
      <!-- <view class="mb-8">
         <view class="text-[10px] font-bold uppercase tracking-widest mb-3" :style="{ color: mutedColor }">翻页模式</view>
         <view class="rounded-xl p-1 flex" :style="{ backgroundColor: isDarkMode ? '#44403c' : '#fafaf9' }">
           <view @click="emit('update:pageMode', 'scroll')"
             class="flex-1 py-2 rounded-lg text-xs font-bold text-center transition-all"
             :style="pageMode === 'scroll' ? (isDarkMode ? { backgroundColor: '#57534e', color: '#fafaf9' } : { backgroundColor: '#ffffff', color: '#1c1917' }) : { color: mutedColor }">
             滚动
           </view>
           <view @click="emit('update:pageMode', 'click')"
             class="flex-1 py-2 rounded-lg text-xs font-bold text-center transition-all"
             :style="pageMode === 'click' ? (isDarkMode ? { backgroundColor: '#57534e', color: '#fafaf9' } : { backgroundColor: '#ffffff', color: '#1c1917' }) : { color: mutedColor }">
             点击
           </view>
         </view>
      </view> -->

      <!-- Reading Mode (Three Options) -->
      <view class="mb-8">
         <view class="text-[10px] font-bold uppercase tracking-widest mb-3" :style="{ color: mutedColor }">阅读模式</view>
         <view class="rounded-xl p-1 flex" :style="{ backgroundColor: isDarkMode ? '#44403c' : '#fafaf9' }">
           <view @click="emit('update:readingMode', 'light')"
             class="flex-1 py-2 rounded-lg text-xs font-bold text-center transition-all"
             :style="readingMode === 'light' ? (isDarkMode ? { backgroundColor: '#57534e', color: '#fafaf9' } : { backgroundColor: '#ffffff', color: '#1c1917' }) : { color: mutedColor }">
             日间
           </view>
           <view @click="emit('update:readingMode', 'dark')"
             class="flex-1 py-2 rounded-lg text-xs font-bold text-center transition-all"
             :style="readingMode === 'dark' ? (isDarkMode ? { backgroundColor: '#57534e', color: '#fafaf9' } : { backgroundColor: '#ffffff', color: '#1c1917' }) : { color: mutedColor }">
             夜间
           </view>
           <view @click="emit('update:readingMode', 'sepia')"
             class="flex-1 py-2 rounded-lg text-xs font-bold text-center transition-all"
             :style="readingMode === 'sepia' ? (isDarkMode ? { backgroundColor: '#57534e', color: '#fafaf9' } : { backgroundColor: '#ffffff', color: '#1c1917' }) : { color: mutedColor }">
             护眼
           </view>
         </view>
      </view>

      <!-- Hide StatusBar Toggle -->
      <view class="flex items-center justify-between">
         <text class="text-[10px] font-bold uppercase tracking-widest" :style="{ color: mutedColor }">隐藏状态栏</text>
         <switch :checked="hideStatusBar" @change="emit('update:hideStatusBar', $event.detail.value)" color="#0d9488" />
      </view>
    </view>
  </view>
</template>

<style scoped>
.cubic-bezier { transition-timing-function: cubic-bezier(0.16, 1, 0.3, 1); }
</style>
