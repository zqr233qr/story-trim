<script setup lang="ts">
import { ref, computed, nextTick, watch, getCurrentInstance } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useUserStore } from '@/stores/user'
import { useBookStore } from '@/stores/book'
import { api } from '@/api'
import ModeConfigModal from '@/components/ModeConfigModal.vue'
import SettingsPanel from '@/components/SettingsPanel.vue'
import ChapterList from '@/components/ChapterList.vue'
import BatchTaskModal from '@/components/BatchTaskModal.vue'
import GenerationTerminal from '@/components/GenerationTerminal.vue'

const userStore = useUserStore()
const bookStore = useBookStore()
const instance = getCurrentInstance()

// --- 1. 状态定义 ---
const bookId = ref(0)
const statusBarHeight = ref(uni.getSystemInfoSync().statusBarHeight || 0)
const menuVisible = ref(false)
const showChapterList = ref(false)
const showConfigModal = ref(false)
const showBatchModal = ref(false)
const showSettings = ref(false)
const isMagicActive = ref(false)
const isDarkMode = ref(false)
const fontSize = ref(18)
const pageMode = ref<'scroll' | 'click'>('scroll')

const showTerminal = ref(false)
const generatingTitle = ref('')
const streamingContent = ref('')
const toastMsg = ref('')
const showToast = ref(false)
const isAiLoading = ref(false)
const isTextTransitioning = ref(false)
const scrollTop = ref(0)
const lastScrollTop = ref(0)

const recordedChapterId = ref(0) // 本章节进度记录锁
let progressTimer: any = null
let preloadTimer: any = null

// 图标常量 (本地静态文件)
const icons = {
  back: '/static/icons/back.svg',
  menu: '/static/icons/menu.svg',
  prev: '/static/icons/prev.svg',
  batch: '/static/icons/batch.svg',
  next: '/static/icons/next.svg',
  settings: '/static/icons/settings.svg'
}

// 滑动窗口分页核心状态
const currentPageIndex = ref(0)
const swiperCurrent = ref(0)
const isSwiperReady = ref(false)
const isWindowShifting = ref(false)

const prevPages = ref<string[][]>([])
const currPages = ref<string[][]>([])
const nextPages = ref<string[][]>([])
const textToMeasure = ref<string[]>([]) // 专门用于测量的中转变量

// *** 显式 UI 状态 ***
const currentTextLines = ref<string[]>([])

// --- 2. 计算属性 ---
const activeBook = computed(() => bookStore.activeBook)
const activeChapter = computed(() => {
  if (!activeBook.value) return null
  return activeBook.value.chapters[activeBook.value.activeChapterIndex]
})

const activeModeName = computed(() => {
  if (!isMagicActive.value) return '原文'
  const modeId = activeBook.value?.activeModeId
  if (!modeId || modeId === 'original') return '原文'
  const prompt = bookStore.prompts.find(p => p.id.toString() === modeId || p.id === parseInt(modeId))
  return prompt ? prompt.name : modeId
})

const combinedPages = computed(() => {
  return [...prevPages.value, ...currPages.value, ...nextPages.value]
})

const relativePageInfo = computed(() => {
  const idx = swiperCurrent.value
  const prevLen = prevPages.value.length
  const currLen = currPages.value.length
  if (idx < prevLen) return `${idx + 1} / ${prevLen}`
  if (idx < prevLen + currLen) return `${idx - prevLen + 1} / ${currLen}`
  return `${idx - prevLen - currLen + 1} / ${nextPages.value.length}`
})

const getPageTitle = (pIdx: number) => {
  const chapters = activeBook.value?.chapters
  if (!chapters) return ''
  const currIdx = activeBook.value!.activeChapterIndex
  const prevLen = prevPages.value.length
  const currLen = currPages.value.length
  if (pIdx < prevLen) return chapters[currIdx - 1]?.title
  if (pIdx < prevLen + currLen) return chapters[currIdx]?.title
  return chapters[currIdx + 1]?.title
}

