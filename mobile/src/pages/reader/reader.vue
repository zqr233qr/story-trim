<script setup lang="ts">
import { ref, computed, nextTick, watch, getCurrentInstance } from 'vue'
import { onLoad, onUnload, onBackPress } from '@dcloudio/uni-app'
import { useUserStore } from '@/stores/user'
import { useBookStore } from '@/stores/book'
import { useNetworkStore } from '@/stores/network'
import { useToastStore } from '@/stores/toast'
import { api } from '@/api'
import { trimStreamByChapterId, trimStreamByMd5 } from '@/api/trim'
import { taskApi } from '@/api/task'
import { pointsApi } from '@/api/points'
import { TTSPlayer } from '@/utils/tts'
import ModeConfigModal from '@/components/ModeConfigModal.vue'
import SettingsPanel from '@/components/SettingsPanel.vue'
import ChapterList from '@/components/ChapterList.vue'
import ChapterTrimModal from '@/components/ChapterTrimModal.vue'
import GenerationTerminal from '@/components/GenerationTerminal.vue'
import LoginConfirmModal from '@/components/LoginConfirmModal.vue'
import SimpleAlertModal from '@/components/SimpleAlertModal.vue'
import Renderjs from './components/Renderjs.vue'

const userStore = useUserStore()
const bookStore = useBookStore()
const networkStore = useNetworkStore()
const toastStore = useToastStore()
const instance = getCurrentInstance()

// 登录引导相关
const showLoginModal = ref(false)
const loginTipContent = ref("")

const openLoginModal = (msg: string) => {
  loginTipContent.value = msg
  showLoginModal.value = true
}

const handleLoginConfirm = () => {
  uni.navigateTo({ url: "/pages/login/login" })
}

const chooseBackgroundImage = async () => {
  try {
    const res = await uni.chooseImage({ count: 1, sourceType: ['album'] })
    const path = res.tempFilePaths?.[0]
    if (!path) return
    readingBgImage.value = path
    uni.setStorageSync('readingBgImage', path)
  } catch (error) {
    console.warn('[Reader] Choose background image failed', error)
  }
}

const clearBackgroundImage = () => {
  readingBgImage.value = ''
  uni.removeStorageSync('readingBgImage')
}

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

interface PageLine {
  text: string
  paraIndex: number
  isFirstLine: boolean
  isParagraphEnd: boolean
}

interface ChapterPage {
  chapterIndex: number
  pageIndex: number
  content: PageLine[]
  chapterTitle: string
  isFirstPage: boolean
}

interface PaginationCacheEntry {
  pages?: PageLine[][]
  lines?: PageLine[]
  paragraphs?: string[]
  contentSize: number
  mode: 'page' | 'line'
}

// 章节精简状态信息。
interface ChapterTrimStatus {
  trimmedIds: number[]
  processingIds: number[]
}

// 章节精简展示项。
interface ChapterTrimOption {
  id: number
  index: number
  title: string
  status: 'available' | 'trimmed' | 'processing'
}

// --- 1. 状态定义 ---
const bookId = ref(0)
const statusBarHeight = ref(uni.getSystemInfoSync().statusBarHeight || 0)
const menuVisible = ref(false)
const showChapterList = ref(false)
const showConfigModal = ref(false)
const showChapterTrimModal = ref(false)
const showSettings = ref(false)
const showTrimConfirmModal = ref(false)
const trimConfirmContent = ref('')
const pendingTrimPayload = ref<{ promptId: number; chapterIds: number[] } | null>(null)

// 指定章节精简积分余额。
const pointsBalance = ref(0)
// 当前选择的精简模式（用于拉取状态）。
const chapterTrimPromptId = ref(0)
// 指定章节精简状态。
const chapterTrimStatus = ref<ChapterTrimStatus>({ trimmedIds: [], processingIds: [] })
const isMagicActive = ref(false)
const readingMode = ref<'light' | 'dark' | 'sepia'>(uni.getStorageSync('readingMode') as 'light' | 'dark' | 'sepia' || 'light')
const readingBgImage = ref(uni.getStorageSync('readingBgImage') || '')
const bgOverlayColor = ref('rgba(0,0,0,0.25)')
const modeColors = {
  light: { bg: '#fafaf9', text: '#1c1917' },
  dark: { bg: '#0c0a09', text: '#e5e5e5' },
  sepia: { bg: '#F5E6D3', text: '#5D4E37' }
}
const fontSize = ref(18)
const pageMode = ref<'scroll' | 'click'>(uni.getStorageSync('pageMode') as 'scroll' | 'click' || 'scroll')
const hideStatusBar = ref(uni.getStorageSync('hideStatusBar') === 'true')

// 用户全局偏好（跨书籍）
const userPreferredModeId = ref(parseInt(uni.getStorageSync('userPreferredModeId') || '0'))
// 当前章节实际显示的模式
const actualDisplayModeId = ref(0)

const showTerminal = ref(false)
const generatingTitle = ref('')
const streamingContent = ref('')
const toastMsg = ref('')
const showToast = ref(false)

// 已读章节索引缓存。
const readChapterIndexes = ref<number[]>([])
const isAiLoading = ref(false)
const isTextTransitioning = ref(false)
const scrollTop = ref(0)
const lastScrollTop = ref(0)

// --- TTS 听书相关 ---
const isTtsSpeaking = ref(false)
const showTtsPanel = ref(false)
const ttsRate = ref(1.0)
const ttsCurrentIndex = ref(-1)
const ttsSleepValue = ref(0) // 0-100 用于 slider 显示
const ttsPlayer = ref<any>(null) // 改为 ref，延迟初始化

const getSleepLabel = (val: number) => {
  if (val <= 0) return '不开启'
  return `${val} 分钟`
}

const handleSleepSliderChange = (e: any) => {
  const val = e.detail.value
  ttsSleepValue.value = val
  
  if (val <= 0) {
    if (ttsPlayer.value) ttsPlayer.value.setSleepTimer('off')
  } else {
    if (ttsPlayer.value) ttsPlayer.value.setSleepTimer(val)
    showNotification(`已设置：${val}分钟后关闭`)
  }
}

// 懒滚动：只有当段落跑出可视区域（或即将跑出）时才滚动
const lazyScrollToParagraph = (index: number) => {
  const query = uni.createSelectorQuery().in(instance)
  query.selectAll('.content-para').boundingClientRect()
  query.select('.scroll-view-content').boundingClientRect() // 获取容器位置信息
  query.exec((res) => {
    const paras = res[0] as any[]
    const container = res[1] as any // 容器信息
    
    if (!paras || !paras[index] || !container) return
    
    const para = paras[index]
    const windowHeight = uni.getSystemInfoSync().windowHeight
    const bottomThreshold = windowHeight * 0.85 // 屏幕底部 15% 的触发线
    const topThreshold = statusBarHeight.value + 60 // 顶部导航栏高度
    
    // 1. 如果段落底部超过了触发线 (需要向下滚)
    // 2. 或者段落顶部已经被遮挡 (需要向上滚 - 极少发生但为了保险)
    if (para.bottom > bottomThreshold || para.top < topThreshold) {
       // 计算目标位置：将该段落滚动到屏幕上方 1/3 处，阅读体验最佳
       // scrollTop = 当前容器scrollTop + (段落top - 容器top) - 目标视觉偏移
       const targetScrollTop = scrollTop.value + (para.top - container.top) - (windowHeight * 0.2)
       
       uni.pageScrollTo({
         scrollTop: Math.max(0, targetScrollTop),
         duration: 400
       })
       scrollTop.value = Math.max(0, targetScrollTop)
    }
  })
}

