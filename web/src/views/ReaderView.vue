<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, type Prompt } from '../api'
import { useBookStore } from '../stores/book'
import { 
  ArrowLeft, Download, CheckCircle2, Loader2, Play, Pause, Scissors, 
  BookOpen, X, Layout, Columns, ChevronDown
} from 'lucide-vue-next'

const router = useRouter()
const bookStore = useBookStore()

// 1. 基础状态
const chapters = computed(() => bookStore.chapters)
const fileName = computed(() => bookStore.bookTitle)
const selectedIndex = ref<number>(0)
const isTrimming = ref(false)
const viewMode = ref<'split' | 'trimmed' | 'original'>('trimmed')
const showSidebar = ref(false)
const readingMode = ref<'scroll' | 'page'>('scroll')
const showControls = ref(true)
const isChapterLoading = ref(false)
const currentPage = ref(0)
const totalPages = ref(1)

// 提示词模板
const prompts = ref<Prompt[]>([])
const selectedPromptID = ref<number>(2) 
const selectedPrompt = computed(() => prompts.value.find(p => p.id === selectedPromptID.value))

// 2. 缓存池
const trimmedCache = ref<Record<number, Record<number, string>>>({})
const streamingContent = ref('')

const trimmedContainer = ref<HTMLElement | null>(null)
const pageContentRef = ref<HTMLElement | null>(null)

const selectedChapter = computed(() => 
  selectedIndex.value >= 0 ? chapters.value[selectedIndex.value] : null
)

const currentDisplayContent = computed(() => {
  const pID = selectedPromptID.value
  if (trimmedCache.value[pID]?.[selectedIndex.value]) {
    return trimmedCache.value[pID][selectedIndex.value]
  }
  if (isTrimming.value) {
    return streamingContent.value
  }
  return ''
})

const isCurrentChapterTrimmed = computed(() => {
  if (!selectedChapter.value) return false
  const pID = selectedPromptID.value
  return bookStore.trimmedIDs.includes(selectedChapter.value.id) || !!trimmedCache.value[pID]?.[selectedIndex.value]
})

const calculatePages = async () => {
  if (readingMode.value !== 'page') return
  await nextTick(); await nextTick()
  if (pageContentRef.value) {
    const el = pageContentRef.value
    const gap = 48
    const calculated = Math.round((el.scrollWidth + gap) / (el.clientWidth + gap))
    totalPages.value = Math.max(1, calculated)
  }
}

const fetchChapterContent = async (index: number) => {
  const chap = chapters.value[index]
  if (!chap) return

  const pID = selectedPromptID.value
  
  isChapterLoading.value = !chap.content 
  try {
    const res = await api.getChapter(chap.id, pID) 
    if (res.data.code === 0) {
      const data = res.data.data
      bookStore.chapters[index].content = data.content // 更新 store 中的内容
      
      if (data.trimmed_content) {
        if (!trimmedCache.value[pID]) trimmedCache.value[pID] = {}
        trimmedCache.value[pID][index] = data.trimmed_content
        bookStore.markChapterTrimmed(chap.id)
      }
    }
  } catch (err) {
    console.error(err)
  } finally {
    isChapterLoading.value = false
  }
}

const selectChapter = (index: number, fromEnd = false) => {
  if (index < 0 || index >= chapters.value.length) return
  selectedIndex.value = index
  fetchChapterContent(index)
  if (window.innerWidth < 1024) showSidebar.value = false
  
  nextTick(() => {
    if (fromEnd) {
      setTimeout(() => { calculatePages().then(() => { currentPage.value = totalPages.value - 1 }) }, 100)
    } else {
      currentPage.value = 0
    }
    if (trimmedContainer.value) trimmedContainer.value.scrollTop = 0
  })
}

const nextPage = () => {
  if (currentPage.value < totalPages.value - 1) currentPage.value++
  else if (selectedIndex.value < chapters.value.length - 1) selectChapter(selectedIndex.value + 1)
}

const prevPage = () => {
  if (currentPage.value > 0) currentPage.value--
  else if (selectedIndex.value > 0) selectChapter(selectedIndex.value - 1, true)
}