const isFirstPageOfChapter = (pIdx: number) => {
  const prevLen = prevPages.value.length
  const currLen = currPages.value.length
  return pIdx === 0 || pIdx === prevLen || pIdx === prevLen + currLen
}

// --- 3. 核心逻辑 (UI同步、分页、预加载) ---

// 核心：从 Store 同步数据到 UI 状态
const syncUI = () => {
  if (!activeChapter.value) {
    currentTextLines.value = ['加载中...']
    return
  }
  
  let modeKey = 'original'
  if (isMagicActive.value) {
     const id = activeBook.value?.activeModeId
     if (id && id !== 'original') {
        modeKey = `mode_${id}` // 必须加上前缀以匹配 Store
     }
  }

  // 优先取缓存，否则取原文，再否则提示
  const lines = activeChapter.value.modes[modeKey] || activeChapter.value.modes['original'] || ['暂无内容']
  
  // 避免无意义的赋值触发重绘
  if (JSON.stringify(currentTextLines.value) !== JSON.stringify(lines)) {
    // console.log('[UI] Syncing text lines:', modeKey, lines.length)
    currentTextLines.value = lines
  }
}

const measureText = async (text: string[]): Promise<string[][]> => {
  if (text.length === 0) return []
  
  textToMeasure.value = text
  await nextTick()
  await new Promise(resolve => setTimeout(resolve, 150))
  
  const info = uni.getSystemInfoSync()
  const viewHeight = info.windowHeight - 160
  
  return new Promise((resolve) => {
    const query = uni.createSelectorQuery().in(instance)
    query.selectAll('.measurer-para').boundingClientRect()
    query.exec((res) => {
      const rects = res[0] as any[]
      if (!rects || rects.length === 0) return resolve([text])

      let currentPage: string[] = []
      let currentHeight = 0
      const pages: string[][] = []

      rects.forEach((rect, idx) => {
        const paraText = text[idx]
        const h = rect ? rect.height : 40
        if (currentHeight + h > viewHeight && currentPage.length > 0) {
          pages.push(currentPage)
          currentPage = [paraText]
          currentHeight = h
        } else {
          currentPage.push(paraText)
          currentHeight += h + 24
        }
      })
      if (currentPage.length > 0) pages.push(currentPage)
      resolve(pages)
    })
  })
}

const getChapterText = (idx: number): string[] => {
  const chapters = activeBook.value?.chapters
  if (!chapters || idx < 0 || idx >= chapters.length) return []
  const chap = chapters[idx]
  
  let modeKey = 'original'
  if (isMagicActive.value) {
     const id = activeBook.value?.activeModeId
     if (id && id !== 'original') {
        modeKey = `mode_${id}`
     }
  }
  
  return chap.modes[modeKey] || chap.modes['original'] || []
}

// 加载/刷新滑动窗口 (翻页模式的唯一入口)
const refreshWindow = async (targetPos: 'first' | 'last' | 'keep' = 'first') => {
  if (pageMode.value !== 'click' || !activeBook.value) return
  
  // 确保数据最新
  syncUI()
  
  isSwiperReady.value = false
  isTextTransitioning.value = true
  
  const currentIndex = activeBook.value.activeChapterIndex
  const currentChapter = activeBook.value.chapters[currentIndex]
  
  // 1. 确保当前章原文已加载
  if (!currentChapter.isLoaded) {
    await bookStore.fetchChapter(activeBook.value.id, currentChapter.id)
    syncUI() // 加载完再同步一次
  }
  
  // 2. 并行加载相邻章节 (不阻塞)
  if (currentIndex > 0) bookStore.fetchChapter(bookId.value, activeBook.value.chapters[currentIndex-1].id)
  if (currentIndex < activeBook.value.chapters.length - 1) bookStore.fetchChapter(bookId.value, activeBook.value.chapters[currentIndex+1].id)

  // 3. 依次测量当前窗口的三章内容
  currPages.value = await measureText(currentTextLines.value)
  
  const prevText = getChapterText(currentIndex - 1)
  prevPages.value = prevText.length > 0 ? await measureText(prevText) : []
  
  const nextText = getChapterText(currentIndex + 1)
  nextPages.value = nextText.length > 0 ? await measureText(nextText) : []

  // 4. 重置测量层并定位索引
  textToMeasure.value = []
  const targetIdx = targetPos === 'last' ? prevPages.value.length + currPages.value.length - 1 : prevPages.value.length
  
  swiperCurrent.value = targetIdx
  currentPageIndex.value = targetIdx
  
  await nextTick()
  isSwiperReady.value = true
  isTextTransitioning.value = false
}