// 监听 TTS 进度变化，触发自动滚动
watch(ttsCurrentIndex, (newIndex) => {
  if (newIndex < 0) return

  if (pageMode.value === 'scroll') {
    lazyScrollToParagraph(newIndex)
    return
  }

  if (pageMode.value === 'click') {
    const targetPage = ttsPageIndexMap.value.get(newIndex)
    if (targetPage === undefined) return
    if (isPaginating.value || isNavigationLocked.value) return
    if (activePageIndex.value !== targetPage) {
      activePageIndex.value = targetPage
    }
  }
})

// 初始化 TTS (在 init 中调用)
const initTTS = () => {
  if (ttsPlayer.value) return
  
  ttsPlayer.value = new TTSPlayer({
    rate: ttsRate.value,
    title: activeChapter.value?.title || '听书',
    singer: activeBook.value?.title || 'Story Trim',
    onRangeStart: (index) => {
      console.log('[Reader][TTS] range start', {
        index,
        chapterIndex: activeBook.value?.activeChapterIndex,
        chapterId: activeChapter.value?.id
      })
      ttsCurrentIndex.value = index
      // 滚动逻辑已移至 watch(ttsCurrentIndex)
    },
    onEnd: () => {
      if (activeBook.value && activeBook.value.activeChapterIndex < activeBook.value.chapters.length - 1) {
        switchChapter(activeBook.value.activeChapterIndex + 1).then(() => {
          setTimeout(() => {
            startTTS()
          }, 1000)
        })
      } else {
        isTtsSpeaking.value = false
      }
    },
    onNext: () => {
       if (activeBook.value && activeBook.value.activeChapterIndex < activeBook.value.chapters.length - 1) {
          switchChapter(activeBook.value.activeChapterIndex + 1).then(() => {
             setTimeout(() => startTTS(), 500)
          })
       }
    },
    onPrev: () => {
       if (activeBook.value && activeBook.value.activeChapterIndex > 0) {
          switchChapter(activeBook.value.activeChapterIndex - 1).then(() => {
             setTimeout(() => startTTS(), 500)
          })
       }
    }
  })
}

const toggleTtsPanel = () => {
  showTtsPanel.value = !showTtsPanel.value
}

const updateTtsRate = (newRate: number) => {
  ttsRate.value = newRate
  if (ttsPlayer.value) ttsPlayer.value.updateOptions({ rate: newRate })
}

const toggleTTS = () => {
  if (isTtsSpeaking.value) {
    if (ttsPlayer.value) ttsPlayer.value.pause()
    isTtsSpeaking.value = false
  } else {
    startTTS()
  }
}

const startTTS = () => {
  if (!ttsPlayer.value) initTTS()
  
  if (currentTextLines.value.length > 0) {
    console.log('[Reader][TTS] start', {
      chapterIndex: activeBook.value?.activeChapterIndex,
      chapterId: activeChapter.value?.id,
      lines: currentTextLines.value.length,
      ttsCurrentIndex: ttsCurrentIndex.value
    })
    if (ttsCurrentIndex.value > 0) {
       ttsPlayer.value.play(ttsCurrentIndex.value)
    } else {
       ttsPlayer.value.setLines(currentTextLines.value)
       ttsPlayer.value.play(0)
    }
    isTtsSpeaking.value = true
    menuVisible.value = false 
  }
}

const stopTTS = () => {
  if (ttsPlayer.value) ttsPlayer.value.stop()
  isTtsSpeaking.value = false
  ttsCurrentIndex.value = -1
  showTtsPanel.value = false 
}

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
  settings: '/static/icons/settings.svg',
  headphone: '/static/icons/headphone.svg'
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

// 判断段落内容是否有效（非空且含可见文字）
const hasValidLines = (lines?: string[]) => {
  if (!lines || !Array.isArray(lines)) return false
  return lines.some(line => line && line.trim().length > 0)
}

// 获取精简模式信息
const getCurrentModeInfo = () => {
  if (!isMagicActive.value || !activeBook.value?.activeModeId) return null
  const promptId = parseInt(activeBook.value.activeModeId)
  const prompt = bookStore.prompts.find(p => p.id === promptId)
  return prompt ? { name: prompt.name, description: prompt.description } : null
}

// 获取已读章节缓存 Key。
const getReadIndexKey = () => {
  const bookMd5 = activeBook.value?.book_md5
  if (bookMd5) return `read_indexes_${bookMd5}`
  return `read_indexes_${activeBook.value?.id || 0}`
}

// 读取已读章节缓存。
const loadReadIndexes = () => {
  const key = getReadIndexKey()
  const raw = uni.getStorageSync(key)
  if (!raw) {
    readChapterIndexes.value = []
    return
  }
  if (Array.isArray(raw)) {
    readChapterIndexes.value = raw.map((v) => Number(v)).filter((v) => !Number.isNaN(v))
    return
  }
  try {
    const parsed = JSON.parse(raw as string)
    readChapterIndexes.value = Array.isArray(parsed)
      ? parsed.map((v) => Number(v)).filter((v) => !Number.isNaN(v))
      : []
  } catch (e) {
    readChapterIndexes.value = []
  }
}

