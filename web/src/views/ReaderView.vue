<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useBookStore } from '../stores/book'
import ModeConfigModal from '../components/ModeConfigModal.vue'
import SettingsPanel from '../components/SettingsPanel.vue'
import ChapterList from '../components/ChapterList.vue'
import BatchTaskModal from '../components/BatchTaskModal.vue'

const route = useRoute()
const router = useRouter()
const bookStore = useBookStore()

const bookId = parseInt(route.params.id as string)

// --- 1. 状态定义 ---
const menuVisible = ref(true)
const showChapterList = ref(false)
const showConfigModal = ref(false)
const showBatchModal = ref(false)
const showSettings = ref(false)
const isMagicActive = ref(false)
const isDarkMode = ref(false)
const fontSize = ref(18)
const pageMode = ref<'scroll' | 'click'>('scroll')
const currentPageIndex = ref(0)
const totalPages = ref(1)
const isTextTransitioning = ref(false)

// --- 2. 计算属性 ---
const activeBook = computed(() => bookStore.activeBook)
const activeChapter = computed(() => bookStore.activeChapter)

const currentText = computed(() => {
  if (!activeChapter.value) return []
  const modeKey = isMagicActive.value ? (activeBook.value?.activeModeId || 'original') : 'original'
  return activeChapter.value.modes[modeKey] || activeChapter.value.modes['original'] || ['加载中...']
})

const activeModeName = computed(() => {
  const map: Record<string, string> = { 'original': '原文', 'dewater': '标准沉浸', 'summary': '轻度精简', 'speed': '极简速读' }
  const key = isMagicActive.value ? (activeBook.value?.activeModeId || 'original') : 'original'
  return map[key] || key
})

// --- 3. 核心分页算法 ---

/**
 * 计算总页数并定位
 * @param targetPosition 'start' 定位到第一页, 'end' 定位到最后一页
 */
const calculatePages = async (targetPosition: 'start' | 'end' = 'start') => {
  const container = document.getElementById('reader-viewport')
  const canvas = document.getElementById('reader-canvas')
  if (!container || !canvas) return

  // 1. 暂时取消位移以便测量
  const originalTransition = canvas.style.transition
  canvas.style.transition = 'none'
  
  await nextTick()
  
  // 2. 高精度测量
  // 在 CSS Columns 模式下，scrollWidth 代表了所有列的总宽度
  const viewW = container.getBoundingClientRect().width
  const totalW = canvas.scrollWidth
  
  if (viewW > 0) {
    totalPages.value = Math.max(1, Math.round(totalW / viewW))
    
    // 3. 定位页码
    if (targetPosition === 'end') {
      currentPageIndex.value = totalPages.value - 1
    } else {
      currentPageIndex.value = 0
    }
    
    console.log(`[Paging] TotalW: ${totalW}, ViewW: ${viewW}, Pages: ${totalPages.value}, SetIndex: ${currentPageIndex.value}`)
  }
  
  // 4. 恢复过渡动画
  setTimeout(() => {
    canvas.style.transition = originalTransition
  }, 50)
}

// 切换章节逻辑（处理跨章定位）
const switchChapter = (index: number, targetPosition: 'start' | 'end' = 'start') => {
  if (!activeBook.value) return
  
  // 开启转场遮罩，避免闪烁
  isTextTransitioning.value = true
  
  // 切换数据
  bookStore.setChapter(index)
  
  // 核心：等待数据渲染后重新分页
  setTimeout(async () => {
    if (pageMode.value === 'click') {
      await calculatePages(targetPosition)
    } else {
      const container = document.getElementById('reader-viewport')
      if (container) container.scrollTop = 0
    }
    isTextTransitioning.value = false
  }, 60)
}

const nextChapter = () => {
  if (!activeBook.value) return
  const nextIdx = activeBook.value.activeChapterIndex + 1
  if (nextIdx < activeBook.value.chapters.length) {
    switchChapter(nextIdx, 'start')
  } else {
    alert('已经是最后一章了')
  }
}