// 预加载逻辑 (3s触发检查)
const handlePreloadCheck = (currentIndex: number) => {
  clearTimeout(preloadTimer)
  preloadTimer = setTimeout(async () => {
    if (!activeBook.value) return
    const chapters = activeBook.value.chapters
    const missingIds: number[] = []
    
    for (let i = 1; i <= 3; i++) {
      const nextIdx = currentIndex + i
      if (nextIdx < chapters.length && !chapters[nextIdx].isLoaded) break 
      if (i === 3) return 
    }

    for (let i = 1; i <= 5; i++) {
      const nextIdx = currentIndex + i
      if (nextIdx < chapters.length) {
        const chap = chapters[nextIdx]
        if (!chap.isLoaded) missingIds.push(chap.id)
      }
    }

    if (missingIds.length > 0) {
      const promptId = isMagicActive.value ? Number(activeBook.value.activeModeId) : 0
      await bookStore.fetchBatchChapters(missingIds, promptId)
    }
  }, 3000)
}

// 进度确认逻辑 (5s确认)
const handleProgressTracking = (chapterId: number) => {
  clearTimeout(progressTimer)
  if (recordedChapterId.value === chapterId) return

  progressTimer = setTimeout(async () => {
    if (activeBook.value && activeChapter.value?.id === chapterId) {
      const promptId = isMagicActive.value ? Number(activeBook.value.activeModeId) : 0
      await bookStore.updateProgress(activeBook.value.id, chapterId, promptId)
      recordedChapterId.value = chapterId
    }
  }, 5000)
}

// --- 4. 监听与生命周期 ---
watch(isDarkMode, (val) => {
  uni.setNavigationBarColor({
    frontColor: val ? '#ffffff' : '#000000',
    backgroundColor: val ? '#0c0a09' : '#fafaf9'
  })
}, { immediate: true })

watch([fontSize, pageMode, isMagicActive], () => {
  if (pageMode.value === 'click') setTimeout(() => refreshWindow(), 100)
  else syncUI() // 滚动模式下也要同步
})

onLoad((options) => {
  uni.setKeepScreenOn({ keepScreenOn: true })
  if (options && options.id) {
    bookId.value = parseInt(options.id)
    init()
  }
})

const init = async () => {
  uni.showLoading({ title: '加载中...' })
  await Promise.all([
    bookStore.fetchBookDetail(bookId.value),
    bookStore.fetchPrompts()
  ])
  uni.hideLoading()
  
  if (bookStore.activeBook?.status === 'new') showConfigModal.value = true
  if (!bookStore.activeBook?.activeModeId && bookStore.prompts.length > 0) {
     bookStore.activeBook!.activeModeId = bookStore.prompts[0].id.toString()
  }
  
  syncUI() // 初始化 UI
  if (pageMode.value === 'click') refreshWindow()
}

// --- 5. 事件处理 ---
const handleScroll = (e: any) => {
  if (pageMode.value !== 'scroll') return
  const currentScrollTop = e.detail.scrollTop
  const delta = currentScrollTop - lastScrollTop.value
  if (Math.abs(delta) > 50) {
    if (delta > 0 && currentScrollTop > 100) menuVisible.value = false
    else if (delta < 0) menuVisible.value = true
    lastScrollTop.value = currentScrollTop
  }
}

