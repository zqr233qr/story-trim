<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'
import { useBookStore } from '../stores/book'
import { 
  ArrowLeft, Type, Download, CheckCircle2, 
  Loader2, Play, Pause, Scissors, ThumbsUp, RotateCcw,
  BookOpen, X, Layout, Columns
} from 'lucide-vue-next'

const router = useRouter()
const bookStore = useBookStore()

// 1. 基础状态
const chapters = computed(() => bookStore.chapters)
const fileName = computed(() => bookStore.bookTitle)
const selectedIndex = ref<number>(0)
const isTrimming = ref(false)
const isBatchProcessing = ref(false)
const shouldStopBatch = ref(false)
const viewMode = ref<'split' | 'trimmed' | 'original'>('trimmed')
const showSidebar = ref(false)
const readingMode = ref<'scroll' | 'page'>('scroll')
const showControls = ref(true)
const currentPage = ref(0)
const totalPages = ref(1)
const trimmedCache = ref<Record<number, string>>({})
const streamingContent = ref('')

const columnGap = 48 // 固定的列间距

// 2. 元素引用
const trimmedContainer = ref<HTMLElement | null>(null)
const pageContentRef = ref<HTMLElement | null>(null)

// 3. 计算属性
const selectedChapter = computed(() => 
  selectedIndex.value >= 0 ? chapters.value[selectedIndex.value] : null
)

const currentDisplayContent = computed(() => {
  if (trimmedCache.value[selectedIndex.value]) return trimmedCache.value[selectedIndex.value]
  if (isTrimming.value && !trimmedCache.value[selectedIndex.value]) return streamingContent.value
  return ''
})

// 4. 辅助函数
const calculatePages = async () => {
  if (readingMode.value !== 'page') return
  
  await nextTick()
  await nextTick()
  
  if (pageContentRef.value) {
    const el = pageContentRef.value
    const clientWidth = el.clientWidth
    const scrollWidth = el.scrollWidth
    
    if (clientWidth > 0) {
      /**
       * 核心修复：精确的分页计算公式
       * 在 CSS Column 布局中：scrollWidth = totalPages * clientWidth + (totalPages - 1) * gap
       * 推导得出：totalPages = (scrollWidth + gap) / (clientWidth + gap)
       * 我们加一个 5px 的微小容错偏移以应对浏览器渲染舍入误差
       */
      const calculated = Math.round((scrollWidth + columnGap - 5) / (clientWidth + columnGap))
      totalPages.value = Math.max(1, calculated)
      
      if (currentPage.value >= totalPages.value) {
        currentPage.value = totalPages.value - 1
      }
    }
  }
}

const selectChapter = (index: number, fromEnd = false) => {
  if (index < 0 || index >= chapters.value.length) return
  selectedIndex.value = index
  
  nextTick(async () => {
    if (fromEnd) {
      await calculatePages()
      currentPage.value = Math.max(0, totalPages.value - 1)
    } else {
      currentPage.value = 0
    }
    if (trimmedContainer.value) trimmedContainer.value.scrollTop = 0
  })
}

const nextPage = () => {
  if (currentPage.value < totalPages.value - 1) {
    currentPage.value++
  } else {
    if (selectedIndex.value < chapters.value.length - 1) {
      selectChapter(selectedIndex.value + 1)
    }
  }
}

const prevPage = () => {
  if (currentPage.value > 0) {
    currentPage.value--
  } else {
    if (selectedIndex.value > 0) {
      selectChapter(selectedIndex.value - 1, true)
    }
  }
}

const handleScreenClick = (e: MouseEvent | TouchEvent) => {
  const x = 'touches' in e ? (e as TouchEvent).touches[0].clientX : (e as MouseEvent).clientX
  const width = window.innerWidth
  
  if (readingMode.value === 'scroll') {
    if (x > width * 0.3 && x < width * 0.7) showControls.value = !showControls.value
    return
  }

  if (x < width * 0.3) prevPage()
  else if (x > width * 0.7) nextPage()
  else showControls.value = !showControls.value
}

// 5. 监听器
watch([readingMode, viewMode], () => {
  if (readingMode.value === 'page') {
    if (viewMode.value === 'split') viewMode.value = 'trimmed'
    currentPage.value = 0
    setTimeout(calculatePages, 50)
  }
})

watch(currentDisplayContent, () => {
  if (readingMode.value === 'page') calculatePages()
})