const prevChapter = () => {
  if (!activeBook.value) return
  const prevIdx = activeBook.value.activeChapterIndex - 1
  if (prevIdx >= 0) {
    // 关键：进入上一章并定位到最后一页
    switchChapter(prevIdx, 'end')
  } else {
    alert('已经是第一章了')
  }
}

const navigatePage = (direction: number) => {
  if (direction === 1) {
    // 下一页
    if (currentPageIndex.value + 1 < totalPages.value) {
      currentPageIndex.value++
    } else {
      nextChapter()
    }
  } else {
    // 上一页
    if (currentPageIndex.value > 0) {
      currentPageIndex.value--
    } else {
      prevChapter()
    }
  }
}

// --- 4. 监听与生命周期 ---
watch(isDarkMode, (val) => {
  document.body.style.backgroundColor = val ? '#1c1917' : '#fafaf9'
}, { immediate: true })

watch([currentText, fontSize, pageMode], () => {
  if (pageMode.value === 'click') {
    calculatePages('start')
  }
})

onMounted(() => {
  if (!bookStore.activeBook || bookStore.activeBook.id !== bookId) bookStore.setActiveBook(bookId)
  if (!bookStore.activeBook) {
    router.replace('/shelf')
    return
  }
  
  if (bookStore.activeBook.status === 'new') {
    showConfigModal.value = true
  } else if (bookStore.activeBook.status === 'ready') {
    isMagicActive.value = true
  }
  
  if (pageMode.value === 'click') {
    calculatePages('start')
  }
})

// --- 5. 事件处理 ---
const handleContentClick = (e: MouseEvent) => {
  if (menuVisible.value) {
    menuVisible.value = false
    return
  }
  if (pageMode.value === 'scroll') {
    menuVisible.value = true
    return
  }
  
  const w = window.innerWidth
  const x = e.clientX
  
  if (x < w * 0.3) {
    navigatePage(-1)
  } else if (x > w * 0.7) {
    navigatePage(1)
  } else {
    menuVisible.value = true
  }
}

const toggleMagic = () => {
  if (!activeBook.value || activeBook.value.status !== 'ready') {
    alert('暂无精简数据，请先配置 AI')
    showConfigModal.value = true
    return
  }
  isTextTransitioning.value = true
  setTimeout(() => {
    isMagicActive.value = !isMagicActive.value
    isTextTransitioning.value = false
    if (pageMode.value === 'click') calculatePages('start')
  }, 300)
}

const handleStartProcess = (modeId: string) => {
  showConfigModal.value = false
  bookStore.updateBookStatus(bookId, 'processing')
  setTimeout(() => {
    bookStore.updateBookStatus(bookId, 'ready')
    if (activeBook.value) {
      activeBook.value.activeModeId = modeId
      isMagicActive.value = true
      if (pageMode.value === 'click') calculatePages('start')
    }
  }, 2000)
}

const handleModeChange = (modeId: string) => {
  if (activeBook.value) {
    activeBook.value.activeModeId = modeId
    if (!isMagicActive.value) isMagicActive.value = true
    isTextTransitioning.value = true
    setTimeout(() => {
      isTextTransitioning.value = false
      if (pageMode.value === 'click') calculatePages('start')
    }, 300)
  }
}

const handleBatchTask = (modeId: string) => {
  showBatchModal.value = false
  alert(`已启动全书 [${modeId}] 任务，请在书架查看进度`)
}
</script>

