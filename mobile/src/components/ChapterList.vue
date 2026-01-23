<script setup lang="ts">
import { ref, watch, nextTick, computed, getCurrentInstance } from 'vue'

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
  readIndexes?: number[],
  readingMode: 'light' | 'dark' | 'sepia',
  modeColors: ModeColors
}>()

const emit = defineEmits(['close', 'select'])

const scrollTop = ref(0)
const instance = getCurrentInstance()

const calcScrollTop = () => {
  if (!instance || props.activeChapterIndex < 0) return
  const query = uni.createSelectorQuery().in(instance)
  query.select('.chapter-list-container').boundingClientRect()
  query.select('.chapter-list-container').scrollOffset()
  query.select(`#chapter-${props.activeChapterIndex}`).boundingClientRect()
  query.exec((res) => {
    const container = res?.[0] as { top?: number; height?: number } | undefined
    const scroll = res?.[1] as { scrollTop?: number } | undefined
    const target = res?.[2] as { top?: number } | undefined
    if (!container?.height || target?.top === undefined) return
    const containerTop = container.top || 0
    const currentScroll = scroll?.scrollTop || 0
    const offset = Math.max(0, target.top - containerTop + currentScroll - container.height / 2 + 40)
    scrollTop.value = offset
  })
}

watch(() => [props.show, props.activeChapterIndex], async ([newShow]) => {
  if (newShow) {
    await nextTick()
    setTimeout(() => {
      calcScrollTop()
    }, 0)
  }
})

const isDarkMode = props.readingMode === 'dark' || props.readingMode === 'sepia'

const panelBg = props.readingMode === 'light' ? '#fafaf9' : '#1c1917'
const panelBorder = props.readingMode === 'light' ? '#e7e5e4' : '#44403c'
const textColor = props.readingMode === 'light' ? '#1c1917' : '#e5e5e5'
const mutedColor = props.readingMode === 'light' ? '#78716c' : '#a8a29e'
const itemBg = props.readingMode === 'light' ? '#f5f5f4' : '#292524'
const activeBg = props.readingMode === 'light' ? '#ccfbf1' : '#134e4a'
const activeText = props.readingMode === 'light' ? '#0f766e' : '#5eead4'
const unreadText = textColor
const readText = props.readingMode === 'light' ? '#d6d3d1' : '#57534e'

const readSet = computed(() => new Set(props.readIndexes || []))
const isRead = (index: number) => readSet.value.has(index)
</script>

<template>
  <view class="fixed inset-0 z-[100] flex overflow-hidden pointer-events-none transition-all duration-300"
        :class="show ? 'opacity-100' : 'opacity-0 invisible'">

    <!-- Backdrop -->
    <view @click.stop="emit('close')" class="absolute inset-0 bg-black/40 backdrop-blur-[1px] pointer-events-auto"></view>

    <!-- Drawer -->
    <view :style="{ backgroundColor: panelBg, borderColor: panelBorder }"
          class="relative w-4/5 max-w-[300px] h-full shadow-2xl flex flex-col pointer-events-auto transition-transform duration-500 cubic-bezier border-r overflow-hidden"
          :class="show ? 'translate-x-0' : '-translate-x-full'">

      <view class="p-6 border-b shrink-0" :style="{ borderColor: panelBorder }">
        <view class="text-lg font-bold" :style="{ color: textColor }">目录</view>
        <view class="text-[10px] mt-1" :style="{ color: mutedColor }">共 {{ chapters.length }} 章</view>
      </view>

        <scroll-view scroll-y class="chapter-list-container flex-1 overflow-y-auto px-2 min-h-0"
          :scroll-top="scrollTop"
          :scroll-with-animation="true">
          <view v-for="(chap, index) in chapters" :key="chap.id"
            :id="`chapter-${index}`"
            @click="emit('select', index)"
            :class="[
              activeChapterIndex === index
                ? 'border-l-4 border-teal-500'
                : ''
            ]"
             :style="activeChapterIndex === index
              ? { backgroundColor: activeBg, color: activeText }
              : { backgroundColor: itemBg, color: isRead(index) ? readText : unreadText }"

            class="px-4 py-4 text-sm rounded-r-lg mb-1 transition-all flex items-center overflow-hidden">
            <text class="truncate block flex-1 min-w-0">{{ chap.title }}</text>
          </view>
        </scroll-view>
    </view>
  </view>
</template>

<style scoped>
.cubic-bezier { transition-timing-function: cubic-bezier(0.16, 1, 0.3, 1); }
</style>