// 写入已读章节缓存。
const markChapterRead = (index: number) => {
  if (index < 0) return
  if (!readChapterIndexes.value.includes(index)) {
    readChapterIndexes.value = [...readChapterIndexes.value, index]
    const key = getReadIndexKey()
    uni.setStorageSync(key, JSON.stringify(readChapterIndexes.value))
  }
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

// 虚拟章节管理核心状态
const virtualPages = ref<ChapterPage[]>([])
const activePageIndex = ref(0)
const isSwiperReady = ref(false)
const isNavigationLocked = ref(false)

// *** 显式 UI 状态 ***
const currentTextLines = ref<string[]>([])
const isPaginating = ref(false)
const paginationCache = new Map<string, PaginationCacheEntry>()
const ttsPageIndexMap = ref(new Map<number, number>())
const renderjsPayload = ref<any>(null)
const measuredLineHeight = ref(1.7)
const measuredLineHeightPx = ref(0)
let pendingRenderResolve: ((value: { entry: PaginationCacheEntry; pages?: ChapterPage[] }) => void) | null = null


// --- 2. 计算属性 ---
const activeBook = computed(() => bookStore.activeBook)
const activeChapter = computed(() => {
  if (!activeBook.value) return null
  return activeBook.value.chapters[activeBook.value.activeChapterIndex]
})

const activeModeName = computed(() => {
  // 使用 actualDisplayModeId 判断，而不是 isMagicActive
  if (actualDisplayModeId.value <= 0) return '原文'
  const modeId = actualDisplayModeId.value.toString()
  const prompt = bookStore.prompts.find(p => p.id.toString() === modeId || p.id === parseInt(modeId))
  return prompt ? prompt.name : modeId
})

// 获取默认精简模式。
const getDefaultTrimPromptId = () => {
  if (userPreferredModeId.value > 0) return userPreferredModeId.value
  const activeId = Number(activeBook.value?.activeModeId || 0)
  if (activeId > 0) return activeId
  return bookStore.prompts[0]?.id || 0
}

const chapterTrimOptions = computed<ChapterTrimOption[]>(() => {
  const chapters = activeBook.value?.chapters || []
  const processingSet = new Set(chapterTrimStatus.value.processingIds)
  const trimmedSet = new Set(chapterTrimStatus.value.trimmedIds)

  return chapters.map((chapter, index) => {
    const chapterId = chapter.cloud_id || chapter.id
    let status: ChapterTrimOption['status'] = 'available'
    if (processingSet.has(chapterId)) {
      status = 'processing'
    } else if (trimmedSet.has(chapterId)) {
      status = 'trimmed'
    }
    return {
      id: chapterId,
      index: index + 1,
      title: chapter.title,
      status
    }
  })
})

const combinedPages = computed(() => {
  return virtualPages.value.map(p => ({
    key: `${p.chapterIndex}-${p.pageIndex}`,
    content: p.content
  }))
})

const relativePageInfo = computed(() => {
  if (virtualPages.value.length === 0) return ''
  const current = virtualPages.value[activePageIndex.value]
  if (!current) return ''
  const sameChapterPages = virtualPages.value.filter(p => p.chapterIndex === current.chapterIndex)
  const currentIndex = sameChapterPages.findIndex(p => p.pageIndex === current.pageIndex)
  return `${currentIndex + 1} / ${sameChapterPages.length}`
})

const getPageTitle = (pIdx: number) => {
  return activeChapter.value?.title || virtualPages.value[pIdx]?.chapterTitle || ''
}

const isFirstPageOfChapter = (pIdx: number) => {
  return virtualPages.value[pIdx]?.isFirstPage || false
}

const getParagraphGroups = (lines: PageLine[]) => {
  const groups: { paraIndex: number; lines: PageLine[] }[] = []
  let current: { paraIndex: number; lines: PageLine[] } | null = null

  lines.forEach((line) => {
    if (!current || current.paraIndex !== line.paraIndex) {
      current = { paraIndex: line.paraIndex, lines: [line] }
      groups.push(current)
      return
    }
    current.lines.push(line)
  })

  return groups
}

const mapRenderLinesToParagraphs = (lines: any[]) => {
  const hasTitle = lines.some(line => line.isTitle)
  const offset = hasTitle ? 1 : 0
  const grouped = new Map<number, string[]>()

  lines.forEach((line) => {
    if (line.isTitle) return
    const paraIndex = line.pIndex - offset
    if (paraIndex < 0) return
    if (!grouped.has(paraIndex)) grouped.set(paraIndex, [])
    grouped.get(paraIndex)!.push(line.text)
  })

  return Array.from(grouped.values()).map(items => items.join(''))
}

const mapRenderPages = (pages: any[]) => {
  const flat: any[] = []
  pages.forEach((page: any[]) => {
    page.forEach((line) => flat.push(line))
  })

  const hasTitle = flat.some(line => line.isTitle)
  const offset = hasTitle ? 1 : 0
  const lastIndexMap = new Map<number, number>()

  flat.forEach((line, index) => {
    if (line.isTitle) return
    const paraIndex = line.pIndex - offset
    if (paraIndex < 0) return
    lastIndexMap.set(paraIndex, index)
  })

  let flatIndex = 0
  return pages.map((page: any[]) =>
    page.filter((line) => !line.isTitle).map((line) => {
      const paraIndex = line.pIndex - offset
      const isParagraphEnd = lastIndexMap.get(paraIndex) === flatIndex
      const mapped = {
        text: line.text,
        paraIndex,
        isFirstLine: line.pFirst,
        isParagraphEnd
      }
      flatIndex += 1
      return mapped
    })
  )
}

const measureLineHeight = () => {
  const query = uni.createSelectorQuery().in(instance)
  query.select('.line-height-probe').boundingClientRect()
  query.exec((res) => {
    const rect = res?.[0] as { height?: number } | undefined
    if (!rect?.height) return
    measuredLineHeightPx.value = rect.height
    measuredLineHeight.value = rect.height / fontSize.value
  })
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

  let lines = activeChapter.value.modes[modeKey]
  if (!hasValidLines(lines)) {
    lines = activeChapter.value.modes['original']
  }
  if (!hasValidLines(lines)) {
    lines = ['暂无内容']
  }

  if (lines && lines.length) {
    lines = lines.map(line => line.replace(/^[ \t\u3000\xA0]+/, ''))
  }
  
  if (JSON.stringify(currentTextLines.value) !== JSON.stringify(lines)) {
    currentTextLines.value = lines
  }
}

const updateContainerContent = ({ params, chapterPageList }: any) => {
  const inferredIsPage = Array.isArray(chapterPageList?.[0])
  const detectedMode = params?.options?.sizeInfo?.type || (pageMode.value === 'click' ? 'page' : 'line')
  console.log('[Reader][Renderjs] updateContainerContent', {
    hasParams: !!params,
    hasList: !!chapterPageList,
    mode: detectedMode,
    inferredIsPage,
    listLength: Array.isArray(chapterPageList) ? chapterPageList.length : -1,
    lineHeightRatio: params?.options?.sizeInfo?.lineHeight,
    measuredLineHeight: measuredLineHeight.value,
    measuredLineHeightPx: measuredLineHeightPx.value
  })
  if (!params?.options?.cacheKey) return
  const cacheKey = params.options.cacheKey as string
  const mode = detectedMode as 'page' | 'line'
  const contentSize = params.options.contentSize as number

  if (!chapterPageList) return

  if (mode === 'page' || inferredIsPage) {
    const mappedPages = mapRenderPages(chapterPageList)
    const pages: ChapterPage[] = mappedPages.map((pageContent, pageIndex) => ({
      chapterIndex: params.options.chapterIndex,
      pageIndex,
      content: pageContent,
      chapterTitle: params.data?.title || '',
      isFirstPage: pageIndex === 0
    }))

    const pageMap = new Map<number, number>()
    pages.forEach((page, pageIndex) => {
      page.content.forEach(line => {
        if (!pageMap.has(line.paraIndex)) {
          pageMap.set(line.paraIndex, pageIndex)
        }
      })
    })
    ttsPageIndexMap.value = pageMap

    const entry: PaginationCacheEntry = {
      pages: mappedPages,
      contentSize,
      mode: 'page'
    }
    paginationCache.set(cacheKey, entry)

    pendingRenderResolve?.({ entry, pages })
    pendingRenderResolve = null
    virtualPages.value = pages
    isSwiperReady.value = true
  } else {
    const lines = chapterPageList
    const paragraphs = mapRenderLinesToParagraphs(lines)
    const entry: PaginationCacheEntry = {
      lines,
      paragraphs,
      contentSize,
      mode: 'line'
    }
    paginationCache.set(cacheKey, entry)

    pendingRenderResolve?.({ entry })
    pendingRenderResolve = null
    currentTextLines.value = paragraphs
  }

  isPaginating.value = false
  isNavigationLocked.value = false
  isTextTransitioning.value = false
}

defineExpose({ updateContainerContent })

const requestRenderjsPagination = (payload: any) => {
  renderjsPayload.value = payload
  return new Promise<{ entry: PaginationCacheEntry; pages?: ChapterPage[] }>((resolve) => {
    pendingRenderResolve = resolve
  })
}

const buildRenderPayload = (chapterIndex: number, text: string, cacheKey: string, contentSize: number, mode: 'page' | 'line') => {
  const info = uni.getSystemInfoSync()
  const titleSize = fontSize.value + 6
  const lineHeightRatio = measuredLineHeight.value || 1.7
  const paragraphGap = mode === 'page' ? 0 : 16

  return {
    data: {
      title: activeBook.value?.chapters[chapterIndex]?.title || '',
      content: text
    },
    options: {
      cacheKey,
      contentSize,
      chapterIndex,
      pageWidth: info.windowWidth,
      pageHeight: info.windowHeight,
      statusBarHeight: statusBarHeight.value || 0,
      bookOption: {
        sizeInfo: {
          lrPadding: 24,
          infoHeight: 0,
          tPadding: 60,
          bPadding: 24,
          p: fontSize.value,
          lineHeight: lineHeightRatio,
          margin: paragraphGap,
          title: titleSize,
          titleLineHeight: 1.5,
          titleGap: 40,
          type: mode
        }
      }
    }
  }
}

const getChapterText = (idx: number): string[] => {
  const chapters = activeBook.value?.chapters
  if (!chapters || idx < 0 || idx >= chapters.length) return []
  const chap = chapters[idx]

  let modeKey = 'original'
  // 使用actualDisplayModeId而不是isMagicActive和activeBook.value?.activeModeId
  if (actualDisplayModeId.value > 0) {
     modeKey = `mode_${actualDisplayModeId.value}`
  }

  const modeLines = chap.modes[modeKey]
  if (hasValidLines(modeLines)) return modeLines

  const originalLines = chap.modes['original']
  if (hasValidLines(originalLines)) return originalLines

  return []
}

const buildVirtualWindow = async (chapterIndex: number, targetPage: 'first' | 'last' | 'keep' = 'first') => {
  if (!activeBook.value || isPaginating.value) return

  isSwiperReady.value = false
  isNavigationLocked.value = true
  isPaginating.value = true

  const chapter = activeBook.value.chapters[chapterIndex]
  if (!chapter) {
    isSwiperReady.value = true
    isNavigationLocked.value = false
    isTextTransitioning.value = false
    isPaginating.value = false
    return
  }

  if (!chapter.isLoaded) {
    await bookStore.fetchChapter(activeBook.value.id, chapter.id)
  }

  if (actualDisplayModeId.value > 0) {
    const promptId = actualDisplayModeId.value
    if (chapter.trimmed_prompt_ids?.includes(promptId)) {
      await bookStore.fetchChapterTrim(activeBook.value.id, chapter.id, promptId)
    }
  }

  console.log('[Reader][Click] Build pages', {
    chapterIndex,
    chapterId: chapter.id,
    actualDisplayModeId: actualDisplayModeId.value,
    trimmedIds: chapter.trimmed_prompt_ids,
    originalLength: chapter.modes?.original?.length || 0,
    trimmedLength: actualDisplayModeId.value > 0 ? (chapter.modes?.[`mode_${actualDisplayModeId.value}`]?.length || 0) : 0
  })

  let text = getChapterText(chapterIndex)
  if (!hasValidLines(text)) {
    console.warn('[Reader][Click] Content empty, retry fetch', {
      chapterIndex,
      chapterId: chapter.id,
      actualDisplayModeId: actualDisplayModeId.value
    })
    await bookStore.fetchChapter(activeBook.value.id, chapter.id)
    text = getChapterText(chapterIndex)
  }

  if (!hasValidLines(text)) {
    console.warn('[Reader][Click] Content still empty, fallback placeholder', {
      chapterIndex,
      chapterId: chapter.id,
      actualDisplayModeId: actualDisplayModeId.value
    })
    text = ['暂无内容']
  }

  const contentSize = text.join('').length
  const mode = pageMode.value === 'click' ? 'page' : 'line'
  const cacheKey = `${chapter.id}_${actualDisplayModeId.value}_${fontSize.value}_${contentSize}_${mode}`
  let cachedEntry = paginationCache.get(cacheKey)

  if (cachedEntry && cachedEntry.contentSize !== contentSize) {
    paginationCache.delete(cacheKey)
    cachedEntry = undefined
  }

  if (cachedEntry) {
    console.log('[Reader][Click] paginate cache hit', {
      chapterIndex,
      pages: cachedEntry.pages?.length || cachedEntry.lines?.length
    })
  } else {
    console.log('[Reader][Click] paginate start', {
      chapterIndex,
      chapterId: chapter.id,
      textLines: text.length,
      actualDisplayModeId: actualDisplayModeId.value
    })
  }

  if (cachedEntry) {
    if (mode === 'page' && cachedEntry.pages) {
      const pages: ChapterPage[] = cachedEntry.pages.map((pageContent, pageIndex) => ({
        chapterIndex,
        pageIndex,
        content: pageContent,
        chapterTitle: chapter.title,
        isFirstPage: pageIndex === 0
      }))
      virtualPages.value = pages
      const pageMap = new Map<number, number>()
      pages.forEach((page, pageIndex) => {
        page.content.forEach(line => {
          if (!pageMap.has(line.paraIndex)) {
            pageMap.set(line.paraIndex, pageIndex)
          }
        })
      })
      ttsPageIndexMap.value = pageMap
      isSwiperReady.value = true
    }

    if (mode === 'line' && cachedEntry.paragraphs) {
      currentTextLines.value = cachedEntry.paragraphs
    }
  } else {
    const payload = buildRenderPayload(chapterIndex, text.join('\n'), cacheKey, contentSize, mode)
    isPaginating.value = true
    await requestRenderjsPagination(payload)
  }

  const ttsLines = mode === 'line'
    ? currentTextLines.value
    : text.map(line => line.replace(/^[ \t\u3000\xA0]+/, ''))

  if (mode === 'page') {
    currentTextLines.value = ttsLines
  }

  if (isTtsSpeaking.value && ttsPlayer.value) {
    console.log('[Reader][TTS] reset lines after rebuild', {
      chapterIndex,
      chapterId: chapter.id,
      lines: ttsLines.length
    })
    ttsPlayer.value.setLines(ttsLines)
    ttsCurrentIndex.value = 0
    ttsPlayer.value.play(0)
  }

  if (mode === 'page') {
    if (targetPage === 'last') {
      activePageIndex.value = Math.max(0, virtualPages.value.length - 1)
    } else if (targetPage === 'keep') {
      activePageIndex.value = Math.min(activePageIndex.value, Math.max(0, virtualPages.value.length - 1))
    } else {
      activePageIndex.value = 0
    }
  }

  await nextTick()
  isPaginating.value = false
  isNavigationLocked.value = false
  isTextTransitioning.value = false
}

const handlePageNavigation = async (direction: 'prev' | 'next') => {
  if (!activeBook.value || isNavigationLocked.value || virtualPages.value.length === 0) return

  const currentChapterIndex = activeBook.value.activeChapterIndex
  const totalChapters = activeBook.value.chapters.length

  if (direction === 'next') {
    if (activePageIndex.value < virtualPages.value.length - 1) {
      activePageIndex.value++
      return
    }
    if (currentChapterIndex < totalChapters - 1) {
      await switchChapter(currentChapterIndex + 1, 'start')
    } else {
      showNotification('已至末尾')
    }
    return
  }

  if (activePageIndex.value > 0) {
    activePageIndex.value--
    return
  }

  if (currentChapterIndex > 0) {
    await switchChapter(currentChapterIndex - 1, 'end')
  } else {
    showNotification('已至首页')
  }
}

const rebuildVirtualWindow = async (chapterIdx: number, targetPage: 'first' | 'last' | 'keep' = 'first') => {
  await buildVirtualWindow(chapterIdx, targetPage)
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
      
      // 1. 更新本地数据库 (总是执行)
      // #ifdef APP-PLUS
      await repo.updateProgress(bookId.value, chapterId, promptId)
      // #endif

      // 2. 上报云端 (仅已登录且书籍已同步)
      if (userStore.isLoggedIn() && activeBook.value.sync_state !== 0) {
         await bookStore.updateProgress(activeBook.value.id, chapterId, promptId)
      }
      
      recordedChapterId.value = chapterId
    }
  }, 5000)
}