const processChapter = async (index: number) => {
  const chap = chapters.value[index]
  if (!chap || trimmedCache.value[index]) return
  selectedIndex.value = index
  streamingContent.value = ''
  isTrimming.value = true
  api.trimStream(chap.content, chap.id, 
    (text) => {
      streamingContent.value += text
      if (selectedIndex.value === index && readingMode.value === 'scroll' && trimmedContainer.value) {
        trimmedContainer.value.scrollTop = trimmedContainer.value.scrollHeight
      }
    },
    () => { isTrimming.value = false },
    () => {
      trimmedCache.value[index] = streamingContent.value
      streamingContent.value = ''
      isTrimming.value = false
      calculatePages()
    }
  )
}

const exportNovel = () => {
  let content = `StoryTrim Export: ${fileName.value}\n\n`
  chapters.value.forEach((chap, idx) => {
    content += `### ${chap.title}\n\n${trimmedCache.value[idx] || chap.content}\n\n`
  })
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a'); a.href = url; a.download = `${fileName.value}_trimmed.txt`; a.click()
}

onMounted(() => {
  if (chapters.value.length === 0) return router.replace('/dashboard')
  if (window.innerWidth < 1024) {
    viewMode.value = 'trimmed'
    readingMode.value = 'page'
  }
  chapters.value.forEach((chap, idx) => { if (chap.trimmed_content) trimmedCache.value[idx] = chap.trimmed_content })
  calculatePages()
  window.addEventListener('resize', () => {
    calculatePages()
  })
})
onUnmounted(() => window.removeEventListener('resize', calculatePages))
</script>

