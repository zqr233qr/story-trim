<script setup lang="ts">
import { ref, watch, computed } from 'vue'

interface ModeColors {
  light: { bg: string; text: string };
  dark: { bg: string; text: string };
  sepia: { bg: string; text: string };
}

interface Prompt {
  id: number;
  name: string;
  description?: string;
}

const props = defineProps<{
  show: boolean,
  bookTitle: string,
  chapterTitle: string,
  prompts: Prompt[],
  trimmedIds?: number[],
  readingMode: 'light' | 'dark' | 'sepia',
  modeColors: ModeColors,
  userPreferredModeId?: number
}>()
const emit = defineEmits(['close', 'start'])

const selectedId = ref<number | string>('')

watch(() => [props.prompts, props.userPreferredModeId], ([newPs, newPrefId]) => {
  if (newPs && newPs.length > 0 && !selectedId.value) {
    // 优先使用用户偏好，否则使用第一个
    selectedId.value = newPrefId ? newPrefId : newPs[0].id
  }
}, { immediate: true })

const panelBg = computed(() => {
  switch (props.readingMode) {
    case 'dark': return '#1c1917'
    case 'sepia': return '#f5e6d3'
    default: return '#ffffff'
  }
})

const panelBorder = computed(() => {
  switch (props.readingMode) {
    case 'dark': return '#44403c'
    case 'sepia': return '#d4c4a8'
    default: return '#e7e5e4'
  }
})

const textColor = computed(() => {
  switch (props.readingMode) {
    case 'dark': return '#e5e5e5'
    case 'sepia': return '#5d4e37'
    default: return '#1c1917'
  }
})

const mutedColor = computed(() => {
  switch (props.readingMode) {
    case 'dark': return '#a8a29e'
    case 'sepia': return '#8b7355'
    default: return '#78716c'
  }
})

const itemBg = computed(() => {
  switch (props.readingMode) {
    case 'dark': return '#292524'
    case 'sepia': return '#fdf8f0'
    default: return '#fafaf9'
  }
})

const selectedBg = computed(() => {
  switch (props.readingMode) {
    case 'dark': return '#134e4a'
    case 'sepia': return '#fef3c7'
    default: return '#f0fdfa'
  }
})

const selectedBorder = computed(() => {
  switch (props.readingMode) {
    case 'dark': return '#0f766e'
    case 'sepia': return '#d97706'
    default: return '#14b8a6'
  }
})

const footerBg = computed(() => panelBg.value)

const footerBorder = computed(() => panelBorder.value)

const buttonPrimaryBg = computed(() => {
  switch (props.readingMode) {
    case 'dark': return '#0f766e'
    case 'sepia': return '#b45309'
    default: return '#0f766e'
  }
})
</script>

<template>
  <view class="fixed inset-0 z-[100] flex items-end sm:items-center justify-center pointer-events-none transition-all duration-300"
        :class="show ? 'opacity-100' : 'opacity-0 invisible'">

    <!-- Backdrop -->
    <view @click.stop="emit('close')" class="absolute inset-0 bg-black/40 backdrop-blur-[1px] pointer-events-auto"></view>

    <!-- Content Card -->
    <view :style="{ backgroundColor: panelBg, borderColor: panelBorder }"
          class="relative w-full max-w-md rounded-2xl overflow-hidden mb-4 sm:mb-0 flex flex-col max-h-[90vh] pointer-events-auto transition-all duration-500 cubic-bezier border"
          :class="show ? 'translate-y-0 scale-100' : 'translate-y-10 scale-95'">

        <view class="p-6 pb-2 shrink-0">
          <view class="text-xl font-bold" :style="{ color: textColor }">AI 精简设置</view>
          <view class="text-sm mt-1" :style="{ color: mutedColor }">《{{ bookTitle }}》{{ chapterTitle }}</view>
        </view>

        <view class="p-6 space-y-3 overflow-y-auto">
          <view v-for="prompt in prompts" :key="prompt.id"
            @click="selectedId = prompt.id"
            :style="selectedId === prompt.id
              ? { backgroundColor: selectedBg, borderColor: selectedBorder }
              : { backgroundColor: itemBg, borderColor: panelBorder }"
            class="flex items-start gap-3 p-3 border rounded-xl cursor-pointer transition-all relative overflow-hidden">

            <view v-if="trimmedIds?.includes(prompt.id)" class="absolute top-0 right-0 bg-teal-500 text-white text-[10px] px-2 py-0.5 rounded-bl-lg font-bold">
              已缓存
            </view>

            <view>
              <view class="font-bold text-sm" :style="{ color: textColor }">{{ prompt.name }}</view>
              <view class="text-xs mt-1 leading-relaxed" :style="{ color: mutedColor }">{{ prompt.description || '暂无描述' }}</view>
            </view>
          </view>
        </view>

        <view :style="{ backgroundColor: footerBg, borderColor: footerBorder }" class="p-4 border-t flex gap-3 shrink-0">
          <view @click="emit('close')" :style="{ backgroundColor: itemBg, color: mutedColor }" class="flex-1 py-3 font-medium text-sm rounded-xl text-center">稍后</view>
          <view @click="emit('start', selectedId)" :style="{ backgroundColor: buttonPrimaryBg }" class="flex-1 py-3 text-white font-medium text-sm rounded-xl shadow-lg text-center active:scale-95 transition-transform">
            {{ trimmedIds?.includes(Number(selectedId)) ? '开始阅读' : '开始精简' }}
          </view>
        </view>
    </view>
  </view>
</template>

<style scoped>
.cubic-bezier { transition-timing-function: cubic-bezier(0.16, 1, 0.3, 1); }
</style>
