<script setup lang="ts">
import { ref, computed, nextTick, watch, getCurrentInstance } from 'vue'
import { onLoad, onUnload } from '@dcloudio/uni-app'
import { useUserStore } from '@/stores/user'
import { useBookStore } from '@/stores/book'
import { api } from '@/api'
import { trimStreamByChapterId, trimStreamByMd5 } from '@/api/trim'
import ModeConfigModal from '@/components/ModeConfigModal.vue'
import SettingsPanel from '@/components/SettingsPanel.vue'
import ChapterList from '@/components/ChapterList.vue'
import BatchTaskModal from '@/components/BatchTaskModal.vue'
import GenerationTerminal from '@/components/GenerationTerminal.vue'

const userStore = useUserStore()
const bookStore = useBookStore()
const instance = getCurrentInstance()

// #ifdef APP-PLUS
import { AppRepository } from '@/adapter/app-repository'
const repo = new AppRepository()
// #endif

interface ReadingHistory {
  last_chapter_id: number
  last_prompt_id: number
  updated_at?: string
}

interface LocalReadingHistory {
  last_chapter_id: number
  last_prompt_id: number
  updated_at: number
}

// --- 1. 状态定义 ---
const bookId = ref(0)
const statusBarHeight = ref(uni.getSystemInfoSync().statusBarHeight || 0)
const menuVisible = ref(false)
const showChapterList = ref(false)
const showConfigModal = ref(false)
const showBatchModal = ref(false)
const showSettings = ref(false)
const isMagicActive = ref(false)
const readingMode = ref<'light' | 'dark' | 'sepia'>(uni.getStorageSync('readingMode') as 'light' | 'dark' | 'sepia' || 'light')
const modeColors = {
  light: { bg: '#fafaf9', text: '#1c1917' },
  dark: { bg: '#0c0a09', text: '#e5e5e5' },
  sepia: { bg: '#F5E6D3', text: '#5D4E37' }
}
const fontSize = ref(18)
const pageMode = ref<'scroll' | 'click'>('scroll')
const hideStatusBar = ref(uni.getStorageSync('hideStatusBar') === 'true')

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
let menuAutoHideTimer: any = null // 菜单栏自动隐藏定时器

// 图标常量 (本地静态文件)
const icons = {
  back: '/static/icons/back.svg',
  menu: '/static/icons/menu.svg',
  prev: '/static/icons/prev.svg',
  batch: '/static/icons/batch.svg',
  next: '/static/icons/next.svg',
  settings: '/static/icons/settings.svg'
}

// 计算精简百分比（基于字符数）
const calculateTrimRatio = (original: string[], trimmed: string[]): number => {
  if (!original || original.length === 0 || !trimmed || trimmed.length === 0) return 0
  const originalChars = original.join('').length
  const trimmedChars = trimmed.join('').length
  if (originalChars === 0) return 0
  const ratio = Math.round((1 - trimmedChars / originalChars) * 100)
  return Math.max(0, ratio)
}

// 获取精简模式信息
const getCurrentModeInfo = () => {
  if (!isMagicActive.value || !activeBook.value?.activeModeId) return null
  const promptId = parseInt(activeBook.value.activeModeId)
  const prompt = bookStore.prompts.find(p => p.id === promptId)
  return prompt ? { name: prompt.name, description: prompt.description } : null
}