const onSwiperChange = (e: any) => {
  if (!isSwiperReady.value || isWindowShifting.value) return
  const newIdx = e.detail.current
  const prevCount = prevPages.value.length
  const currCount = currPages.value.length
  
  if (newIdx < prevCount) {
    isWindowShifting.value = true
    switchChapter(activeBook.value!.activeChapterIndex - 1, 'end').then(() => { isWindowShifting.value = false })
  } else if (newIdx >= prevCount + currCount) {
    isWindowShifting.value = true
    switchChapter(activeBook.value!.activeChapterIndex + 1, 'start').then(() => { isWindowShifting.value = false })
  } else {
    swiperCurrent.value = newIdx
    currentPageIndex.value = newIdx
  }
}

const handleContentClick = (e: any) => {
  if (menuVisible.value) { menuVisible.value = false; return }
  if (pageMode.value === 'click') {
    const info = uni.getSystemInfoSync()
    const x = e.detail.x
    if (x < info.windowWidth * 0.3) {
      if (swiperCurrent.value > 0) swiperCurrent.value--
      else if (activeBook.value!.activeChapterIndex > 0) switchChapter(activeBook.value!.activeChapterIndex - 1, 'end')
    } else if (x > info.windowWidth * 0.7) {
      if (swiperCurrent.value < combinedPages.value.length - 1) swiperCurrent.value++
      else switchChapter(activeBook.value!.activeChapterIndex + 1, 'start')
    } else { menuVisible.value = true }
  } else { menuVisible.value = true }
}

const showNotification = (msg: string) => {
  toastMsg.value = msg
  showToast.value = true
  setTimeout(() => { showToast.value = false }, 3500)
}

const handleTerminalClose = () => {
  showTerminal.value = false
  showNotification('AI 精简将在后台继续...')
}

const watchBatchTask = (taskId: string, bookName: string) => {
  // ... (Keep existing logic)
}

const handleStartProcess = async (modeId: string | number, isBatch: boolean = false) => {
  const promptId = typeof modeId === 'string' ? parseInt(modeId) : modeId
  
  // 单章精简 (混合模式)
  const isTrimmed = activeChapter.value?.trimmed_prompt_ids?.some(id => Number(id) === promptId)
  if (isTrimmed) {
    showConfigModal.value = false
    await switchToMode(promptId.toString())
    return
  }

  const rawContent = activeChapter.value?.modes['original']?.join('\n')
  if (!rawContent) {
    showNotification('无法获取原文内容')
    return
  }

  showConfigModal.value = false
  generatingTitle.value = bookStore.prompts.find(p => p.id === promptId)?.name || 'AI Processing'
  streamingContent.value = ''
  showTerminal.value = true
  
  // 调用无状态流式接口
  // console.log('[Reader] Starting stream request...')
  api.trimStreamRaw(rawContent, promptId, 
    (text) => { 
      streamingContent.value += text 
    }, 
    (err) => { 
      console.error('[Reader] Stream error:', err)
      showTerminal.value = false
      showNotification('失败: ' + err) 
    }, 
    async () => {
      // console.log('[Reader] Stream done. Total len:', streamingContent.value.length)
      
      // 1. 显式更新 UI (瞬时上屏)
      const lines = streamingContent.value.split('\n')
      currentTextLines.value = lines // <--- 关键修改：直接赋值，不等待 Store
      
      // 2. 更新 Store 状态 (用于下次切换)
      activeBook.value!.activeModeId = promptId.toString()
      isMagicActive.value = true
      
      // 3. 异步保存到持久化层
      if (activeChapter.value) {
        await bookStore.saveChapterTrim(activeBook.value!.id, activeChapter.value!.id, promptId, streamingContent.value)
        if (!activeChapter.value.trimmed_prompt_ids.includes(promptId)) {
           activeChapter.value.trimmed_prompt_ids.push(promptId)
        }
      }
      
      // 4. 关闭终端并刷新窗口
      if (showTerminal.value) {
        setTimeout(() => {
          showTerminal.value = false
          if (pageMode.value === 'click') refreshWindow('keep')
        }, 800)
      } else { 
        showNotification(`精简完成`) 
      }
    }
  )
}

