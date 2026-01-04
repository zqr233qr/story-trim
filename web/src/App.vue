<script setup lang="ts">
import { ref, computed, nextTick } from 'vue'
import { api, type Chapter } from './api'
import { 
  Upload, BookOpen, Scissors, Loader2, ChevronRight, CheckCircle2, 
  Download, Play, Pause, Moon, Sun, AlertCircle 
} from 'lucide-vue-next'

const chapters = ref<Chapter[]>([])
const selectedIndex = ref<number>(-1)
const isLoading = ref(false)
const isTrimming = ref(false)
const isBatchProcessing = ref(false)
const shouldStopBatch = ref(false)
const fileName = ref('')
const isDarkMode = ref(false)

// 缓存: index -> content
const trimmedCache = ref<Record<number, string>>({})
const streamingContent = ref('')
const trimmedContainer = ref<HTMLElement | null>(null)

const selectedChapter = computed(() => 
  selectedIndex.value >= 0 ? chapters.value[selectedIndex.value] : null
)

// 切换夜间模式
const toggleDarkMode = () => {
  isDarkMode.value = !isDarkMode.value
  if (isDarkMode.value) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}

const handleFileUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  isLoading.value = true
  fileName.value = file.name
  try {
    const res = await api.upload(file)
    if (res.data.code === 0) {
      chapters.value = res.data.data.chapters
      selectedIndex.value = 0
      trimmedCache.value = {}
      streamingContent.value = ''
    }
  } catch (err) {
    alert('上传失败')
  } finally {
    isLoading.value = false
  }
}