// 监听菜单显示状态，同步关闭听书面板
watch(menuVisible, (val) => {
  if (!val) {
    showTtsPanel.value = false
  }
})

// 监听阅读模式变化，同步修改导航栏颜色
watch(readingMode, (val) => {
  const isDark = val === 'dark' || val === 'sepia'
  uni.setNavigationBarColor({
    frontColor: isDark ? '#ffffff' : '#000000',
    backgroundColor: isDark ? '#0c0a09' : '#fafaf9'
  })
}, { immediate: true })

watch([fontSize, pageMode, actualDisplayModeId], () => {
  if (!activeBook.value) return
  if (isPaginating.value) return
  nextTick(() => measureLineHeight())
  setTimeout(() => buildVirtualWindow(activeBook.value!.activeChapterIndex, 'keep'), 100)
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
  nextTick(() => measureLineHeight())

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
  saveProgress() // 保存进度
  if (ttsPlayer.value) ttsPlayer.value.stop() // 停止听书
  
  // #ifdef APP-PLUS
  // 退出时恢复状态栏显示
  console.log('[StatusBar] Restore status bar')
  plus.navigator.setFullscreen(false)
  // #endif
})

// 返回拦截：优先关闭弹窗。
onBackPress(() => {
  const hasModal = showChapterList.value
    || showConfigModal.value
    || showChapterTrimModal.value
    || showSettings.value
    || showTerminal.value
    || showLoginModal.value
    || showTtsPanel.value

  if (hasModal) {
    showChapterList.value = false
    showConfigModal.value = false
    showChapterTrimModal.value = false
    showSettings.value = false
    showTerminal.value = false
    showLoginModal.value = false
    showTtsPanel.value = false
    menuVisible.value = false
    return true
  }

  return false
})