const switchToMode = async (id: string, showModalOnFailure = true) => {
  // console.log('[Reader] switchToMode start', id)
  
  // 1. 尝试从本地缓存加载
  const lines = await bookStore.fetchChapterTrim(activeBook.value!.id, activeChapter.value!.id, parseInt(id))
  
  if (lines) {
    // console.log('[Reader] Cache Hit')
    activeBook.value!.activeModeId = id
    isMagicActive.value = true
    syncUI() // 显式同步
    if (pageMode.value === 'click') refreshWindow('keep')
  } else {
    // console.log('[Reader] Cache Miss')
    if (showModalOnFailure) {
      try {
        const rawToken = uni.getStorageSync('token')
        const isLogin = !!rawToken
        
        if (!isLogin) {
          uni.showModal({
            title: '需要登录',
            content: '离线模式仅支持阅读本地内容。使用 AI 精简功能需要登录账户。',
            showCancel: true,
            confirmText: '去登录',
            success: (res) => {
              if (res.confirm) uni.navigateTo({ url: '/pages/login/login' })
            }
          })
          return
        }
        showConfigModal.value = true
      } catch (e) {
        uni.showToast({ title: '系统错误', icon: 'none' })
      }
    } else {
      showNotification('暂无离线精简内容')
    }
  }
}

const toggleMagic = () => {
  if (isMagicActive.value) {
    isMagicActive.value = false
    syncUI() // 切回原文
    if (pageMode.value === 'click') refreshWindow('keep')
  } else {
    const targetMode = activeBook.value?.activeModeId || (bookStore.prompts[0]?.id.toString())
    if (targetMode) switchToMode(targetMode, true)
    else showConfigModal.value = true
  }
}

const switchChapter = async (index: number, targetPosition: 'start' | 'end' = 'start') => {
  if (index < 0 || index >= activeBook.value!.chapters.length) return
  isTextTransitioning.value = true
  clearTimeout(progressTimer)
  clearTimeout(preloadTimer)

  if (pageMode.value === 'scroll') {
    scrollTop.value = 1
    nextTick(() => { scrollTop.value = 0 })
  }
  
  activeBook.value!.activeChapterIndex = index
  
  if (pageMode.value === 'click') {
    await refreshWindow(targetPosition === 'end' ? 'last' : 'first')
  } else {
    await bookStore.setChapter(index)
    syncUI() // 切换章节后同步
    isTextTransitioning.value = false
  }

  const chapId = activeBook.value!.chapters[index].id
  handlePreloadCheck(index)
  handleProgressTracking(chapId)
}
</script>