const processChapter = async () => {
  if (!selectedChapter.value) return
  const pID = selectedPromptID.value
  
  streamingContent.value = ''
  isTrimming.value = true
  
  await api.trimStream(
    selectedChapter.value.id,
    pID,
    (text) => {
      streamingContent.value += text
      if (readingMode.value === 'scroll' && trimmedContainer.value) {
        trimmedContainer.value.scrollTop = trimmedContainer.value.scrollHeight
      }
    },
    (err) => { isTrimming.value = false },
    () => {
      if (!trimmedCache.value[pID]) trimmedCache.value[pID] = {}
      trimmedCache.value[pID][selectedIndex.value] = streamingContent.value
      streamingContent.value = ''
      isTrimming.value = false
      bookStore.markChapterTrimmed(selectedChapter.value!.id)
      calculatePages()
    }
  )
}

// 批量精简：调用后台任务接口
const startBatchTask = async () => {
  try {
    const res = await api.startBatchTrim(bookStore.currentBookId, selectedPromptID.value)
    if (res.data.code === 0) {
      bookStore.startTaskPolling(res.data.data.task_id)
    }
  } catch (e) {
    alert('启动任务失败')
  }
}

onMounted(async () => {
  if (chapters.value.length === 0) return router.replace('/dashboard')
  
  try {
    const res = await api.getPrompts()
    if (res.data.code === 0) prompts.value = res.data.data
  } catch (e) {}

  if (bookStore.lastReadInfo) {
    selectedPromptID.value = bookStore.lastReadInfo.last_prompt_id
    const idx = chapters.value.findIndex(c => c.id === bookStore.lastReadInfo!.last_chapter_id)
    if (idx >= 0) selectedIndex.value = idx
  }

  selectChapter(selectedIndex.value)
  window.addEventListener('resize', calculatePages)
})

watch([readingMode, viewMode, selectedPromptID], () => {
  if (readingMode.value === 'page') {
    currentPage.value = 0
    if (viewMode.value === 'split') viewMode.value = 'trimmed'
    setTimeout(calculatePages, 100)
  }
  fetchChapterContent(selectedIndex.value)
})

onUnmounted(() => window.removeEventListener('resize', calculatePages))
</script>