// 核心精简逻辑，支持传入 index 以便批量调用
const processChapter = async (index: number) => {
  const chap = chapters.value[index]
  if (!chap) return

  // 如果已有缓存，直接跳过（除非强制重新生成，暂不支持）
  if (trimmedCache.value[index]) return

  // 选中当前正在处理的章节，方便用户围观
  selectedIndex.value = index
  streamingContent.value = ''
  isTrimming.value = true
  
  // 等待流式完成的 Promise
  return new Promise<void>((resolve, reject) => {
    api.trimStream(
      chap.content,
      (text) => {
        streamingContent.value += text
        // 只有当前显示的章节是正在处理的章节时，才自动滚动
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

// 按钮触发的单个处理
const startTrim = async () => {
  if (selectedIndex.value < 0) return
  await processChapter(selectedIndex.value)
}

// 批量处理
const toggleBatchProcess = async () => {
  if (isBatchProcessing.value) {
    shouldStopBatch.value = true
    return
  }

  isBatchProcessing.value = true
  shouldStopBatch.value = false

  // 从当前章节开始，或者从第一个未处理的开始？
  // 策略：从 selectedIndex 开始往后找
  let start = selectedIndex.value
  if (start < 0) start = 0

  for (let i = start; i < chapters.value.length; i++) {
    if (shouldStopBatch.value) break
    
    if (!trimmedCache.value[i]) {
      try {
        await processChapter(i)
        // 稍微停顿一下，避免 API 速率限制（虽然我们也没限）
        await new Promise(r => setTimeout(r, 500))
      } catch (e) {
        console.error("Batch stopped due to error at chapter", i)
        break
      }
    }
  }
  isBatchProcessing.value = false
  shouldStopBatch.value = false
}

// 导出全本
const exportNovel = () => {
  let content = `StoryTrim Export: ${fileName.value}\n\n`
  
  chapters.value.forEach((chap, idx) => {
    content += `### ${chap.title}\n\n`
    if (trimmedCache.value[idx]) {
      content += trimmedCache.value[idx]
    } else {
      content += `[未处理原文]\n${chap.content}`
    }
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

const selectChapter = (index: number) => {
  selectedIndex.value = index
  // 切换章节时，如果之前的流还在跑，streamingContent 会继续被写入，
  // 但 currentDisplayContent 会因为 index 变化而不再显示 streamingContent
  // 除非我们切回正在处理的章节。
  // 这里的逻辑稍微有点复杂：streamingContent 应该绑定到“正在处理的那个章节ID”，而不仅是全局变量。
  // 简单起见：MVP 假设用户不会在生成时乱切，或者切走后看不到实时流也没关系。
  // 改进：processChapter 里设置了 selectedIndex = index，会强制跳过去。
}

const currentDisplayContent = computed(() => {
  if (trimmedCache.value[selectedIndex.value]) {
    return trimmedCache.value[selectedIndex.value]
  }
  // 只有当选中的章节正是当前正在生成的章节时，才显示流
  // 这里的 isTrimming 是全局的，稍微有点不严谨，但配合 processChapter 里的逻辑够用了
  if (isTrimming.value && !trimmedCache.value[selectedIndex.value]) {
    return streamingContent.value
  }
  return ''
})

// 计算进度
const progress = computed(() => {
  const total = chapters.value.length
  if (total === 0) return 0
  const done = Object.keys(trimmedCache.value).length
  return Math.round((done / total) * 100)
})
</script>

<template>
  <div :class="{ 'dark': isDarkMode }" class="h-screen flex flex-col transition-colors duration-300">
    <div class="flex-1 flex overflow-hidden bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100">
      
      <!-- 侧边栏 -->
      <aside v-if="chapters.length > 0" class="w-72 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex flex-col shrink-0">
        <!-- 头部信息 -->
        <div class="p-4 border-b border-gray-100 dark:border-gray-700 flex flex-col gap-3">
          <div class="flex items-center gap-2">
            <BookOpen class="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
            <h1 class="font-bold truncate text-sm" :title="fileName">{{ fileName }}</h1>
          </div>
          
          <!-- 进度条 -->
          <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
            <div class="bg-indigo-600 h-2 rounded-full transition-all duration-500" :style="{ width: progress + '%' }"></div>
          </div>
          <div class="flex justify-between text-xs text-gray-500 dark:text-gray-400">
            <span>{{ Object.keys(trimmedCache).length }} / {{ chapters.length }} 章</span>
            <span>{{ progress }}%</span>
          </div>

          <!-- 批量操作区 -->
          <div class="flex gap-2">
            <button 
              @click="toggleBatchProcess"
              class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-xs font-medium rounded-lg transition-colors"
              :class="isBatchProcessing 
                ? 'bg-red-100 text-red-700 hover:bg-red-200 dark:bg-red-900/30 dark:text-red-300' 
                : 'bg-indigo-50 text-indigo-700 hover:bg-indigo-100 dark:bg-indigo-900/30 dark:text-indigo-300'"
            >
              <Pause v-if="isBatchProcessing" class="w-3.5 h-3.5" />
              <Play v-else class="w-3.5 h-3.5" />
              {{ isBatchProcessing ? '停止' : '自动处理' }}
            </button>
            <button 
              @click="exportNovel"
              class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-xs font-medium bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600 transition-colors"
            >
              <Download class="w-3.5 h-3.5" />
              导出
            </button>
          </div>
        </div>

        <!-- 目录列表 -->
        <nav class="flex-1 overflow-y-auto custom-scrollbar">
          <button 
            v-for="(chap, idx) in chapters" 
            :key="idx"
            @click="selectChapter(idx)"
            class="w-full text-left px-4 py-3 text-sm transition-colors flex items-center justify-between group border-l-4"
            :class="[
              selectedIndex === idx 
                ? 'bg-indigo-50 dark:bg-indigo-900/20 border-indigo-600 text-indigo-700 dark:text-indigo-300' 
                : 'border-transparent hover:bg-gray-50 dark:hover:bg-gray-700/50 text-gray-600 dark:text-gray-400'
            ]"
          >
            <span class="truncate pr-2">{{ chap.title }}</span>
            <div class="flex items-center shrink-0">
              <Loader2 v-if="isBatchProcessing && selectedIndex === idx" class="w-3.5 h-3.5 animate-spin text-indigo-500 mr-2" />
              <CheckCircle2 v-else-if="trimmedCache[idx]" class="w-3.5 h-3.5 text-green-500 mr-2" />
            </div>
          </button>
        </nav>
        
        <!-- 底部模式切换 -->
        <div class="p-4 border-t border-gray-100 dark:border-gray-700">
           <button @click="toggleDarkMode" class="flex items-center gap-2 text-xs text-gray-500 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200">
             <Moon v-if="!isDarkMode" class="w-4 h-4" />
             <Sun v-else class="w-4 h-4" />
             {{ isDarkMode ? '切换亮色' : '切换暗色' }}
           </button>
        </div>
      </aside>

      <!-- 主内容区 -->
      <main class="flex-1 flex flex-col relative min-w-0 bg-white dark:bg-gray-900">
        <!-- 空状态 -->
        <div v-if="chapters.length === 0" class="flex-1 flex flex-col items-center justify-center p-8">
          <div class="max-w-md w-full text-center space-y-6">
            <div class="w-20 h-20 bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 rounded-2xl flex items-center justify-center mx-auto mb-6">
              <Upload class="w-10 h-10" />
            </div>
            <div>
              <h2 class="text-2xl font-bold dark:text-white">StoryTrim</h2>
              <p class="text-gray-500 dark:text-gray-400 mt-2">AI 驱动的小说智能精简工具</p>
            </div>
            <label class="block group cursor-pointer">
              <input 
                type="file" 
                accept=".txt"
                @change="handleFileUpload"
                class="hidden"
              />
              <div class="w-full py-4 px-6 rounded-xl border-2 border-dashed border-gray-300 dark:border-gray-600 group-hover:border-indigo-500 dark:group-hover:border-indigo-400 transition-colors bg-gray-50 dark:bg-gray-800/50 text-gray-500 dark:text-gray-400">
                 <span v-if="!isLoading">点击选择或拖拽 TXT 文件</span>
                 <span v-else class="flex items-center justify-center gap-2 text-indigo-600">
                    <Loader2 class="w-5 h-5 animate-spin" /> 解析中...
                 </span>
              </div>
            </label>
          </div>
        </div>

        <!-- 阅读器 -->
        <template v-else-if="selectedChapter">
          <header class="h-14 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between px-6 shrink-0 shadow-sm z-10">
            <h2 class="font-semibold text-lg truncate pr-4 dark:text-white">{{ selectedChapter.title }}</h2>
            <button 
              @click="startTrim"
              :disabled="isTrimming || !!trimmedCache[selectedIndex]"
              class="flex items-center gap-2 px-4 py-1.5 rounded-lg text-sm font-medium transition-colors shadow-sm"
              :class="trimmedCache[selectedIndex] 
                ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 cursor-default'
                : (isTrimming ? 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-400' : 'bg-indigo-600 text-white hover:bg-indigo-700 dark:hover:bg-indigo-500')"
            >
              <CheckCircle2 v-if="trimmedCache[selectedIndex]" class="w-4 h-4" />
              <Loader2 v-else-if="isTrimming" class="w-4 h-4 animate-spin" />
              <Scissors v-else class="w-4 h-4" />
              {{ trimmedCache[selectedIndex] ? '已完成' : (isTrimming ? '处理中...' : '精简本章') }}
            </button>
          </header>

          <div class="flex-1 overflow-hidden flex divide-x divide-gray-200 dark:divide-gray-700">
            <!-- 左侧原文 -->
            <section class="flex-1 overflow-y-auto p-8 leading-relaxed text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-900 custom-scrollbar">
              <div class="max-w-2xl mx-auto">
                <span class="text-xs font-bold text-gray-400 uppercase tracking-widest block mb-4 select-none">Original</span>
                <div class="whitespace-pre-wrap text-lg font-serif">{{ selectedChapter.content }}</div>
              </div>
            </section>

            <!-- 右侧精简文 -->
            <section ref="trimmedContainer" class="flex-1 overflow-y-auto p-8 leading-relaxed bg-slate-50 dark:bg-gray-800/50 custom-scrollbar">
              <div class="max-w-2xl mx-auto min-h-[50vh]">
                <span class="text-xs font-bold text-gray-400 uppercase tracking-widest block mb-4 select-none flex items-center justify-between">
                  Trimmed (AI)
                  <span v-if="isTrimming && !trimmedCache[selectedIndex]" class="text-xs text-indigo-500 animate-pulse">Typing...</span>
                </span>
                
                <div v-if="currentDisplayContent" class="whitespace-pre-wrap text-lg font-serif text-slate-800 dark:text-slate-200">
                  {{ currentDisplayContent }}<span v-if="isTrimming && !trimmedCache[selectedIndex]" class="inline-block w-2 h-5 align-middle bg-indigo-500 animate-pulse ml-1"></span>
                </div>
                
                <div v-else class="flex flex-col items-center justify-center h-64 border-2 border-dashed border-gray-200 dark:border-gray-700 rounded-xl text-gray-400 dark:text-gray-600 select-none">
                  <Scissors class="w-8 h-8 mb-2 opacity-50" />
                  <p>未处理</p>
                </div>
              </div>
            </section>
          </div>
        </template>
      </main>
    </div>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.5);
  border-radius: 3px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background-color: rgba(107, 114, 128, 0.8);
}

/* Dark mode 适配 */
:global(.dark) .custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: rgba(75, 85, 99, 0.5);
}
:global(.dark) .custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background-color: rgba(156, 163, 175, 0.8);
}
</style>