<template>
  <view :class="isDarkMode ? 'bg-stone-950 text-stone-300' : 'bg-[#fafaf9] text-stone-800'"
        class="h-screen w-full flex flex-col relative overflow-hidden transition-colors duration-300">
    
    <!-- Top Bar -->
    <view v-if="menuVisible" class="fixed top-0 inset-x-0 z-[80] flex flex-col border-b bg-inherit shadow-sm transition-colors duration-300">
      <view :style="{ height: statusBarHeight + 'px' }"></view>
      <view class="h-12 flex items-center justify-between px-4">
        <view @click="uni.navigateBack()" class="p-2 active:opacity-50 transition-opacity">
          <image :src="icons.back" mode="aspectFit" style="width: 44rpx; height: 44rpx;" class="opacity-70"></image>
        </view>
        <text class="font-bold text-sm truncate max-w-[200px]">{{ activeBook?.title }}</text>
        <view class="w-10"></view>
      </view>
    </view>

    <!-- Main Viewport -->
    <view class="flex-1 min-h-0 w-full" @click="handleContentClick">
      
      <!-- 1. Scroll Mode -->
      <scroll-view v-if="pageMode === 'scroll'" scroll-y class="h-full" :scroll-top="scrollTop" @scroll="handleScroll">
        <view class="p-6 pb-32 transition-opacity duration-300" :style="{ fontSize: fontSize + 'px', paddingTop: (statusBarHeight + 60) + 'px' }" :class="{ 'opacity-0': isTextTransitioning }">
          <view class="text-2xl font-bold mb-10 text-center">{{ activeChapter?.title }}</view>
          
          <!-- Explicit UI Binding -->
          <view v-for="(para, idx) in currentTextLines" :key="idx" class="mb-6 indent-8 leading-loose text-justify">
            {{ para }}
          </view>
          
          <view v-if="activeBook && activeBook.activeChapterIndex < activeBook.chapters.length - 1" class="mt-12 mb-8 flex justify-center">
            <view @click.stop="switchChapter(activeBook.activeChapterIndex + 1)" class="px-8 py-2 rounded-full text-sm font-bold bg-stone-200 text-stone-600">下一章</view>
          </view>
        </view>
      </scroll-view>

      <!-- 2. Click Mode (Swiper Window) -->
      <swiper v-else-if="isSwiperReady"
        class="h-full" 
        :current="swiperCurrent" 
        @change="onSwiperChange"
        :duration="300">
        <swiper-item v-for="(page, pIdx) in combinedPages" :key="pIdx">
          <view class="p-6 h-full flex flex-col transition-opacity duration-300" :style="{ fontSize: fontSize + 'px', paddingTop: (statusBarHeight + 60) + 'px' }" :class="{ 'opacity-0': isTextTransitioning }">
            <view v-if="isFirstPageOfChapter(pIdx)" class="text-2xl font-bold mb-10 text-center">
              {{ getPageTitle(pIdx) }}
            </view>
            <view v-for="(para, idx) in page" :key="idx" class="mb-6 indent-8 leading-loose text-justify">
              {{ para }}
            </view>
          </view>
        </swiper-item>
      </swiper>
    </view>

    <!-- Invisible Measurer -->
    <view class="fixed top-0 left-0 w-full pointer-events-none invisible" style="z-index: -1;">
      <view class="p-6" :style="{ fontSize: fontSize + 'px' }">
        <view v-for="(para, idx) in textToMeasure" :key="idx" class="measurer-para mb-6 indent-8 leading-loose text-justify">
          {{ para }}
        </view>
      </view>
    </view>

    <!-- Overlays -->
    <view class="fixed top-20 right-6 opacity-40 z-10 pointer-events-none" :style="{ top: (statusBarHeight + 60) + 'px' }">
      <text class="text-[10px] font-mono border px-1 rounded">{{ activeModeName }}</text>
    </view>

    <view v-if="pageMode === 'click'" class="fixed bottom-6 right-6 opacity-40 z-10">
      <text class="text-[10px] font-bold">{{ relativePageInfo }}</text>
    </view>

    <!-- Controls -->
    <view v-if="menuVisible" class="fixed bottom-24 right-6 z-40">
      <view @click.stop="toggleMagic" @longpress="showConfigModal = true"
        :class="isMagicActive ? 'bg-teal-500 text-white rotate-12' : 'bg-stone-800 text-white'"
        class="w-14 h-14 rounded-full flex items-center justify-center shadow-xl active:scale-90 transition-all select-none">
        <text v-if="!isAiLoading" class="text-2xl">🪄</text>
        <text v-else class="animate-spin text-xl">⏳</text>
      </view>
    </view>

    <view v-if="menuVisible" class="fixed bottom-0 inset-x-0 bg-inherit border-t pb-safe z-[80]">
      <view class="h-16 flex items-center justify-around px-2">
        <view @click.stop="showChapterList = true" class="flex flex-col items-center gap-1 w-14 active:opacity-50 transition-opacity">
          <image :src="icons.menu" mode="aspectFit" style="width: 44rpx; height: 44rpx;" class="opacity-70"></image>
          <text class="text-[10px] text-stone-400">目录</text>
        </view>
        <view @click.stop="switchChapter(activeBook!.activeChapterIndex - 1)" class="flex flex-col items-center gap-1 w-14 active:opacity-50 transition-opacity">
          <image :src="icons.prev" mode="aspectFit" style="width: 44rpx; height: 44rpx;" class="opacity-70"></image>
          <text class="text-[10px] text-stone-400">上一章</text>
        </view>
        <view @click.stop="showBatchModal = true" class="flex flex-col items-center gap-1 w-14 active:opacity-50 transition-opacity">
          <image :src="icons.batch" mode="aspectFit" style="width: 44rpx; height: 44rpx;" class="opacity-70"></image>
          <text class="text-[10px] text-stone-400">全书处理</text>
        </view>
        <view @click.stop="switchChapter(activeBook!.activeChapterIndex + 1)" class="flex flex-col items-center gap-1 w-14 active:opacity-50 transition-opacity">
          <image :src="icons.next" mode="aspectFit" style="width: 44rpx; height: 44rpx;" class="opacity-70"></image>
          <text class="text-[10px] text-stone-400">下一章</text>
        </view>
        <view @click.stop="showSettings = true" class="flex flex-col items-center gap-1 w-14 active:opacity-50 transition-opacity">
          <image :src="icons.settings" mode="aspectFit" style="width: 44rpx; height: 44rpx;" class="opacity-70"></image>
          <text class="text-[10px] text-stone-400">设置</text>
        </view>
      </view>
    </view>

    <!-- Modals -->
    <ChapterList :show="showChapterList" :chapters="activeBook?.chapters || []" :active-chapter-index="activeBook?.activeChapterIndex || 0" :active-mode-id="activeBook?.activeModeId" :is-dark-mode="isDarkMode" @close="showChapterList = false" @select="(idx) => { showChapterList = false; switchChapter(idx) }" />
    <BatchTaskModal :show="showBatchModal" :book-title="activeBook?.title || ''" :prompts="bookStore.prompts" :is-dark-mode="isDarkMode" @close="showBatchModal = false" @confirm="(id) => handleStartProcess(id, true)" />
    <ModeConfigModal :show="showConfigModal" :book-title="activeBook?.title || ''" :chapter-title="activeChapter?.title || ''" :prompts="bookStore.prompts" :trimmed-ids="activeChapter?.trimmed_prompt_ids || []" :is-dark-mode="isDarkMode" @close="showConfigModal = false" @start="handleStartProcess" />
    <SettingsPanel :show="showSettings" :modes="bookStore.prompts.map(p => p.id.toString())" :prompts="bookStore.prompts" :active-mode="activeBook?.activeModeId || ''" :font-size="fontSize" :is-dark-mode="isDarkMode" :page-mode="pageMode" @close="showSettings = false" @update:active-mode="switchToMode" @update:font-size="fontSize = $event" @update:is-dark-mode="isDarkMode = $event" @update:page-mode="pageMode = $event" />
    <GenerationTerminal :show="showTerminal" :content="streamingContent" :title="generatingTitle" :is-dark-mode="isDarkMode" @close="handleTerminalClose" />
    <view v-if="showToast" class="fixed bottom-40 left-1/2 -translate-x-1/2 bg-stone-900 text-white px-4 py-2 rounded-full text-xs z-[110] shadow-2xl">{{ toastMsg }}</view>
  </view>
</template>

<style>
.pb-safe { padding-bottom: env(safe-area-inset-bottom); }
::-webkit-scrollbar { display: none; width: 0; height: 0; color: transparent; }
</style>
