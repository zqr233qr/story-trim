<script setup lang="ts">
import { ref, computed, nextTick, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'
import { useBookStore } from '../stores/book'
import { 
  ArrowLeft, Type, Download, CheckCircle2, 
  Loader2, Play, Pause, Scissors, ThumbsUp, RotateCcw
} from 'lucide-vue-next'

const router = useRouter()
const bookStore = useBookStore()

const chapters = computed(() => bookStore.chapters)
const fileName = computed(() => bookStore.bookTitle)

const selectedIndex = ref<number>(0)
const isTrimming = ref(false)
const isBatchProcessing = ref(false)
const shouldStopBatch = ref(false)
const viewMode = ref<'split' | 'trimmed' | 'original'>('split')

const trimmedCache = ref<Record<number, string>>({})
const streamingContent = ref('')
const trimmedContainer = ref<HTMLElement | null>(null)

const selectedChapter = computed(() => 
  selectedIndex.value >= 0 ? chapters.value[selectedIndex.value] : null
)

onMounted(() => {
  if (chapters.value.length === 0) {
    router.replace('/dashboard')
    return
  }
  // 初始化已有的精简内容
  chapters.value.forEach((chap, idx) => {
    if (chap.trimmed_content) {
      trimmedCache.value[idx] = chap.trimmed_content
    }
  })
})

const processChapter = async (index: number) => {
  const chap = chapters.value[index]
  if (!chap) return
  if (trimmedCache.value[index]) return

  selectedIndex.value = index
  streamingContent.value = ''
  isTrimming.value = true
  
  return new Promise<void>((resolve, reject) => {
    api.trimStream(
      chap.content,
      chap.id,
      (text) => {
        streamingContent.value += text
        if (selectedIndex.value === index) {
          nextTick(() => {
            if (trimmedContainer.value) {
              trimmedContainer.value.scrollTop = trimmedContainer.value.scrollHeight
            }
          })
        }
      },
      (err) => {
        console.error(err)
        isTrimming.value = false
        reject(err)
      },
      () => {
        trimmedCache.value[index] = streamingContent.value
        streamingContent.value = ''
        isTrimming.value = false
        resolve()
      }
    )
  })
}

const startTrim = async () => {
  if (selectedIndex.value < 0) return
  await processChapter(selectedIndex.value)
}

const toggleBatchProcess = async () => {
  if (isBatchProcessing.value) {
    shouldStopBatch.value = true
    return
  }
  isBatchProcessing.value = true
  shouldStopBatch.value = false

  for (let i = selectedIndex.value; i < chapters.value.length; i++) {
    if (shouldStopBatch.value) break
    if (!trimmedCache.value[i]) {
      try {
        await processChapter(i)
        await new Promise(r => setTimeout(r, 500))
      } catch (e) {
        break
      }
    }
  }
  isBatchProcessing.value = false
}

const exportNovel = () => {
  let content = `StoryTrim Export: ${fileName.value}\n\n`
  chapters.value.forEach((chap, idx) => {
    content += `### ${chap.title}\n\n`
    content += trimmedCache.value[idx] ? trimmedCache.value[idx] : `[未处理原文]\n${chap.content}`
    content += `\n\n${'-'.repeat(20)}\n\n`
  })
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${fileName.value}_trimmed.txt`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

const currentDisplayContent = computed(() => {
  if (trimmedCache.value[selectedIndex.value]) return trimmedCache.value[selectedIndex.value]
  if (isTrimming.value && !trimmedCache.value[selectedIndex.value]) return streamingContent.value
  return ''
})
</script>

<template>
  <div class="h-screen flex flex-col bg-[#F9F7F1]">
    
    <!-- Reader Header -->
    <header class="h-14 bg-white/50 backdrop-blur border-b border-gray-200/50 flex items-center justify-between px-6 shrink-0 sticky top-0 z-20">
      <div class="flex items-center gap-4">
        <button @click="router.push('/dashboard')" class="p-2 hover:bg-black/5 rounded-full transition-colors">
          <ArrowLeft class="w-5 h-5 text-gray-600" />
        </button>
        <div class="min-w-0">
          <h1 class="font-bold text-gray-800 text-sm truncate max-w-[200px]" :title="fileName">{{ fileName }}</h1>
          <p class="text-xs text-gray-500 truncate" v-if="selectedChapter">{{ selectedChapter.title }}</p>
        </div>
      </div>
      
      <!-- 视图模式切换 -->
      <div class="flex items-center gap-2 bg-white rounded-lg p-1 shadow-sm border border-gray-200">
        <button 
          @click="viewMode = 'split'"
          class="px-3 py-1.5 rounded-md text-xs font-medium transition-all"
          :class="viewMode === 'split' ? 'bg-teal-50 text-teal-700 shadow-sm' : 'text-gray-500 hover:bg-gray-50'"
        >对照模式</button>
        <button 
          @click="viewMode = 'trimmed'"
          class="px-3 py-1.5 rounded-md text-xs font-medium transition-all"
          :class="viewMode === 'trimmed' ? 'bg-teal-50 text-teal-700 shadow-sm' : 'text-gray-500 hover:bg-gray-50'"
        >精简版</button>
        <button 
          @click="viewMode = 'original'"
          class="px-3 py-1.5 rounded-md text-xs font-medium transition-all"
          :class="viewMode === 'original' ? 'bg-teal-50 text-teal-700 shadow-sm' : 'text-gray-500 hover:bg-gray-50'"
        >原版</button>
      </div>

      <div class="flex items-center gap-2">
        <div class="flex items-center gap-1 bg-white rounded-lg p-1 border border-gray-200 mr-2">
          <button 
            @click="startTrim"
            :disabled="isTrimming || !!trimmedCache[selectedIndex]"
            class="px-3 py-1.5 rounded-md text-xs font-medium transition-all flex items-center gap-2"
            :class="trimmedCache[selectedIndex] ? 'text-teal-700 bg-teal-50' : 'text-gray-600 hover:bg-gray-50 disabled:opacity-50'"
          >
            <Loader2 v-if="isTrimming" class="w-3 h-3 animate-spin" />
            <Scissors v-else class="w-3 h-3" />
            {{ trimmedCache[selectedIndex] ? '已精简' : (isTrimming ? '处理中' : '开始精简') }}
          </button>
          <button 
            @click="toggleBatchProcess"
            class="px-3 py-1.5 rounded-md text-xs font-medium text-gray-500 hover:bg-gray-50 transition-all flex items-center gap-2"
            :class="isBatchProcessing ? 'text-red-600 bg-red-50' : ''"
          >
            <Pause v-if="isBatchProcessing" class="w-3 h-3" />
            <Play v-else class="w-3 h-3" />
            {{ isBatchProcessing ? '停止' : '自动' }}
          </button>
        </div>

        <button class="p-2 text-gray-400 hover:text-gray-900 transition-colors">
          <Type class="w-5 h-5" />
        </button>
        <button @click="exportNovel" class="bg-black text-white px-4 py-1.5 rounded-full text-xs font-medium hover:bg-gray-800 transition-colors shadow-lg shadow-gray-900/10 flex items-center gap-2">
          <Download class="w-3 h-3" /> 导出
        </button>
      </div>
    </header>

    <!-- Reader Body -->
    <div class="flex-1 overflow-hidden flex relative">
      
      <!-- Sidebar (Chapters) -->
      <aside class="w-64 bg-white border-r border-gray-100 flex flex-col shrink-0 hidden lg:flex">
        <div class="p-4 overflow-y-auto custom-scrollbar flex-1 space-y-1">
          <div class="text-xs font-bold text-gray-400 uppercase tracking-widest px-3 mb-2">Chapters</div>
          <button 
            v-for="(chap, idx) in chapters"
            :key="idx"
            @click="selectedIndex = idx"
            class="w-full text-left px-3 py-2 rounded-lg text-sm transition-colors flex justify-between items-center group"
            :class="selectedIndex === idx ? 'bg-teal-50 text-teal-700 font-medium' : 'text-gray-500 hover:bg-gray-50'"
          >
            <span class="truncate">{{ chap.title }}</span>
            <CheckCircle2 v-if="trimmedCache[idx]" class="w-3.5 h-3.5 text-teal-500" />
            <Loader2 v-else-if="isBatchProcessing && selectedIndex === idx" class="w-3.5 h-3.5 animate-spin text-teal-500" />
          </button>
        </div>
      </aside>

      <!-- Content Area -->
      <main class="flex-1 flex overflow-hidden">
        
        <!-- Left: Original -->
        <div 
          v-if="viewMode === 'split' || viewMode === 'original'"
          class="flex-1 overflow-y-auto p-8 lg:p-12 border-r border-gray-200/50 bg-white"
        >
          <div class="max-w-2xl mx-auto" v-if="selectedChapter">
            <span class="inline-block px-2 py-1 bg-gray-100 text-gray-500 text-[10px] font-bold uppercase tracking-wider rounded mb-6">Original Text</span>
            <article class="prose prose-slate prose-lg font-serif leading-loose text-gray-500 whitespace-pre-wrap">
              {{ selectedChapter.content }}
            </article>
          </div>
        </div>

        <!-- Right: Trimmed -->
        <div 
          v-if="viewMode === 'split' || viewMode === 'trimmed'"
          ref="trimmedContainer" 
          class="flex-1 overflow-y-auto p-8 lg:p-12 bg-[#F9F7F1]"
        >
          <div class="max-w-2xl mx-auto">
             <span class="inline-block px-2 py-1 bg-teal-100 text-teal-700 text-[10px] font-bold uppercase tracking-wider rounded mb-6">AI Trimmed</span>
             
             <div v-if="currentDisplayContent" class="prose prose-slate prose-lg font-serif leading-loose text-gray-800 whitespace-pre-wrap">
               {{ currentDisplayContent }}<span v-if="isTrimming && !trimmedCache[selectedIndex]" class="inline-block w-1.5 h-5 bg-teal-500 ml-1 animate-pulse align-middle"></span>
             </div>

             <div v-else class="flex flex-col items-center justify-center h-64 border-2 border-dashed border-gray-200 rounded-xl text-gray-400">
                <Scissors class="w-8 h-8 mb-2 opacity-50" />
                <p>点击上方“开始精简”</p>
             </div>

             <!-- Floating Feedback -->
             <div v-if="trimmedCache[selectedIndex]" class="fixed bottom-8 right-8 flex gap-2">
                 <button class="bg-white p-3 rounded-full shadow-lg border border-gray-100 text-gray-400 hover:text-teal-600 hover:scale-110 transition-all">
                     <ThumbsUp class="w-5 h-5" />
                 </button>
                 <button @click="trimmedCache[selectedIndex] = ''; startTrim()" class="bg-white p-3 rounded-full shadow-lg border border-gray-100 text-gray-400 hover:text-red-500 hover:scale-110 transition-all">
                     <RotateCcw class="w-5 h-5" />
                 </button>
             </div>
          </div>
        </div>

      </main>
    </div>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #e2e8f0;
  border-radius: 2px;
}
</style>