<template>
  <div :class="isDarkMode ? 'bg-stone-900 text-stone-300' : 'bg-[#fafaf9] text-stone-800'" class="h-screen w-full flex flex-col relative transition-colors duration-300 overflow-hidden">
    
    <!-- Top Bar -->
    <transition name="slide-down">
      <div v-show="menuVisible" 
        :class="isDarkMode ? 'bg-stone-900/95 border-stone-800' : 'bg-white/95 border-stone-100'"
        class="fixed top-0 inset-x-0 h-14 backdrop-blur z-30 flex items-center justify-between px-4 shadow-sm border-b transition-colors">
        <button @click="router.push('/shelf')" :class="isDarkMode ? 'text-stone-400 hover:bg-stone-800' : 'text-stone-500 hover:bg-stone-100'" class="p-2 rounded-full transition-colors">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path></svg>
        </button>
        <span :class="isDarkMode ? 'text-stone-300' : 'text-stone-800'" class="font-bold text-sm truncate max-w-[200px]">{{ activeBook?.title }}</span>
        <div class="w-10"></div>
      </div>
    </transition>

    <!-- Reader Viewport -->
    <main 
      id="reader-viewport"
      @click="handleContentClick" 
      :class="[
        pageMode === 'scroll' ? 'overflow-y-auto' : 'overflow-hidden touch-none',
        'flex-1 relative no-scrollbar'
      ]">
      
      <!-- 主画布：使用 Transform 进行物理位移 -->
      <div 
        id="reader-canvas"
        :style="pageMode === 'click' ? {
          columnWidth: '100vw',
          columnGap: '0px',
          columnFill: 'auto',
          height: '100%',
          transform: `translateX(-${currentPageIndex * 100}%)`,
          transition: isTextTransitioning ? 'none' : 'transform 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94)'
        } : { height: 'auto' }"
        class="w-full h-full"
        :class="isTextTransitioning ? 'opacity-0' : 'opacity-100'">
        
        <article 
          :style="{ fontSize: fontSize + 'px' }" 
          :class="[
            isDarkMode ? 'prose-invert text-stone-400' : 'text-stone-800',
            pageMode === 'click' ? 'max-w-none px-8 py-20' : 'max-w-2xl mx-auto px-6 py-20'
          ]"
          class="prose prose-stone font-serif-sc text-justify leading-loose transition-colors select-none">
          
          <h2 class="text-2xl font-bold mb-10 text-center" :class="isDarkMode ? 'text-stone-100' : 'text-stone-900'">
            {{ activeChapter?.title }}
          </h2>

          <div class="content-body">
             <p v-for="(para, index) in currentText" :key="index" 
                class="mb-6 indent-8 break-inside-avoid-column">{{ para }}</p>
          </div>
          
          <div v-if="pageMode === 'scroll'" class="h-32"></div>
        </article>
      </div>

      <!-- 影子容器 (仅用于静默测量，用户看不见) -->
      <div 
        id="shadow-measurer" 
        class="absolute top-0 left-0 w-full pointer-events-none invisible"
        style="height: auto; overflow: visible;">
        <article :style="{ fontSize: fontSize + 'px' }" class="prose px-8 py-20 leading-loose">
          <h2 class="text-2xl font-bold mb-10 text-center">{{ activeChapter?.title }}</h2>
          <p v-for="(para, index) in currentText" :key="index" class="mb-6 indent-8">{{ para }}</p>
        </article>
      </div>
    </main>

    <!-- UI Overlays -->
    
    <!-- TOP-RIGHT: Mode Watermark (As requested) -->
    <div class="fixed top-16 right-6 pointer-events-none transition-opacity duration-500 z-10"
         :class="menuVisible ? 'opacity-0' : 'opacity-40'">
      <span class="text-[9px] font-mono tracking-widest border px-1.5 py-0.5 rounded transition-colors shadow-sm" 
            :class="isDarkMode ? 'text-stone-700 border-stone-800 bg-stone-950/50' : 'text-stone-300 border-stone-100 bg-white/50'">
         {{ isMagicActive ? activeModeName : '原文' }}
      </span>
    </div>

    <!-- BOTTOM-RIGHT: Page Number -->
    <div v-if="pageMode === 'click'" 
         class="fixed bottom-6 right-6 pointer-events-none transition-opacity duration-500 z-10"
         :class="menuVisible ? 'opacity-0' : 'opacity-40'">
      <span class="text-[10px] font-bold tracking-tighter" :class="isDarkMode ? 'text-stone-700' : 'text-stone-400'">
         {{ currentPageIndex + 1 }} / {{ totalPages }}
      </span>
    </div>

    <transition name="pop">
      <div v-show="menuVisible" class="fixed bottom-20 right-6 z-40">
        <button @click.stop="toggleMagic"
          :class="[
            isMagicActive ? 'bg-teal-500 text-white shadow-teal-500/40 rotate-12' : (isDarkMode ? 'bg-stone-700 text-stone-200 shadow-black/50' : 'bg-stone-800 text-white shadow-stone-800/30')
          ]"
          class="w-14 h-14 rounded-full flex items-center justify-center shadow-lg transition-all transform active:scale-95 hover:-translate-y-1 hover:shadow-xl">
          <svg v-if="!isTextTransitioning" class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
          <svg v-else class="animate-spin w-7 h-7" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
        </button>
      </div>
    </transition>

    <transition name="slide-up">
      <div v-show="menuVisible" 
        :class="isDarkMode ? 'bg-stone-900/95 border-stone-800' : 'bg-white/95 border-stone-100'"
        class="fixed bottom-0 inset-x-0 backdrop-blur z-30 border-t pb-safe transition-colors">
        <div class="h-16 flex items-center justify-between px-8 max-w-2xl mx-auto relative">
          <button @click.stop="showChapterList = true" :class="isDarkMode ? 'text-stone-500' : 'text-stone-400'" class="flex flex-col items-center gap-1 w-12 transition-colors">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7"></path></svg>
            <span class="text-[10px]">目录</span>
          </button>
          <button @click.stop="showBatchModal = true" :class="isDarkMode ? 'text-stone-500' : 'text-stone-400'" class="flex flex-col items-center gap-1 w-16 transition-colors">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path></svg>
            <span class="text-[10px]">全书处理</span>
          </button>
          <button @click.stop="showSettings = true" :class="isDarkMode ? 'text-stone-500' : 'text-stone-400'" class="flex flex-col items-center gap-1 w-12 transition-colors">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path></svg>
            <span class="text-[10px]">设置</span>
          </button>
        </div>
      </div>
    </transition>

    <!-- Modals -->
    <ModeConfigModal :show="showConfigModal" :book-title="activeBook?.title || ''" :is-dark-mode="isDarkMode" @close="showConfigModal = false" @start="handleStartProcess" />
    <SettingsPanel :show="showSettings" :modes="Object.keys(activeChapter?.modes || {}).filter(k => k !== 'original')" :active-mode="activeBook?.activeModeId || ''" v-model:font-size="fontSize" v-model:is-dark-mode="isDarkMode" v-model:page-mode="pageMode" @close="showSettings = false" @update:active-mode="handleModeChange" @add-mode="showSettings = false; showConfigModal = true" />
    <ChapterList :show="showChapterList" :chapters="activeBook?.chapters || []" :active-chapter-index="activeBook?.activeChapterIndex || 0" :is-dark-mode="isDarkMode" @close="showChapterList = false" @select="(idx) => { showChapterList = false; switchChapter(idx) }" />
    <BatchTaskModal :show="showBatchModal" :book-title="activeBook?.title || ''" :is-dark-mode="isDarkMode" @close="showBatchModal = false" @confirm="handleBatchTask" />

  </div>
</template>

<style scoped>
.slide-up-enter-active, .slide-up-leave-active, 
.slide-down-enter-active, .slide-down-leave-active { transition: transform 0.3s ease; }
.slide-up-enter-from, .slide-up-leave-to { transform: translateY(100%); }
.slide-down-enter-from, .slide-down-leave-to { transform: translateY(-100%); }
.pop-enter-active, .pop-leave-active { transition: transform 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275), opacity 0.2s; }
.pop-enter-from, .pop-leave-to { transform: scale(0) rotate(-45deg); opacity: 0; }
.pb-safe { padding-bottom: env(safe-area-inset-bottom); }
</style>