// 显示模式切换提示
const showModeSwitchTip = (chapter: any, promptId: number) => {
  if (!chapter || !isMagicActive.value) {
    showNotification('已切换为原文')
    return
  }
  
  const prompt = bookStore.prompts.find(p => p.id === promptId)
  if (!prompt) return
  
  const original = chapter.modes['original']
  const trimmed = chapter.modes[`mode_${promptId}`]
  
  if (trimmed) {
    const ratio = calculateTrimRatio(original, trimmed)
    showNotification(`已切换为「${prompt.name}」，精简 ${ratio}%`)
  } else {
    showNotification(`已切换为「${prompt.name}」`)
  }
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

  // 预加载逻辑 (统一入口)
const triggerPreload = async () => {
  if (!activeBook.value) return
  
  const currentIdx = activeBook.value.activeChapterIndex
  const totalChapters = activeBook.value.chapters.length
  
  for (let i = 1; i <= 2; i++) {
    const targetIdx = currentIdx + i
    if (targetIdx >= totalChapters) break
    
    const chapter = activeBook.value.chapters[targetIdx]
    await preloadChapter(chapter)
  }
}

// 单章预加载（只预加载原文）
const preloadChapter = async (chapter: any) => {
  if (!chapter) return
  if (!chapter.isLoaded) {
    await bookStore.fetchChapter(bookId.value, chapter.id)
  }
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
// 监听阅读模式变化，同步修改导航栏颜色
watch(readingMode, (val) => {
  const isDark = val === 'dark' || val === 'sepia'
  uni.setNavigationBarColor({
    frontColor: isDark ? '#ffffff' : '#000000',
    backgroundColor: isDark ? '#0c0a09' : '#fafaf9'
  })
}, { immediate: true })

watch([fontSize, pageMode, isMagicActive], () => {
  if (pageMode.value === 'click') setTimeout(() => refreshWindow(), 100)
  else syncUI() // 滚动模式下也要同步
})

// 监听状态栏隐藏设置变化
watch(hideStatusBar, (val) => {
  // #ifdef APP-PLUS
  console.log('[StatusBar] Setting changed, hideStatusBar:', val)
  plus.navigator.setFullscreen(!!val)
  
  // 开启隐藏状态栏时，菜单栏也一同隐藏
  if (val) {
    menuVisible.value = false
  }
  // #endif
})

onLoad((options) => {
  uni.setKeepScreenOn({ keepScreenOn: true })

  // #ifdef APP-PLUS
  // 进入时根据设置控制状态栏和菜单栏
  plus.navigator.setFullscreen(!!hideStatusBar.value)
  if (hideStatusBar.value) {
    menuVisible.value = false // 菜单栏默认隐藏
  }
  // #endif

  console.log('options ===', options)
  if (options && options.id) {
    bookId.value = parseInt(options.id)
    init()
  }
})

onUnload(() => {
  uni.setKeepScreenOn({ keepScreenOn: false })
  clearTimeout(progressTimer)
  
  // #ifdef APP-PLUS
  // 退出时恢复状态栏显示
  console.log('[StatusBar] Restore status bar')
  plus.navigator.setFullscreen(false)
  // #endif
})

const init = async () => {
  uni.showLoading({ title: '加载中...' })
  await Promise.all([
    bookStore.fetchBookDetail(bookId.value),
    bookStore.fetchPrompts()
  ])
  uni.hideLoading()

  // 1. 决定起始章节索引
  const startIndex = await determineStartChapter()
  if (activeBook.value) {
    activeBook.value.activeChapterIndex = startIndex
  }

  // 2. 恢复精简模式
  const historyPromptId = await getHistoryPromptId()
  if (historyPromptId > 0 && activeBook.value) {
    activeBook.value.activeModeId = historyPromptId.toString()
    isMagicActive.value = true
  }

  // 3. 立即加载当前章节
  await loadCurrentChapter()

  // 4. 立即触发预加载
  triggerPreload()

  if (bookStore.activeBook?.status === 'new') showConfigModal.value = true
  if (!bookStore.activeBook?.activeModeId && bookStore.prompts.length > 0) {
     bookStore.activeBook!.activeModeId = bookStore.prompts[0].id.toString()
  }

  syncUI()
  if (pageMode.value === 'click') refreshWindow()
}

// 决定起始章节索引
const determineStartChapter = async (): Promise<number> => {
  if (!activeBook.value) return 0

  // sync_state=0: 仅本地
  if (activeBook.value.sync_state === 0) {
    // #ifdef APP-PLUS
    const local = await repo.getReadingHistory(bookId.value)
    if (local) {
      const idx = activeBook.value.chapters.findIndex(c => c.id === local.last_chapter_id)
      if (idx !== -1) return idx
    }
    // #endif
    return 0
  }

  // sync_state=1/2: 并行获取本地 + 云端，比对时间戳
  // #ifdef APP-PLUS
  const [local, cloudHistory] = await Promise.all([
    repo.getReadingHistory(bookId.value),
    fetchCloudReadingHistory()
  ])

  let selected = local
  if (cloudHistory && cloudHistory.updated_at) {
    if (!local || (cloudHistory.updated_at > local.updated_at)) {
      selected = cloudHistory
    }
  }

  if (selected) {
    const idx = activeBook.value.chapters.findIndex(c => c.id === selected!.last_chapter_id)
    if (idx !== -1) return idx
  }
  // #endif

  // #ifndef APP-PLUS
  // 小程序端：直接使用云端数据
  const cloudData = await fetchCloudReadingHistory()
  if (cloudData) {
    const idx = activeBook.value.chapters.findIndex(c => c.id === cloudData.last_chapter_id)
    if (idx !== -1) return idx
  }
  // #endif

  return 0
}

// 获取历史的精简模式ID
const getHistoryPromptId = async (): Promise<number> => {
  if (!activeBook.value) return 0

  // sync_state=0: 本地
  if (activeBook.value.sync_state === 0) {
    // #ifdef APP-PLUS
    const local = await repo.getReadingHistory(bookId.value)
    return local?.last_prompt_id || 0
    // #endif
    return 0
  }

  // sync_state=1/2: 云端
  // #ifdef APP-PLUS
  const [local, cloudHistory] = await Promise.all([
    repo.getReadingHistory(bookId.value),
    fetchCloudReadingHistory()
  ])

  let selected = local
  if (cloudHistory && cloudHistory.updated_at) {
    if (!local || (cloudHistory.updated_at > local.updated_at)) {
      selected = cloudHistory
    }
  }
  return selected?.last_prompt_id || 0
  // #endif

  // #ifndef APP-PLUS
  const cloudData = await fetchCloudReadingHistory()
  return cloudData?.last_prompt_id || 0
  // #endif
}

// 加载当前章节内容
const loadCurrentChapter = async () => {
  if (!activeBook.value) return
  const idx = activeBook.value.activeChapterIndex
  const chapter = activeBook.value.chapters[idx]
  if (!chapter) return

  // 1. 加载原文（总是需要）
  if (!chapter.isLoaded) {
    await bookStore.fetchChapter(bookId.value, chapter.id)
  }

  // 2. 如果开启了精简模式，查询并加载精简内容
  if (isMagicActive.value && activeBook.value?.activeModeId) {
    const promptId = parseInt(activeBook.value.activeModeId)
    if (promptId > 0) {
      // 先查询精简状态
      await bookStore.ensureTrimmedStatus(chapter.id)
      // 再加载精简内容
      await bookStore.fetchChapterTrim(bookId.value, chapter.id, promptId)
    }
  }
}

// 从云端获取阅读进度
const fetchCloudReadingHistory = async (): Promise<LocalReadingHistory | null> => {
  const cloudBookId = activeBook.value?.cloud_id || activeBook.value?.id
  if (!cloudBookId) return null
  const res = await api.getBookProgress(cloudBookId)
  if (res.code === 0 && res.data) {
    const h = res.data as ReadingHistory
    return {
      last_chapter_id: h.last_chapter_id,
      last_prompt_id: h.last_prompt_id,
      updated_at: h.updated_at ? new Date(h.updated_at).getTime() : 0
    }
  }
  return null
}

// 保存阅读进度
const saveProgress = async () => {
  if (!activeChapter.value) return
  const chapterId = activeChapter.value.id
  const promptId = isMagicActive.value ? parseInt(activeBook.value?.activeModeId || '0') : 0

  // 1. 本地 SQLite
  // #ifdef APP-PLUS
  await repo.updateProgress(bookId.value, chapterId, promptId)
  // #endif

  // 2. 云端上报 (sync_state=1/2)
  // #ifdef APP-PLUS
  if (activeBook.value?.sync_state !== 0 && activeBook.value?.cloud_id) {
    try {
      await bookStore.updateProgress(bookId.value, chapterId, promptId)
    } catch (e) {
      console.warn('[Progress] Sync to cloud failed', e)
    }
  }
  // #endif
}

// 页面卸载时保存进度
onUnload(() => {
  uni.setKeepScreenOn({ keepScreenOn: false })
  saveProgress()
})

// --- 5. 事件处理 ---
const handleScroll = (e: any) => {
  if (pageMode.value !== 'scroll') return
  const currentScrollTop = e.detail.scrollTop
  const delta = currentScrollTop - lastScrollTop.value
  if (Math.abs(delta) > 50) {
    // 向下滚动隐藏菜单栏
    if (delta > 0 && currentScrollTop > 100) {
      menuVisible.value = false
      // 如果开启了隐藏状态栏，系统状态栏也一同隐藏
      if (hideStatusBar.value) {
        // #ifdef APP-PLUS
        plus.navigator.setFullscreen(true)
        // #endif
      }
    }
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
  // 清除之前的自动隐藏定时器
  if (menuAutoHideTimer) {
    clearTimeout(menuAutoHideTimer)
    menuAutoHideTimer = null
  }

  if (menuVisible.value) {
    // 关闭菜单栏
    menuVisible.value = false
    // 如果开启了隐藏状态栏，系统状态栏也一同隐藏
    if (hideStatusBar.value) {
      // #ifdef APP-PLUS
      plus.navigator.setFullscreen(true)
      // #endif
    }
    return
  }

  // 翻页模式：判断点击区域
  if (pageMode.value === 'click') {
    const info = uni.getSystemInfoSync()
    const x = e.detail.x
    if (x < info.windowWidth * 0.3) {
      // 左侧 30%：上一页
      if (swiperCurrent.value > 0) {
        swiperCurrent.value--
      } else if (activeBook.value!.activeChapterIndex > 0) {
        switchChapter(activeBook.value!.activeChapterIndex - 1, 'end')
      }
      return
    } else if (x > info.windowWidth * 0.7) {
      // 右侧 30%：下一页
      if (swiperCurrent.value < combinedPages.value.length - 1) {
        swiperCurrent.value++
      } else {
        switchChapter(activeBook.value!.activeChapterIndex + 1, 'start')
      }
      return
    }
    // 中间区域：唤出菜单栏
  } else {
    // 滚动模式：点击中间区域唤出菜单栏
    const info = uni.getSystemInfoSync()
    const x = e.detail.x
    const centerStart = info.windowWidth * 0.3
    const centerEnd = info.windowWidth * 0.7
    if (x < centerStart || x > centerEnd) {
      return // 点击边缘不处理
    }
  }

  // 唤出菜单栏
  menuVisible.value = true

  // 如果开启了隐藏状态栏，显示系统状态栏，2秒后自动隐藏菜单栏和系统状态栏
  if (hideStatusBar.value) {
    // #ifdef APP-PLUS
    plus.navigator.setFullscreen(false)
    // #endif
    menuAutoHideTimer = setTimeout(() => {
      menuVisible.value = false
      // #ifdef APP-PLUS
      plus.navigator.setFullscreen(true)
      // #endif
    }, 2000)
  }
}

const handleBack = () => {
  // 关闭菜单栏
  menuVisible.value = false
  // 如果开启了隐藏状态栏，系统状态栏也一同隐藏
  if (hideStatusBar.value) {
    // #ifdef APP-PLUS
    plus.navigator.setFullscreen(true)
    // #endif
  }
  // 返回上一页
  uni.navigateBack()
}

const showNotification = (msg: string) => {
  toastMsg.value = msg
  showToast.value = true
  setTimeout(() => { showToast.value = false }, 2000)
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

  // 全书精简模式
  if (isBatch) {
    showBatchModal.value = false

    if (!activeBook.value) return

    const cloudBookId = activeBook.value.cloud_id || activeBook.value.id
    const success = await bookStore.startFullTrimTask(cloudBookId, promptId)
    if (success) {
      showNotification('已加入后台处理，可在书架页查看进度')
    } else {
      showNotification('启动失败')
    }
    return
  }

  // 单章精简 (混合模式)
  const isTrimmed = activeChapter.value?.trimmed_prompt_ids?.some((id: number) => Number(id) === promptId)
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

  // #ifdef APP-PLUS
  const syncState = activeBook.value?.sync_state || 0
  const cloudChapterId = activeChapter.value?.cloud_id || activeChapter.value?.id
  const cloudBookId = activeBook.value?.cloud_id || activeBook.value?.id

  // sync_state=0 (本地书籍): 使用 trimStreamByMd5，传递内容和 MD5
  if (syncState === 0) {
    if (!rawContent || !activeChapter.value?.md5) {
      showTerminal.value = false
      showNotification('本地书籍缺少内容信息')
      return
    }
    console.log('[Reader] Starting stream by MD5 (sync_state=0):', activeChapter.value.md5)
    trimStreamByMd5(
      rawContent,
      activeChapter.value.md5,
      promptId,
      activeBook.value?.bookMD5 || '',
      activeBook.value?.activeChapterIndex || 0,
      (text) => {
        streamingContent.value += text
      },
      (err) => {
        console.error('[Reader] Stream error:', err)
        showTerminal.value = false
        showNotification('失败: ' + err)
      },
      async () => {
        const lines = streamingContent.value.split('\n')
        currentTextLines.value = lines

        activeBook.value!.activeModeId = promptId.toString()
        isMagicActive.value = true

        if (activeChapter.value) {
          await bookStore.saveChapterTrim(activeBook.value!.id, activeChapter.value!.id, promptId, streamingContent.value)
          if (!activeChapter.value.trimmed_prompt_ids.includes(promptId)) {
            activeChapter.value.trimmed_prompt_ids.push(promptId)
          }
        }

        if (showTerminal.value) {
          setTimeout(() => {
            showTerminal.value = false
            if (pageMode.value === 'click') refreshWindow('keep')
            showModeSwitchTip(activeChapter.value, promptId)
          }, 800)
        } else {
          showModeSwitchTip(activeChapter.value, promptId)
        }
      }
    )
    return
  }

  // sync_state=1/2: 使用 trimStreamByChapterId (按章节 ID)
  console.log('[Reader] Starting stream by chapter ID (sync_state=1/2):', cloudChapterId, 'BookID:', cloudBookId)
  trimStreamByChapterId(
    cloudBookId,
    cloudChapterId,
    promptId,
    (text) => {
      streamingContent.value += text
    },
    (err) => {
      console.error('[Reader] Stream error:', err)
      showTerminal.value = false
      showNotification('失败: ' + err)
    },
    async () => {
      const lines = streamingContent.value.split('\n')
      currentTextLines.value = lines

      activeBook.value!.activeModeId = promptId.toString()
      isMagicActive.value = true

      if (activeChapter.value) {
        await bookStore.saveChapterTrim(activeBook.value!.id, activeChapter.value!.id, promptId, streamingContent.value)
        if (!activeChapter.value.trimmed_prompt_ids.includes(promptId)) {
          activeChapter.value.trimmed_prompt_ids.push(promptId)
        }
      }

      if (showTerminal.value) {
        setTimeout(() => {
          showTerminal.value = false
          if (pageMode.value === 'click') refreshWindow('keep')
          showModeSwitchTip(activeChapter.value, promptId)
        }, 800)
      } else {
        showModeSwitchTip(activeChapter.value, promptId)
      }
    })
    // #endif

    // #ifndef APP-PLUS
  // 小程序端：使用 trimStreamByChapterId (按章节 ID)
  trimStreamByChapterId(
    activeBook.value!.id,
    activeChapter.value!.id,
    promptId,
    (text) => {
      streamingContent.value += text
    },
    (err) => {
      console.error('[Reader] Stream error:', err)
      showTerminal.value = false
      showNotification('失败: ' + err)
    },
    async () => {
      const lines = streamingContent.value.split('\n')
      currentTextLines.value = lines

      activeBook.value!.activeModeId = promptId.toString()
      isMagicActive.value = true

      if (activeChapter.value) {
        await bookStore.saveChapterTrim(activeBook.value!.id, activeChapter.value!.id, promptId, streamingContent.value)
        if (!activeChapter.value.trimmed_prompt_ids.includes(promptId)) {
           activeChapter.value.trimmed_prompt_ids.push(promptId)
        }
      }

      if (showTerminal.value) {
        setTimeout(() => {
          showTerminal.value = false
          if (pageMode.value === 'click') refreshWindow('keep')
          showModeSwitchTip(activeChapter.value, promptId)
        }, 800)
      } else {
        showModeSwitchTip(activeChapter.value, promptId)
      }
    })
  // #endif
}

const switchToMode = async (id: string, showModalOnFailure = true) => {
  // 先检查登录状态
  try {
    const rawToken = uni.getStorageSync('token')
    const isLogin = !!rawToken

    if (!isLogin) {
      uni.showModal({
        title: '需要登录',
        content: '本地书籍仅支持阅读原文',
        showCancel: true,
        confirmText: '去登录',
        success: (res: any) => {
          if (res.confirm) uni.navigateTo({ url: '/pages/login/login' })
        }
      })
      return
    }
  } catch (e) {
    uni.showToast({ title: '系统错误', icon: 'none' })
    return
  }

  // 1. 尝试从本地缓存加载
  const lines = await bookStore.fetchChapterTrim(activeBook.value!.id, activeChapter.value!.id, parseInt(id))

  if (lines) {
    activeBook.value!.activeModeId = id
    isMagicActive.value = true
    syncUI()
    if (pageMode.value === 'click') refreshWindow('keep')
    triggerPreload()
    
    // 显示模式切换提示
    const promptId = parseInt(id)
    setTimeout(() => {
      showModeSwitchTip(activeChapter.value, promptId)
    }, 100)
  } else {
    if (showModalOnFailure) {
      showConfigModal.value = true
    } else {
      showNotification('暂无离线精简内容')
    }
  }
}

const getMagicButtonClass = () => {
  if (isMagicActive.value) {
    switch (readingMode.value) {
      case 'light': return 'bg-teal-500 text-white rotate-12'
      case 'dark': return 'bg-teal-600 text-white rotate-12'
      case 'sepia': return 'bg-amber-600 text-white rotate-12'
    }
  }
  switch (readingMode.value) {
    case 'light': return 'bg-stone-200 text-stone-700'
    case 'dark': return 'bg-stone-700 text-stone-200'
    case 'sepia': return 'bg-amber-200 text-amber-800'
  }
  return 'bg-stone-200 text-stone-700'
}

const toggleMagic = () => {
  if (isMagicActive.value) {
    isMagicActive.value = false
    syncUI() // 切回原文
    if (pageMode.value === 'click') refreshWindow('keep')
    showNotification('已切换为原文')
  } else {
    const targetMode = activeBook.value?.activeModeId || (bookStore.prompts[0]?.id.toString())
    if (targetMode) switchToMode(targetMode, true)
    else showConfigModal.value = true
  }
}

const switchChapter = async (index: number, targetPosition: 'start' | 'end' = 'start') => {
  if (index < 0 || index >= activeBook.value!.chapters.length) return
  
  const targetChapter = activeBook.value!.chapters[index]
  
  // 如果当前是精简模式，先查询目标章节的精简状态
  if (isMagicActive.value && activeBook.value?.activeModeId) {
    const promptId = parseInt(activeBook.value.activeModeId)
    if (promptId > 0) {
      console.log('[Debug] Querying trim status for chapter:', targetChapter.id)
      await bookStore.ensureTrimmedStatus(targetChapter.id)
      
      const hasTrimmed = targetChapter.trimmed_prompt_ids?.includes(promptId)
      if (!hasTrimmed) {
        console.log('[Debug] Mode Keep Failed -> Reset to original')
        const prevPromptId = parseInt(activeBook.value.activeModeId)
        const prevPrompt = bookStore.prompts.find(p => p.id === prevPromptId)
        const modeName = prevPrompt?.name || '当前模式'
        showNotification(`「${modeName}」无精简内容，已切回原文`)
        isMagicActive.value = false
      } else {
        console.log('[Debug] Mode Keep Success')
      }
    }
  }

  isTextTransitioning.value = true
  clearTimeout(progressTimer)

  if (pageMode.value === 'scroll') {
    scrollTop.value = 1
    nextTick(() => { scrollTop.value = 0 })
  }
  
  activeBook.value!.activeChapterIndex = index
  
  if (pageMode.value === 'click') {
    await refreshWindow(targetPosition === 'end' ? 'last' : 'first')
  } else {
    await bookStore.setChapter(index)
    syncUI()
    isTextTransitioning.value = false
  }

   const chapId = activeBook.value!.chapters[index].id
   triggerPreload()
   handleProgressTracking(chapId)

   // 如果保持精简模式，显示提示
   if (isMagicActive.value && activeBook.value?.activeModeId) {
     const promptId = parseInt(activeBook.value.activeModeId)
     if (promptId > 0) {
       setTimeout(() => {
         showModeSwitchTip(activeBook.value!.chapters[index], promptId)
       }, 300)
     }
   }
}
</script>

<template>
  <view :style="{ backgroundColor: modeColors[readingMode].bg, color: modeColors[readingMode].text }"
        class="h-screen w-full flex flex-col relative overflow-hidden transition-colors duration-300">

    <!-- Top Bar -->
    <view v-if="menuVisible" class="fixed top-0 inset-x-0 z-[80] flex flex-col border-b bg-inherit shadow-sm transition-colors duration-300" :style="{ backgroundColor: modeColors[readingMode].bg }">
      <view :style="{ height: statusBarHeight + 'px' }"></view>
      <view class="h-12 flex items-center justify-between px-4">
        <view @click="handleBack" class="p-2 active:opacity-50 transition-opacity">
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
        :class="getMagicButtonClass()"
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
    <ChapterList :show="showChapterList" :chapters="activeBook?.chapters || []" :active-chapter-index="activeBook?.activeChapterIndex || 0" :active-mode-id="activeBook?.activeModeId" :reading-mode="readingMode" :mode-colors="modeColors" @close="showChapterList = false" @select="(idx) => { showChapterList = false; switchChapter(idx) }" />
    <BatchTaskModal :show="showBatchModal" :book-title="activeBook?.title || ''" :prompts="bookStore.prompts" :reading-mode="readingMode" :mode-colors="modeColors" @close="showBatchModal = false" @confirm="(id) => handleStartProcess(id, true)" />
    <ModeConfigModal :show="showConfigModal" :book-title="activeBook?.title || ''" :chapter-title="activeChapter?.title || ''" :prompts="bookStore.prompts" :trimmed-ids="activeChapter?.trimmed_prompt_ids || []" :reading-mode="readingMode" :mode-colors="modeColors" @close="showConfigModal = false" @start="handleStartProcess" />
    <SettingsPanel :show="showSettings" :modes="bookStore.prompts.map(p => p.id.toString())" :prompts="bookStore.prompts" :active-mode="activeBook?.activeModeId || ''" :font-size="fontSize" :reading-mode="readingMode" :mode-colors="modeColors" :page-mode="pageMode" :hide-status-bar="hideStatusBar" @close="showSettings = false" @update:active-mode="switchToMode" @update:font-size="fontSize = $event" @update:reading-mode="(val) => { readingMode = val; uni.setStorageSync('readingMode', val) }" @update:page-mode="pageMode = $event" @update:hide-status-bar="(val) => { hideStatusBar = val; uni.setStorageSync('hideStatusBar', val ? 'true' : 'false') }" />
    <GenerationTerminal :show="showTerminal" :content="streamingContent" :title="generatingTitle" :reading-mode="readingMode" :mode-colors="modeColors" @close="handleTerminalClose" />
    <view v-if="showToast" class="fixed bottom-40 left-1/2 -translate-x-1/2 bg-stone-900 text-white px-4 py-2 rounded-full text-xs z-[110] shadow-2xl">{{ toastMsg }}</view>
  </view>
</template>

<style>
.pb-safe { padding-bottom: env(safe-area-inset-bottom); }
::-webkit-scrollbar { display: none; width: 0; height: 0; color: transparent; }
</style>