const init = async () => {
  uni.showLoading({ title: '加载中...' })
  await Promise.all([
    bookStore.fetchBookDetail(bookId.value),
    bookStore.fetchPrompts()
  ])
  uni.hideLoading()

  // 0. 初始化默认偏好 (如果未设置)
  if (userPreferredModeId.value === 0 && bookStore.prompts.length > 0) {
    const defaultPrompt = bookStore.prompts.find(p => p.is_default)
    if (defaultPrompt) {
      console.log('[Reader] Setting default preference:', defaultPrompt.name)
      userPreferredModeId.value = defaultPrompt.id
      uni.setStorageSync('userPreferredModeId', defaultPrompt.id.toString())
    }
  }

  // 1. 决定起始章节索引
   const startIndex = await determineStartChapter()
  if (activeBook.value) {
    activeBook.value.activeChapterIndex = startIndex
  }

  loadReadIndexes()
  markChapterRead(startIndex)

  // 2. 恢复精简模式
  const historyPromptId = await getHistoryPromptId()
  if (historyPromptId > 0 && activeBook.value) {
    activeBook.value.activeModeId = historyPromptId.toString()
    isMagicActive.value = true
  }

  // 3. 恢复用户精简模式偏好（需要检查当前章节是否有缓存）
  if (userPreferredModeId.value > 0 && activeBook.value) {
    const currentChapter = activeBook.value.chapters[activeBook.value.activeChapterIndex]
    await checkAndAdjustMode(currentChapter)
  } else {
    actualDisplayModeId.value = 0
    isMagicActive.value = false
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
  if (pageMode.value === 'click') rebuildVirtualWindow(activeBook.value.activeChapterIndex)

  // 初始化 TTS (此时 activeBook 已就绪)
  initTTS()
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
  if (!userStore.isLoggedIn()) return null
  
  const cloudBookId = activeBook.value?.cloud_id || activeBook.value?.id
  if (!cloudBookId) return null
  
  try {
    const res = await api.getBookProgress(cloudBookId)
    if (res.code === 0 && res.data) {
      const h = res.data as ReadingHistory
      return {
        last_chapter_id: h.last_chapter_id,
        last_prompt_id: h.last_prompt_id,
        updated_at: h.updated_at ? new Date(h.updated_at).getTime() : 0
      }
    }
  } catch (e) {
    console.warn('[Reader] Fetch cloud history failed (network offline?), using local only.', e)
  }
  return null
}

// 保存阅读进度
const saveProgress = async () => {
  if (!activeChapter.value) return
  const chapterId = activeChapter.value.id
  const promptId = isMagicActive.value ? parseInt(activeBook.value?.activeModeId || '0') : 0

  // 1. 本地 SQLite (总是保存)
  // #ifdef APP-PLUS
  await repo.updateProgress(bookId.value, chapterId, promptId)
  // #endif

  // 2. 云端上报 (仅已登录且已同步)
  // #ifdef APP-PLUS
  if (userStore.isLoggedIn() && activeBook.value?.sync_state !== 0 && activeBook.value?.cloud_id) {
    try {
      await bookStore.updateProgress(bookId.value, chapterId, promptId)
    } catch (e) {
      console.warn('[Progress] Sync to cloud failed', e)
    }
  }
  // #endif
}

// 页面卸载时保存进度
// 逻辑已合并到上方的 onUnload

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
      handlePageNavigation('prev')
      return
    } else if (x > info.windowWidth * 0.7) {
      // 右侧 30%：下一页
      handlePageNavigation('next')
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

// 加载积分余额。
const loadPointsBalance = async () => {
  try {
    const res = await pointsApi.getBalance()
    if (res.code === 0) {
      pointsBalance.value = res.data.balance || 0
      return
    }
    showNotification(res.msg || '获取积分失败')
  } catch (e) {
    console.warn('[Reader] load points failed', e)
  }
}

// 加载章节精简状态。
const loadChapterTrimStatus = async () => {
  if (!activeBook.value) return
  const promptId = chapterTrimPromptId.value || getDefaultTrimPromptId()
  if (!promptId) return
  const bookId = activeBook.value.cloud_id || activeBook.value.id
  try {
    const res = await taskApi.getChapterTrimStatus(bookId, promptId)
    if (res.code === 0 && res.data) {
      chapterTrimStatus.value = {
        trimmedIds: res.data.trimmed_chapter_ids || [],
        processingIds: res.data.processing_chapter_ids || []
      }
      return
    }
    showNotification(res.msg || '获取精简状态失败')
  } catch (e) {
    console.warn('[Reader] load chapter trim status failed', e)
  }
}

// 打开指定章节精简弹窗。
const openChapterTrimModal = async () => {
  if (!userStore.isLoggedIn()) {
    openLoginModal('指定章节精简需要登录账号后才能使用，是否现在去登录？')
    return
  }
  if (!activeBook.value) return
  if (activeBook.value.sync_state === 0) {
    showNotification('仅云端书籍支持指定精简')
    return
  }
  // #ifdef APP-PLUS
  if (!networkStore.serverReachable) {
    showNotification('网络不可用，暂不支持精简')
    return
  }
  // #endif
  if (bookStore.prompts.length === 0) {
    await bookStore.fetchPrompts()
  }
  if (bookStore.prompts.length === 0) {
    showNotification('精简模式加载失败')
    return
  }
  chapterTrimPromptId.value = getDefaultTrimPromptId()
  showChapterTrimModal.value = true
  loadPointsBalance()
  loadChapterTrimStatus()
}

// 切换指定章节精简的模式。
const handleChapterTrimPromptChange = (promptId: number) => {
  chapterTrimPromptId.value = promptId
  loadChapterTrimStatus()
}

// 提交指定章节精简任务。
const handleConfirmChapterTrim = (payload: { promptId: number; chapterIds: number[] }) => {
  if (!payload.chapterIds.length) return
  pendingTrimPayload.value = payload
  trimConfirmContent.value = `将精简 ${payload.chapterIds.length} 章，消耗 ${payload.chapterIds.length} 积分，是否继续？`
  showTrimConfirmModal.value = true
}

const submitChapterTrimTask = async () => {
  const payload = pendingTrimPayload.value
  if (!payload || !activeBook.value) return

  showTrimConfirmModal.value = false
  const bookId = activeBook.value.cloud_id || activeBook.value.id
  const res = await taskApi.startChapterTrimTask(bookId, payload.promptId, payload.chapterIds)
  if (res.code === 0) {
    showChapterTrimModal.value = false
    showNotification('任务已创建，可在书架查看进度')
    return
  }

  if (res.code === 6001) {
    showNotification('积分不足')
    return
  }
  if (res.code === 4004) {
    showNotification('章节已精简或处理中')
    return
  }
  showNotification(res.msg || '任务创建失败')
}

const updateUserPreference = async (modeId: number) => {
  userPreferredModeId.value = modeId
  uni.setStorageSync('userPreferredModeId', modeId.toString())

  const prompt = bookStore.prompts.find(p => p.id === modeId)

  if (!activeBook.value || !activeChapter.value) {
    if (prompt) {
      showNotification(`偏好已更新为「${prompt.name}」`)
    }
    return
  }

  const previousModeId = actualDisplayModeId.value
  const hasPreferred = await checkAndAdjustMode(activeChapter.value)

  if (pageMode.value === 'click') {
    await rebuildVirtualWindow(activeBook.value.activeChapterIndex, 'keep')
  } else {
    syncUI()
  }

  if (!hasPreferred) {
    if (prompt) {
      showNotification(`偏好已更新为「${prompt.name}」，当前章节暂无该模式，已显示原文`)
    } else {
      showNotification('当前章节暂无该偏好模式，已显示原文')
    }
    return
  }

  if (prompt) {
    showNotification(`偏好已更新为「${prompt.name}」`)
  }
  if (actualDisplayModeId.value !== previousModeId) {
    showModeSwitchTip(activeChapter.value, actualDisplayModeId.value)
  }
}


const handleTerminalClose = () => {
  showTerminal.value = false
  showNotification('AI 精简将在后台继续...')
}

const watchBatchTask = (taskId: string, bookName: string) => {
  // ... (Keep existing logic)
}

const handleStartProcess = async (modeId: string | number) => {
  // 权限检查：AI 处理功能需要登录
  if (!userStore.isLoggedIn()) {
    showConfigModal.value = false
    openLoginModal('AI 精简功能需要登录账号后才能使用，是否现在去登录？')
    return
  }

  const promptId = typeof modeId === 'string' ? parseInt(modeId) : modeId

  // 单章精简 (混合模式)
  const isTrimmed = activeChapter.value?.trimmed_prompt_ids?.some((id: number) => Number(id) === promptId)
  if (isTrimmed) {
    showConfigModal.value = false
    // 切换到该模式（临时切换，不修改偏好）
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
      activeBook.value?.book_md5 || '',
      activeBook.value?.title || '',
      activeChapter.value?.title || '',
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
        actualDisplayModeId.value = promptId

        if (activeChapter.value) {
          await bookStore.saveChapterTrim(activeBook.value!.id, activeChapter.value!.id, promptId, streamingContent.value)
          if (!activeChapter.value.trimmed_prompt_ids.includes(promptId)) {
            activeChapter.value.trimmed_prompt_ids.push(promptId)
          }
        }

      if (showTerminal.value) {
        setTimeout(() => {
          showTerminal.value = false
          if (pageMode.value === 'click' && activeBook.value) rebuildVirtualWindow(activeBook.value.activeChapterIndex, 'keep')
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
        actualDisplayModeId.value = promptId

        if (activeChapter.value) {
          await bookStore.saveChapterTrim(activeBook.value!.id, activeChapter.value!.id, promptId, streamingContent.value)
          if (!activeChapter.value.trimmed_prompt_ids.includes(promptId)) {
            activeChapter.value.trimmed_prompt_ids.push(promptId)
          }
        }

      if (showTerminal.value) {
        setTimeout(() => {
          showTerminal.value = false
          if (pageMode.value === 'click' && activeBook.value) rebuildVirtualWindow(activeBook.value.activeChapterIndex, 'keep')
          showModeSwitchTip(activeChapter.value, promptId)
        }, 800)
      } else {
        showModeSwitchTip(activeChapter.value, promptId)
      }
      }
    )
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
      actualDisplayModeId.value = promptId

      if (activeChapter.value) {
        await bookStore.saveChapterTrim(activeBook.value!.id, activeChapter.value!.id, promptId, streamingContent.value)
        if (!activeChapter.value.trimmed_prompt_ids.includes(promptId)) {
           activeChapter.value.trimmed_prompt_ids.push(promptId)
        }
      }

      if (showTerminal.value) {
        setTimeout(() => {
          showTerminal.value = false
          if (pageMode.value === 'click' && activeBook.value) rebuildVirtualWindow(activeBook.value.activeChapterIndex, 'keep')
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
            if (pageMode.value === 'click' && activeBook.value) rebuildVirtualWindow(activeBook.value.activeChapterIndex, 'keep')
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
      openLoginModal('本地书籍仅支持阅读原文，切换精简模式需要登录账号。');
      return
    }
  } catch (e) {
    toastStore.show({ message: '系统错误', type: 'error' })
    return
  }

  const promptId = parseInt(id)

  // 更新实际显示模式（不修改用户偏好）
  if (promptId > 0) {
    actualDisplayModeId.value = promptId
  } else {
    actualDisplayModeId.value = 0
  }

  // 尝试从本地缓存加载
  const lines = await bookStore.fetchChapterTrim(activeBook.value!.id, activeChapter.value!.id, promptId)

  if (lines) {
    // 成功切换
    activeBook.value!.activeModeId = id
    isMagicActive.value = true
    syncUI()
    if (pageMode.value === 'click') rebuildVirtualWindow(activeBook.value.activeChapterIndex, 'keep')
    triggerPreload()

    // 显示模式切换提示
    setTimeout(() => {
      showModeSwitchTip(activeChapter.value, promptId)
    }, 100)
  } else {
    // 失败时重置为原文，不弹窗
    actualDisplayModeId.value = 0
    isMagicActive.value = false
    syncUI()
    if (pageMode.value === 'click' && activeBook.value) rebuildVirtualWindow(activeBook.value.activeChapterIndex, 'keep')
    showNotification('暂无离线精简内容')
  }
}

const getMagicButtonClass = () => {
  if (actualDisplayModeId.value > 0) {  // 使用 actualDisplayModeId
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

const checkAndAdjustMode = async (chapter: any) => {
  if (userPreferredModeId.value === 0) {
    actualDisplayModeId.value = 0
    isMagicActive.value = false
    return false
  }

  const promptId = userPreferredModeId.value
  await bookStore.ensureTrimmedStatus(chapter.id)
  const hasTrimmed = chapter.trimmed_prompt_ids?.includes(promptId)

  if (hasTrimmed) {
    const lines = await bookStore.fetchChapterTrim(activeBook.value!.id, chapter.id, promptId)
    if (hasValidLines(lines || [])) {
      actualDisplayModeId.value = promptId
      isMagicActive.value = true
      if (activeBook.value) {
        activeBook.value.activeModeId = promptId.toString()
      }
      return true
    }
    console.warn('[Reader][Mode] Preferred trim empty, fallback original', {
      chapterId: chapter.id,
      promptId
    })
  }

  actualDisplayModeId.value = 0
  isMagicActive.value = false
  return false
}

const toggleMagic = async () => {
  if (actualDisplayModeId.value !== 0) {
    actualDisplayModeId.value = 0
    isMagicActive.value = false
    syncUI()
    if (pageMode.value === 'click' && activeBook.value) rebuildVirtualWindow(activeBook.value.activeChapterIndex, 'keep')
    showNotification('已切换为原文')
    return
  }

  // #ifdef APP-PLUS
  if (!networkStore.serverReachable) {
    showNotification('网络不可用，暂不支持精简')
    return
  }
  // #endif

  if (bookStore.prompts.length === 0) {
    await bookStore.fetchPrompts()
    if (bookStore.prompts.length === 0) {
      showNotification('精简模式加载失败')
      return
    }
  }

  const targetModeId = userPreferredModeId.value
  if (targetModeId === 0) {
    showConfigModal.value = true
    return
  }

  if (activeChapter.value) {
    await bookStore.ensureTrimmedStatus(activeChapter.value.id)
  }

  const hasTrimmed = activeChapter.value?.trimmed_prompt_ids?.includes(targetModeId)
  if (hasTrimmed) {
    await switchToMode(targetModeId.toString(), false)
  } else {
    showConfigModal.value = true
  }
}

const switchChapter = async (index: number, targetPosition: 'start' | 'end' = 'start') => {
  if (index < 0 || index >= activeBook.value!.chapters.length) return

  // 手动切换章节时，如果正在听书，则停止
  if (isTtsSpeaking.value) {
    if (ttsPlayer.value) ttsPlayer.value.stop()
    isTtsSpeaking.value = false
    ttsCurrentIndex.value = -1
  }

  const targetChapter = activeBook.value!.chapters[index]
  const previousActualModeId = actualDisplayModeId.value

  // 检查并调整目标章节的模式（基于用户偏好）
  await checkAndAdjustMode(targetChapter)

  isTextTransitioning.value = true
  clearTimeout(progressTimer)

  if (pageMode.value === 'scroll') {
    scrollTop.value = 1
    nextTick(() => { scrollTop.value = 0 })
  }

  activeBook.value!.activeChapterIndex = index
  markChapterRead(index)

  if (pageMode.value === 'click') {
    await rebuildVirtualWindow(index, targetPosition === 'end' ? 'last' : 'first')
  } else {
    await bookStore.setChapter(index)
    await buildVirtualWindow(index, 'first')
  }

   const chapId = activeBook.value!.chapters[index].id
   triggerPreload()
   handleProgressTracking(chapId)

   if (actualDisplayModeId.value !== previousActualModeId) {
     const promptId = actualDisplayModeId.value
     setTimeout(() => {
       showModeSwitchTip(activeBook.value!.chapters[index], promptId)
     }, 300)
   }
}
</script>

<template>
  <view :style="{ backgroundColor: modeColors[readingMode].bg, color: modeColors[readingMode].text, backgroundImage: readingBgImage ? `url(${readingBgImage})` : 'none', backgroundSize: 'cover', backgroundPosition: 'center' }"
        class="h-screen w-full flex flex-col relative overflow-hidden transition-colors duration-300">
    <view v-if="readingBgImage" class="absolute inset-0 z-0 pointer-events-none" :style="{ backgroundColor: bgOverlayColor }"></view>

    <!-- Top Bar -->
    <view v-if="menuVisible" class="fixed top-0 inset-x-0 z-[80] flex flex-col border-b bg-inherit shadow-sm transition-colors duration-300" :style="{ backgroundColor: modeColors[readingMode].bg }">
      <view :style="{ height: statusBarHeight + 'px' }"></view>
      <view class="h-12 flex items-center justify-between px-4">
        <view @click="handleBack" class="p-2 active:opacity-50 transition-opacity">
          <image :src="icons.back" mode="aspectFit" style="width: 44rpx; height: 44rpx;" class="opacity-70"></image>
        </view>
        <view class="flex flex-col items-center max-w-[200px]">
          <text class="font-bold text-sm truncate max-w-[200px]">{{ activeBook?.title }}</text>
          <text class="text-[10px] opacity-50 mt-0.5">{{ activeModeName }}</text>
        </view>

        <!-- TTS Icon Button -->
        <view @click.stop="toggleTtsPanel" class="p-2 relative group active:opacity-50 transition-opacity">
          <view class="w-8 h-8 rounded-full flex items-center justify-center transition-colors duration-200"
                :class="isTtsSpeaking ? 'bg-teal-500/10' : ''">
            <image :src="icons.headphone" mode="aspectFit" 
                   class="w-5 h-5 transition-all duration-300"
                   :class="[
                     isTtsSpeaking ? 'opacity-100' : 'opacity-70',
                     // 如果是播放状态，使用 CSS filter 染成青色 (teal-500 approx hue-rotate)
                     // 如果是暗黑模式且未播放，反色以显示为白色
                     isTtsSpeaking ? 'sepia hue-rotate-190 saturate-200' : (readingMode === 'dark' ? 'invert' : '')
                   ]" 
                   style="width: 44rpx; height: 44rpx;"></image>
            <view v-if="isTtsSpeaking" class="absolute inset-0 rounded-full border border-teal-500 animate-ping opacity-20"></view>
          </view>
        </view>
      </view>
      
      <!-- TTS Control Panel (Minimalist Modern Style) -->
      <view v-if="showTtsPanel" 
            class="absolute top-14 right-4 w-64 bg-white/80 dark:bg-stone-900/80 backdrop-blur-md shadow-2xl rounded-2xl border border-white/20 dark:border-stone-800/50 z-50 p-4 animate-slide-up-fade">
        
        <!-- Controls Row -->
        <view class="flex items-center justify-between mb-5">
          <view class="flex flex-col">
            <text class="text-[10px] font-bold tracking-widest uppercase opacity-30 mb-1">状态</text>
            <text class="text-xs font-medium" :class="isTtsSpeaking ? 'text-teal-500' : 'text-stone-400'">
              {{ isTtsSpeaking ? '正在朗读...' : '已暂停' }}
            </text>
          </view>
          
          <view class="flex items-center gap-2">
            <!-- Stop -->
            <view @click="stopTTS" class="w-8 h-8 rounded-full flex items-center justify-center bg-stone-200/50 dark:bg-stone-800/50 active:scale-90 transition-transform">
              <view class="w-2.5 h-2.5 bg-stone-500 rounded-[2px]"></view>
            </view>
            <!-- Play/Pause -->
            <view @click="toggleTTS" class="w-10 h-10 rounded-full flex items-center justify-center bg-teal-500 shadow-lg shadow-teal-500/20 active:scale-95 transition-all">
              <view v-if="!isTtsSpeaking" class="ml-0.5 border-l-[10px] border-l-white border-y-[6px] border-y-transparent"></view>
              <view v-else class="flex gap-1">
                <view class="w-1 h-3 bg-white rounded-full"></view>
                <view class="w-1 h-3 bg-white rounded-full"></view>
              </view>
            </view>
          </view>
        </view>
        
        <!-- Speed Control (Slider) -->
        <view class="flex flex-col mb-5">
          <view class="flex justify-between items-center mb-2">
            <text class="text-[10px] font-bold tracking-widest uppercase opacity-30">语速</text>
            <text class="text-xs font-mono font-bold text-teal-600">{{ ttsRate.toFixed(1) }}x</text>
          </view>
          <slider 
            :value="ttsRate" 
            :min="0.5" 
            :max="3.0" 
            :step="0.1" 
            activeColor="#14b8a6" 
            backgroundColor="#e5e7eb" 
            block-size="16"
            @change="(e) => updateTtsRate(e.detail.value)"
            @changing="(e) => ttsRate = e.detail.value"
          />
        </view>

        <!-- Sleep Timer (Slider) -->
        <view class="flex flex-col">
          <view class="flex justify-between items-center mb-2">
            <text class="text-[10px] font-bold tracking-widest uppercase opacity-30">定时关闭</text>
            <text class="text-xs font-mono font-bold text-teal-600">{{ getSleepLabel(ttsSleepValue) }}</text>
          </view>
          <slider 
            :value="ttsSleepValue" 
            :min="0" 
            :max="90" 
            :step="5" 
            activeColor="#14b8a6" 
            backgroundColor="#e5e7eb" 
            block-size="16"
            @change="handleSleepSliderChange"
            @changing="(e) => ttsSleepValue = e.detail.value"
          />
          <view class="flex justify-between px-1 mt-1">
             <text class="text-[10px] text-stone-300">关</text>
             <text class="text-[10px] text-stone-300">30分</text>
             <text class="text-[10px] text-stone-300">60分</text>
             <text class="text-[10px] text-stone-300">90分</text>
          </view>
        </view>
      </view>
    </view>

    <!-- Main Viewport -->
    <view class="flex-1 min-h-0 w-full relative z-10" @click="handleContentClick">
      
      <!-- 1. Scroll Mode -->
      <scroll-view v-if="pageMode === 'scroll'" scroll-y class="h-full scroll-view-content" :scroll-top="scrollTop" @scroll="handleScroll">
        <view class="p-6 pb-32 transition-opacity duration-300" :style="{ fontSize: fontSize + 'px', paddingTop: (statusBarHeight + 60) + 'px' }" :class="{ 'opacity-0': isTextTransitioning }">
          <view class="text-2xl font-bold mb-10 text-center">{{ activeChapter?.title }}</view>
          
          <!-- Explicit UI Binding -->
          <view v-for="(para, idx) in currentTextLines" :key="idx" 
                class="content-para indent-8 leading-loose text-justify transition-colors duration-300"
                :class="{ 'text-teal-500 font-medium': ttsCurrentIndex === idx }">
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
        :current="activePageIndex"
        :duration="300">
        <!-- 核心阅读区域：使用 pageItem.key 确保唯一性 -->
        <swiper-item v-for="(pageItem, pIdx) in combinedPages" :key="pageItem.key">
          <view class="p-6 h-full flex flex-col transition-opacity duration-300" :style="{ fontSize: fontSize + 'px', paddingTop: (statusBarHeight + 60) + 'px' }" :class="{ 'opacity-0': isTextTransitioning }">
            <view class="h-full flex flex-col">
              <view v-if="isFirstPageOfChapter(pIdx)" class="text-2xl font-bold mb-10 text-center">
                {{ getPageTitle(pIdx) }}
              </view>
              <view v-for="(group, gIdx) in getParagraphGroups(pageItem.content)" :key="gIdx">
                <view v-for="(line, idx) in group.lines" :key="idx"
                      class="leading-loose text-justify transition-colors duration-300"
                      :class="{
                        'indent-8': line.isFirstLine,
                        'text-teal-500 font-medium': ttsCurrentIndex !== -1 && line.paraIndex === ttsCurrentIndex
                      }">
                  {{ line.text }}
                </view>
              </view>
            </view>
          </view>
        </swiper-item>
      </swiper>
    </view>

    <!-- Invisible Canvas for measurement -->
    <canvas canvas-id="measure-canvas" class="fixed top-0 left-0 w-0 h-0 opacity-0 pointer-events-none"></canvas>
    <view class="fixed -left-[9999px] -top-[9999px] opacity-0 pointer-events-none">
      <view class="line-height-probe leading-loose" :style="{ fontSize: fontSize + 'px' }">测</view>
    </view>
    <Renderjs :rjsChapter="renderjsPayload" @handelContentUpdated="updateContainerContent" />

    <!-- Overlays -->
    <view v-if="pageMode === 'click'" class="fixed bottom-3 right-6 opacity-40 z-10">
      <text class="text-[10px] font-bold">{{ relativePageInfo }}</text>
    </view>

    <!-- Controls -->
    <view v-if="menuVisible" class="fixed bottom-24 right-6 z-40">
      <view @click.stop="toggleMagic" @longpress="showConfigModal = true"
        :class="getMagicButtonClass()"
        class="w-14 h-14 rounded-full flex items-center justify-center shadow-xl active:scale-90 transition-all select-none">
        <image 
          v-if="!isAiLoading" 
          src="/static/icons/sparkles.svg" 
          class="w-7 h-7 transition-opacity duration-300" 
          :class="isMagicActive ? 'opacity-100 invert brightness-200' : 'opacity-60'" 
        />
        <image 
          v-else 
          src="/static/icons/loading.svg" 
          class="w-6 h-6 animate-spin opacity-60" 
        />
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
        <view @click.stop="openChapterTrimModal" class="flex flex-col items-center gap-1 w-14 active:opacity-50 transition-opacity">
          <image :src="icons.batch" mode="aspectFit" style="width: 44rpx; height: 44rpx;" class="opacity-70"></image>
          <text class="text-[10px] text-stone-400">指定精简</text>
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
    <ChapterList :show="showChapterList" :chapters="activeBook?.chapters || []" :active-chapter-index="activeBook?.activeChapterIndex || 0" :active-mode-id="activeBook?.activeModeId" :read-indexes="readChapterIndexes" :reading-mode="readingMode" :mode-colors="modeColors" @close="showChapterList = false" @select="(idx) => { showChapterList = false; switchChapter(idx) }" />
    <ChapterTrimModal
      :show="showChapterTrimModal"
      :prompts="bookStore.prompts"
      :chapters="chapterTrimOptions"
      :balance="pointsBalance"
      :current-chapter-id="activeChapter?.cloud_id || activeChapter?.id || 0"
      :preferred-mode-id="userPreferredModeId"
      :reading-mode="readingMode"
      @close="showChapterTrimModal = false"
      @confirm="handleConfirmChapterTrim"
      @change-prompt="handleChapterTrimPromptChange"
    />
    <SimpleAlertModal
      :visible="showTrimConfirmModal"
      title="确认精简"
      :content="trimConfirmContent"
      confirm-text="开始精简"
      show-cancel
      cancel-text="再想想"
      @update:visible="showTrimConfirmModal = $event"
      @confirm="submitChapterTrimTask"
      @cancel="pendingTrimPayload = null"
    />
    <ModeConfigModal :show="showConfigModal" :book-title="activeBook?.title || ''" :chapter-title="activeChapter?.title || ''" :prompts="bookStore.prompts" :trimmed-ids="activeChapter?.trimmed_prompt_ids || []" :reading-mode="readingMode" :mode-colors="modeColors" :user-preferred-mode-id="userPreferredModeId" @close="showConfigModal = false" @start="handleStartProcess" />
    <SettingsPanel :show="showSettings" :modes="bookStore.prompts.map(p => p.id.toString())" :prompts="bookStore.prompts" :active-mode="activeBook?.activeModeId || ''" :font-size="fontSize" :reading-mode="readingMode" :mode-colors="modeColors" :page-mode="pageMode" :hide-status-bar="hideStatusBar" :user-preferred-mode-id="userPreferredModeId" :reading-bg-image="readingBgImage" @close="showSettings = false" @update:active-mode="switchToMode" @update:font-size="fontSize = $event" @update:reading-mode="(val) => { readingMode = val; uni.setStorageSync('readingMode', val) }" @update:page-mode="(val) => { pageMode = val; uni.setStorageSync('pageMode', val) }" @update:hide-status-bar="(val) => { hideStatusBar = val; uni.setStorageSync('hideStatusBar', val ? 'true' : 'false') }" @update:user-preferred-mode-id="updateUserPreference" @select:bgImage="chooseBackgroundImage" @clear:bgImage="clearBackgroundImage" />
    <GenerationTerminal :show="showTerminal" :content="streamingContent" :title="generatingTitle" :reading-mode="readingMode" :mode-colors="modeColors" @close="handleTerminalClose" />
    <LoginConfirmModal v-model:visible="showLoginModal" :content="loginTipContent" :reading-mode="readingMode" @confirm="handleLoginConfirm" />
    <view v-if="showToast" class="fixed bottom-40 left-1/2 -translate-x-1/2 bg-stone-900 text-white px-4 py-2 rounded-full text-xs z-[110] shadow-2xl">{{ toastMsg }}</view>
  </view>
</template>

<style>
.pb-safe { padding-bottom: env(safe-area-inset-bottom); }
::-webkit-scrollbar { display: none; width: 0; height: 0; color: transparent; }

@keyframes slideUpFade {
  from { transform: translateY(10px); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}
.animate-slide-up-fade {
  animation: slideUpFade 0.25s cubic-bezier(0.2, 0.8, 0.2, 1) forwards;
}
</style>
