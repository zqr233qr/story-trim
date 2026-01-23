<script setup lang="ts">
import { computed, ref, watch } from 'vue'

// 精简模式定义。
interface Prompt {
  id: number
  name: string
  description?: string
}

// 章节精简选项。
interface ChapterTrimOption {
  id: number
  index: number
  title: string
  status: 'available' | 'trimmed' | 'processing'
}

const props = defineProps<{
  show: boolean
  prompts: Prompt[]
  chapters: ChapterTrimOption[]
  balance: number
  currentChapterId: number
  preferredModeId?: number
  readingMode: 'light' | 'dark' | 'sepia'
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'confirm', payload: { promptId: number; chapterIds: number[] }): void
  (e: 'change-prompt', promptId: number): void
}>()

const selectedPromptId = ref<number | string>('')
const selectedChapterIds = ref<number[]>([])
const scrollIntoViewId = ref('')

// 解析默认精简模式。
const resolveDefaultPromptId = () => {
  const preferredId = props.preferredModeId
  if (preferredId && props.prompts.some((p) => p.id === preferredId)) {
    return preferredId
  }
  return props.prompts[0]?.id || ''
}

// 初始化默认精简模式。
watch(() => [props.show, props.prompts, props.preferredModeId], ([visible]) => {
  if (!visible) return
  selectedPromptId.value = resolveDefaultPromptId()
}, { immediate: true })

// 监听显示与模式切换，通知父组件获取状态。
watch([selectedPromptId, () => props.show], ([promptId, visible]) => {
  if (!visible || !promptId) return
  emit('change-prompt', Number(promptId))
  selectedChapterIds.value = []
})

// 章节数据变化后清理不可选项。
watch(() => props.chapters, (newList) => {
  const availableIds = new Set(newList.filter(item => item.status === 'available').map(item => item.id))
  selectedChapterIds.value = selectedChapterIds.value.filter(id => availableIds.has(id))
})

// 弹窗打开时自动定位到当前章节。
watch([
  () => props.show,
  () => props.currentChapterId,
  () => props.chapters
], ([visible, currentId, chapters]) => {
  if (!visible || chapters.length === 0) return
  const target = chapters.find(item => item.id === currentId) || chapters[0]
  if (!target) return
  scrollIntoViewId.value = ''
  scrollIntoViewId.value = `chapter-item-${target.id}`
}, { immediate: true })

const selectedCount = computed(() => selectedChapterIds.value.length)
const estimatedCost = computed(() => selectedCount.value)
const hasEnoughBalance = computed(() => estimatedCost.value <= props.balance)
const confirmDisabled = computed(() => selectedCount.value === 0 || !hasEnoughBalance.value)


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

const disabledText = computed(() => {
  switch (props.readingMode) {
    case 'dark': return '#57534e'
    case 'sepia': return '#a0947a'
    default: return '#a8a29e'
  }
})

const buttonPrimaryBg = computed(() => {
  switch (props.readingMode) {
    case 'dark': return '#0f766e'
    case 'sepia': return '#b45309'
    default: return '#0f766e'
  }
})

const quickSelectOptions = [3, 5, 10, 20, 50]

// 切换章节选中状态。
const toggleChapter = (item: ChapterTrimOption) => {
  if (item.status !== 'available') return
  if (selectedChapterIds.value.includes(item.id)) {
    selectedChapterIds.value = selectedChapterIds.value.filter(id => id !== item.id)
    return
  }
  selectedChapterIds.value = [...selectedChapterIds.value, item.id]
}

// 生成章节状态文本。
const getStatusLabel = (item: ChapterTrimOption) => {
  if (item.status === 'processing') return '处理中'
  if (item.status === 'trimmed') return '已精简'
  return '可选择'
}

// 获取章节行样式。
const getChapterStyle = (item: ChapterTrimOption) => {
  if (item.status !== 'available') {
    return { backgroundColor: itemBg.value, borderColor: panelBorder.value, color: disabledText.value }
  }
  if (selectedChapterIds.value.includes(item.id)) {
    return { backgroundColor: selectedBg.value, borderColor: selectedBorder.value, color: selectedText.value }
  }
  return { backgroundColor: itemBg.value, borderColor: panelBorder.value, color: textColor.value }
}

// 快速选择后续章节。
const handleQuickSelect = (count: number) => {
  const chapters = props.chapters
  if (chapters.length === 0) return
  const currentIndex = chapters.findIndex(item => item.id === props.currentChapterId)
  const startIndex = currentIndex >= 0 ? currentIndex + 1 : 0
  const candidates = chapters.slice(startIndex).filter(item => item.status === 'available')
  selectedChapterIds.value = candidates.slice(0, count).map(item => item.id)
}

// 点击提交。
const handleConfirm = () => {
  if (confirmDisabled.value) return
  emit('confirm', { promptId: Number(selectedPromptId.value), chapterIds: [...selectedChapterIds.value] })
}
</script>