<template>
  <div class="h-screen flex flex-col bg-[#F9F7F1] overflow-hidden select-none font-sans text-slate-900">
    
    <!-- 顶部导航 -->
    <header 
      class="h-14 bg-white/90 backdrop-blur-md border-b border-gray-200/50 flex items-center justify-between px-4 lg:px-6 shrink-0 fixed top-0 left-0 right-0 z-30 transition-transform duration-200"
      :class="showControls ? 'translate-y-0' : '-translate-y-full'"
    >
      <div class="flex items-center gap-2 lg:gap-4 flex-1 min-w-0">
        <button @click.stop="showSidebar = true" class="lg:hidden p-2"><BookOpen class="w-5 h-5 text-teal-600" /></button>
        <button @click="router.push('/dashboard')" class="hidden lg:block p-2 hover:bg-black/5 rounded-full"><ArrowLeft class="w-5 h-5 text-gray-600" /></button>
        <div class="min-w-0">
          <h1 class="font-bold text-gray-800 text-sm truncate max-w-[100px] lg:max-w-[200px]">{{ fileName }}</h1>
        </div>
      </div>
      
      <div class="flex items-center gap-1 bg-gray-100/50 rounded-lg p-1 border border-gray-200 mx-2">
        <button @click="readingMode = 'scroll'" class="p-1.5 rounded-md transition-colors" :class="readingMode === 'scroll' ? 'bg-white text-teal-600 shadow-sm' : 'text-gray-400'"><Layout class="w-4 h-4" /></button>
        <button @click="readingMode = 'page'" class="p-1.5 rounded-md transition-colors" :class="readingMode === 'page' ? 'bg-white text-teal-600 shadow-sm' : 'text-gray-400'"><Columns class="w-4 h-4" /></button>
        <div class="w-px h-4 bg-gray-300 mx-1"></div>
        <button @click="viewMode = 'trimmed'" class="px-2.5 py-1 rounded-md text-[11px] font-bold" :class="viewMode === 'trimmed' ? 'bg-white text-teal-600 shadow-sm' : 'text-gray-400'">精简</button>
        <button @click="viewMode = 'original'" class="px-2.5 py-1 rounded-md text-[11px] font-bold" :class="viewMode === 'original' ? 'bg-white text-teal-600 shadow-sm' : 'text-gray-400'">原文</button>
        <button v-if="readingMode === 'scroll'" @click="viewMode = 'split'" class="hidden lg:block px-2.5 py-1 rounded-md text-[11px] font-bold" :class="viewMode === 'split' ? 'bg-white text-teal-600 shadow-sm' : 'text-gray-400'">对照</button>
      </div>

      <div class="flex items-center gap-1">
        <button @click="startTrim" :disabled="isTrimming || !!trimmedCache[selectedIndex]" class="p-2 transition-colors" :class="trimmedCache[selectedIndex] ? 'text-teal-500' : 'text-gray-400'"><Loader2 v-if="isTrimming" class="w-4 h-4 animate-spin" /><Scissors v-else class="w-4 h-4" /></button>
        <button @click="exportNovel" class="bg-black text-white p-2 rounded-full hover:bg-gray-800 transition-colors ml-1"><Download class="w-4 h-4" /></button>
      </div>
    </header>

    <!-- Main Content -->
    <div class="flex-1 overflow-hidden flex relative" @mousedown="handleScreenClick">
      <aside class="bg-white border-r border-gray-100 flex flex-col shrink-0 fixed inset-y-0 left-0 w-72 z-50 transition-transform duration-300 transform lg:static lg:translate-x-0" :class="showSidebar ? 'translate-x-0' : '-translate-x-full'">
        <div class="p-4 border-b border-gray-100 flex justify-between items-center lg:hidden"><span class="font-bold text-teal-600">目录</span><button @click="showSidebar = false"><X class="w-5 h-5" text-gray-400 /></button></div>
        <div class="p-4 overflow-y-auto custom-scrollbar flex-1 space-y-1">
          <button v-for="(chap, idx) in chapters" :key="idx" @click.stop="selectChapter(idx)" class="w-full text-left px-3 py-3 lg:py-2 rounded-lg text-sm transition-colors flex justify-between items-center group" :class="selectedIndex === idx ? 'bg-teal-50 text-teal-700 font-bold' : 'text-gray-500 hover:bg-gray-50'"><span class="truncate pr-4">{{ chap.title }}</span><CheckCircle2 v-if="trimmedCache[idx]" class="w-3.5 h-3.5 text-teal-500 shrink-0" /></button>
        </div>
      </aside>

      <main class="flex-1 flex overflow-hidden relative">
        <!-- 分页模式 -->
        <div v-if="readingMode === 'page'" class="flex-1 overflow-hidden px-6 lg:px-12 pt-20 pb-10 lg:pt-24 lg:pb-14">
          <div 
            ref="pageContentRef" 
            class="h-full" 
            :style="{ 
              columnWidth: '100%', 
              columnGap: `${columnGap}px`, 
              transform: `translateX(calc(-${currentPage} * (100% + ${columnGap}px)))`, 
              display: 'block' 
            }"
          >
            <div class="max-w-2xl mx-auto h-full">
               <article class="prose prose-slate prose-lg font-serif leading-loose text-gray-800 whitespace-pre-wrap pb-10">
                 <div class="mb-10 opacity-20 italic font-sans text-xs tracking-widest uppercase">{{ selectedChapter?.title }}</div>
                 {{ viewMode === 'original' ? selectedChapter?.content : currentDisplayContent }}
                 <span v-if="isTrimming && !trimmedCache[selectedIndex]" class="inline-block w-1.5 h-5 bg-teal-500 ml-1 animate-pulse align-middle"></span>
               </article>
            </div>
          </div>
          <div class="absolute bottom-4 left-0 right-0 flex justify-center text-[10px] text-gray-300 font-medium tracking-tighter">
            <span>PAGE {{ currentPage + 1 }} / {{ totalPages }}</span>
          </div>
        </div>

        <!-- 滚动模式 -->
        <div v-else ref="trimmedContainer" class="flex-1 overflow-y-auto scroll-smooth pt-14 px-6 lg:px-12 py-10 lg:py-14">
          <div v-if="viewMode === 'split'" class="flex flex-row min-h-full divide-x divide-gray-200/50">
             <section class="flex-1 p-4 lg:p-8"><article class="prose prose-slate prose-lg font-serif leading-loose text-gray-400 whitespace-pre-wrap">{{ selectedChapter?.content }}</article></section>
             <section class="flex-1 p-4 lg:p-8"><article class="prose prose-slate prose-lg font-serif leading-loose text-gray-800 whitespace-pre-wrap">{{ currentDisplayContent }}</article></section>
          </div>
          <div v-else class="max-w-2xl mx-auto min-h-full">
             <article class="prose prose-slate prose-lg font-serif leading-loose text-gray-800 whitespace-pre-wrap">
               {{ viewMode === 'original' ? selectedChapter?.content : currentDisplayContent }}
             </article>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 2px; }

/* 关键修复：防止段落顶出空白页 */
article p { 
  break-inside: avoid-column; 
  margin-bottom: 1.5em; 
}
article p:last-child {
  margin-bottom: 0;
}

[ref="pageContentRef"] {
  height: 100%;
  column-fill: auto;
}
</style>