<template>
  <div class="h-screen flex flex-col bg-[#F9F7F1] overflow-hidden select-none font-sans text-slate-900">
    <header class="h-14 bg-white/90 backdrop-blur-md border-b border-gray-200/50 flex items-center justify-between px-4 lg:px-6 shrink-0 fixed top-0 left-0 right-0 z-30 transition-transform duration-300" :class="showControls ? 'translate-y-0' : '-translate-y-full'">
      <div class="flex items-center gap-2 lg:gap-4 flex-1 min-w-0">
        <button @click.stop="showSidebar = true" class="lg:hidden p-2"><BookOpen class="w-5 h-5 text-teal-600" /></button>
        <button @click="router.push('/dashboard')" class="hidden lg:block p-2 hover:bg-black/5 rounded-full"><ArrowLeft class="w-5 h-5 text-gray-600" /></button>
        <div class="min-w-0">
          <h1 class="font-bold text-gray-800 text-sm truncate max-w-[100px] lg:max-w-[180px]">{{ fileName }}</h1>
        </div>
      </div>
      
      <div class="flex items-center gap-1 bg-gray-100/50 rounded-lg p-1 border border-gray-200 mx-2">
        <div class="relative group hidden sm:block mr-1">
          <select v-model="selectedPromptID" class="appearance-none bg-white border-none text-[10px] font-bold px-2 py-1 pr-6 rounded shadow-sm focus:ring-0 cursor-pointer text-teal-700">
            <option v-for="p in prompts" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
          <ChevronDown class="w-3 h-3 absolute right-1.5 top-1/2 -translate-y-1/2 pointer-events-none text-teal-600" />
        </div>
        <button @click="readingMode = 'scroll'" class="p-1.5 rounded-md transition-colors" :class="readingMode === 'scroll' ? 'bg-white text-teal-600 shadow-sm' : 'text-gray-400'"><Layout class="w-4 h-4" /></button>
        <button @click="readingMode = 'page'" class="p-1.5 rounded-md transition-colors" :class="readingMode === 'page' ? 'bg-white text-teal-600 shadow-sm' : 'text-gray-400'"><Columns class="w-4 h-4" /></button>
        <div class="w-px h-4 bg-gray-300 mx-1"></div>
        <button @click="viewMode = 'trimmed'" class="px-2.5 py-1 rounded-md text-[11px] font-bold" :class="viewMode === 'trimmed' ? 'bg-white text-teal-600 shadow-sm' : 'text-gray-400'">精简</button>
        <button @click="viewMode = 'original'" class="px-2.5 py-1 rounded-md text-[11px] font-bold" :class="viewMode === 'original' ? 'bg-white text-teal-600 shadow-sm' : 'text-gray-400'">原文</button>
      </div>

      <div class="flex items-center gap-1">
        <!-- 任务进度展示 -->
        <div v-if="bookStore.isTaskRunning" class="flex items-center gap-2 mr-2 bg-slate-100 px-3 py-1 rounded-full">
          <Loader2 class="w-3 h-3 animate-spin text-teal-600" />
          <span class="text-[10px] font-bold text-slate-600">{{ bookStore.taskProgress }}%</span>
        </div>

        <button @click="processChapter" :disabled="isTrimming || isCurrentChapterTrimmed" class="p-2 transition-colors" :class="isCurrentChapterTrimmed ? 'text-teal-500' : 'text-gray-400'">
          <Loader2 v-if="isTrimming" class="w-4 h-4 animate-spin" /><Scissors v-else class="w-4 h-4" />
        </button>
        <button class="bg-black text-white p-2 rounded-full hover:bg-gray-800 transition-colors ml-1"><Download class="w-4 h-4" /></button>
      </div>
    </header>

    <div class="flex-1 overflow-hidden flex relative">
      <aside class="bg-white border-r border-gray-100 flex flex-col shrink-0 fixed inset-y-0 left-0 w-72 z-50 transition-transform duration-300 transform lg:static lg:translate-x-0" :class="showSidebar ? 'translate-x-0' : '-translate-x-full'">
        <div class="p-4 border-b border-gray-100 flex justify-between items-center lg:hidden"><span class="font-bold text-teal-600">目录</span><button @click="showSidebar = false"><X class="w-5 h-5 text-gray-400" /></button></div>
        <div class="p-4 overflow-y-auto custom-scrollbar flex-1 space-y-1">
          <button v-for="(chap, idx) in chapters" :key="idx" @click.stop="selectChapter(idx)" class="w-full text-left px-3 py-3 lg:py-2 rounded-lg text-sm transition-colors flex justify-between items-center group" :class="selectedIndex === idx ? 'bg-teal-50 text-teal-700 font-bold' : 'text-gray-500 hover:bg-gray-50'">
            <span class="truncate pr-4 text-[13px] lg:text-sm">{{ chap.title }}</span>
            <CheckCircle2 v-if="bookStore.trimmedIDs.includes(chap.id)" class="w-3.5 h-3.5 text-teal-500 shrink-0" />
          </button>
        </div>
      </aside>

      <main class="flex-1 flex overflow-hidden relative bg-[#F9F7F1]" @mousedown="() => showControls = !showControls">
        <div v-if="isChapterLoading" class="absolute inset-0 z-20 flex flex-col items-center justify-center bg-[#F9F7F1]/80 backdrop-blur-sm"><Loader2 class="w-8 h-8 animate-spin text-teal-600 mb-4" /><p class="text-xs text-gray-400 uppercase tracking-widest tracking-widest">Loading Content...</p></div>

        <!-- 分页模式 -->
        <div v-if="readingMode === 'page'" class="flex-1 overflow-hidden px-6 lg:px-12 pt-20 pb-10 lg:pt-24 lg:pb-14">
          <div ref="pageContentRef" class="h-full transition-transform duration-300 ease-out" :style="{ columnWidth: 'calc(100vw - 48px)', columnGap: '48px', transform: `translateX(calc(-${currentPage} * (100% + 48px)))`, display: 'block' }">
            <div class="max-w-2xl mx-auto h-full">
               <article class="prose prose-slate prose-lg font-serif leading-loose text-gray-800 whitespace-pre-wrap">
                 <div class="mb-10 opacity-20 italic font-sans text-xs tracking-widest uppercase">{{ selectedChapter?.title }}</div>
                 <template v-if="viewMode === 'original' || currentDisplayContent">
                    {{ viewMode === 'original' ? selectedChapter?.content : currentDisplayContent }}
                    <span v-if="isTrimming" class="inline-block w-1.5 h-5 bg-teal-500 ml-1 animate-pulse align-middle"></span>
                 </template>
                 <div v-else class="flex flex-col items-center justify-center h-64 border-2 border-dashed border-gray-200 rounded-2xl text-gray-400">
                    <Scissors class="w-10 h-10 mb-4 opacity-30" />
                    <p class="text-sm font-medium">当前章节尚未精简</p>
                    <div class="flex gap-3 mt-4">
                      <button @click.stop="processChapter" class="px-6 py-2 bg-teal-600 text-white rounded-full text-xs font-bold shadow-lg shadow-teal-600/20 active:scale-95 transition-all">立即精简本章</button>
                      <button @click.stop="startBatchTask" class="px-6 py-2 bg-slate-800 text-white rounded-full text-xs font-bold shadow-lg shadow-slate-800/20 active:scale-95 transition-all flex items-center gap-2"><Play class="w-3 h-3 text-teal-400" /> 精简全书</button>
                    </div>
                 </div>
               </article>
            </div>
          </div>
          <div class="absolute bottom-4 left-0 right-0 flex justify-center text-[10px] text-gray-300 font-medium tracking-tighter"><span>PAGE {{ currentPage + 1 }} / {{ totalPages }}</span></div>
        </div>

        <!-- 滚动模式 -->
        <div v-else ref="trimmedContainer" class="flex-1 overflow-y-auto scroll-smooth pt-14 px-6 lg:px-12 py-10 lg:py-14">
          <div v-if="viewMode === 'split'" class="flex flex-row min-h-full divide-x divide-gray-200/50">
             <section class="flex-1 p-4 lg:p-8"><article class="prose prose-slate prose-lg font-serif leading-loose text-gray-400 whitespace-pre-wrap">{{ selectedChapter?.content }}</article></section>
             <section class="flex-1 p-4 lg:p-8">
                <article v-if="currentDisplayContent" class="prose prose-slate prose-lg font-serif leading-loose text-gray-800 whitespace-pre-wrap">{{ currentDisplayContent }}<span v-if="isTrimming" class="inline-block w-1.5 h-5 bg-teal-500 ml-1 animate-pulse align-middle"></span></article>
                <div v-else class="flex flex-col items-center justify-center h-full text-gray-400">
                   <Scissors class="w-8 h-8 mb-2 opacity-50" /><button @click="processChapter" class="text-xs text-teal-600 font-bold hover:underline">点击精简本章</button>
                </div>
             </section>
          </div>
          <div v-else class="max-w-2xl mx-auto p-8 lg:p-12 min-h-full">
             <article v-if="viewMode === 'original' || currentDisplayContent" class="prose prose-slate prose-lg font-serif leading-loose text-gray-800 whitespace-pre-wrap">{{ viewMode === 'original' ? selectedChapter?.content : currentDisplayContent }}<span v-if="isTrimming" class="inline-block w-1.5 h-5 bg-teal-500 ml-1 animate-pulse align-middle"></span></article>
             <div v-else class="flex flex-col items-center justify-center h-96 border-2 border-dashed border-gray-200 rounded-3xl text-gray-400">
                <Scissors class="w-12 h-12 mb-4 opacity-20" /><p class="text-sm">尚未精简，点击上方工具栏按钮开始</p>
                <div class="flex gap-4 mt-6">
                  <button @click="processChapter" class="px-8 py-2.5 bg-teal-600 text-white rounded-full text-sm font-bold shadow-xl shadow-teal-600/30 active:scale-95 transition-all">立即精简</button>
                  <button @click="startBatchTask" class="px-8 py-2.5 bg-slate-800 text-white rounded-full text-sm font-bold shadow-xl shadow-slate-800/20 active:scale-95 transition-all flex items-center gap-2"><Play class="w-4 h-4 text-teal-400" /> 精简全书</button>
                </div>
             </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 2px; }
article p { break-inside: avoid-column; margin-bottom: 1.5em; }
[ref="pageContentRef"] { height: 100%; column-fill: auto; }
select { border: none !important; outline: none !important; }
</style>
