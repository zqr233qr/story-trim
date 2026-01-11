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
  prompts: Prompt[],
  readingMode: 'light' | 'dark' | 'sepia',
  modeColors: ModeColors
}>()
const emit = defineEmits(['close', 'confirm'])

const selectedId = ref<number | string>('')

watch(() => props.prompts, (newPs) => {
  if (newPs.length > 0 && !selectedId.value) {
    selectedId.value = newPs[0].id
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

const selectedText = computed(() => {
  switch (props.readingMode) {
    case 'dark': return '#5eead4'
    case 'sepia': return '#b45309'
    default: return '#0f766e'
  }
})

const buttonCancelBg = computed(() => itemBg.value)

const buttonPrimaryBg = computed(() => {
  switch (props.readingMode) {
    case 'dark': return '#0f766e'
    case 'sepia': return '#b45309'
    default: return '#0f766e'
  }
})
</script>

<template>
  <view v-if="show" class="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
    <view :style="{ backgroundColor: panelBg, borderColor: panelBorder }"
          class="w-full max-w-sm rounded-3xl p-6 relative border">
      <view class="text-lg font-bold mb-2 text-center" :style="{ color: textColor }">全书后台处理</view>
      <view class="text-[10px] mb-6 text-center" :style="{ color: mutedColor }">将对《{{ bookTitle }}》进行批量精简。处理将在后台进行，您可以继续阅读。</view>

      <view class="space-y-3 mb-8 max-h-[300px] overflow-y-auto pr-1">
        <view v-for="prompt in prompts" :key="prompt.id"
          @click="selectedId = prompt.id"
          :style="selectedId === prompt.id
            ? { backgroundColor: selectedBg, borderColor: selectedBorder, color: selectedText }
            : { backgroundColor: itemBg, borderColor: panelBorder, color: mutedColor }"
          class="flex items-center gap-3 p-4 border rounded-2xl transition-all">
          <view class="flex-1">
             <view class="text-sm font-bold">{{ prompt.name }}</view>
             <view class="text-[10px] opacity-70 mt-1">{{ prompt.description || '暂无描述' }}</view>
          </view>
        </view>
      </view>

      <view class="flex gap-3">
        <button @click="emit('close')" :style="{ backgroundColor: buttonCancelBg, color: mutedColor }" class="flex-1 h-12 rounded-xl text-sm font-bold flex items-center justify-center">取消</button>
        <button @click="emit('confirm', selectedId)" :style="{ backgroundColor: buttonPrimaryBg }" class="flex-1 h-12 text-white rounded-xl text-sm font-bold shadow-lg flex items-center justify-center">启动任务</button>
      </view>
    </view>
  </view>
</template>