<template>
  <view class="fixed inset-0 z-[100] flex items-end sm:items-center justify-center pointer-events-none transition-all duration-300"
        :class="show ? 'opacity-100' : 'opacity-0 invisible'">
    <view @click.stop="emit('close')" class="absolute inset-0 bg-black/40 backdrop-blur-[1px] pointer-events-auto"></view>

    <view :style="{ backgroundColor: panelBg, borderColor: panelBorder }"
          class="relative w-full max-w-md rounded-2xl overflow-hidden mb-4 sm:mb-0 flex flex-col max-h-[90vh] pointer-events-auto transition-all duration-500 cubic-bezier border"
          :class="show ? 'translate-y-0 scale-100' : 'translate-y-10 scale-95'">
      <view class="p-6 pb-3 shrink-0">
        <view class="flex items-center justify-between">
          <view class="text-xl font-bold" :style="{ color: textColor }">指定章节精简</view>
          <view class="text-[10px] px-2 py-1 rounded-full" :style="{ backgroundColor: itemBg, color: mutedColor }">积分 {{ balance }}</view>
        </view>
      </view>

      <view class="px-6 pb-4">
        <view class="text-xs font-semibold" :style="{ color: mutedColor }">选择精简模式</view>
        <view class="mt-3 grid grid-cols-2 gap-2 max-h-[160px] overflow-y-auto">
          <view v-for="prompt in prompts" :key="prompt.id"
                @click="selectedPromptId = prompt.id"
                :style="selectedPromptId === prompt.id
                  ? { backgroundColor: selectedBg, borderColor: selectedBorder, color: selectedText }
                  : { backgroundColor: itemBg, borderColor: panelBorder, color: mutedColor }"
                class="flex items-center justify-center p-3 border rounded-xl transition-all">
            <view class="text-sm font-bold truncate max-w-full">{{ prompt.name }}</view>
          </view>
        </view>
      </view>

      <view class="px-6 pb-4 flex-1 overflow-hidden">
        <view class="text-xs font-semibold" :style="{ color: mutedColor }">选择章节</view>
        <scroll-view
          scroll-y
          :scroll-into-view="scrollIntoViewId"
          class="mt-3 h-[260px]"
        >
          <view class="grid grid-cols-2 gap-2">
          <view v-for="item in chapters" :key="item.id"
                :id="`chapter-item-${item.id}`"
                @click="toggleChapter(item)"
                :style="getChapterStyle(item)"
                class="relative flex items-center gap-3 p-3 border rounded-xl transition-all">
            <view v-if="item.id === currentChapterId" class="absolute top-2 left-2 text-[9px] px-1.5 py-0.5 rounded-full" :style="{ backgroundColor: selectedBg, color: selectedText }">
              当前
            </view>
            <view class="w-5 h-5 rounded-full border flex items-center justify-center"
                  :style="selectedChapterIds.includes(item.id)
                    ? { borderColor: selectedBorder, backgroundColor: selectedBg }
                    : { borderColor: panelBorder, backgroundColor: 'transparent' }">
              <view v-if="selectedChapterIds.includes(item.id)" class="w-2.5 h-2.5 rounded-full" :style="{ backgroundColor: selectedBorder }"></view>
            </view>
            <view class="flex-1 min-w-0">
              <view class="text-sm font-semibold truncate">{{ item.index }}. {{ item.title }}</view>
              <view class="text-[10px] mt-1" :style="{ color: item.status === 'available' ? mutedColor : disabledText }">
                {{ getStatusLabel(item) }}
              </view>
            </view>
          </view>
          <view v-if="chapters.length === 0" class="text-xs text-center py-6 col-span-2" :style="{ color: mutedColor }">暂无可选章节</view>
        </view>
      </scroll-view>
      <view class="mt-3 flex flex-wrap gap-2">
        <view v-for="count in quickSelectOptions" :key="count"
              @click="handleQuickSelect(count)"
              class="px-2.5 py-1 rounded-full text-[10px] font-semibold border"
              :style="{ backgroundColor: itemBg, borderColor: panelBorder, color: mutedColor }">
          后 {{ count }} 章
        </view>
      </view>
    </view>


      <view :style="{ backgroundColor: panelBg, borderColor: panelBorder }" class="p-4 border-t flex items-center gap-3 shrink-0">
        <view class="flex-1 text-xs" :style="{ color: mutedColor }">
          已选 {{ selectedCount }} 章 · 预计消耗 {{ estimatedCost }} 积分
        </view>
        <view v-if="!hasEnoughBalance" class="text-[10px] text-red-500">余额不足</view>
        <view @click="handleConfirm" :style="{ backgroundColor: confirmDisabled ? itemBg : buttonPrimaryBg, color: confirmDisabled ? mutedColor : '#ffffff' }"
              class="px-4 py-2 rounded-xl text-sm font-semibold shadow-lg">
          开始精简
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
.cubic-bezier { transition-timing-function: cubic-bezier(0.16, 1, 0.3, 1); }
</style>
