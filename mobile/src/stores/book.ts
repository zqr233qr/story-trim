import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api'
import { getBatchChapterContents, getBatchTrimmedById, syncTrimmedStatusById, updateReadingProgress } from '@/api/content'
import { taskApi } from '@/api/task'
import type { Book, Chapter, Prompt } from '@/api'
import { useUserStore } from './user'
// #ifdef APP-PLUS
import { AppRepository } from '@/adapter/app-repository'
import { db } from '@/utils/sqlite'
const repo = new AppRepository()
// #endif

export type { Book, Chapter, Prompt }

export interface BookUI extends Book {
  status: 'new' | 'processing' | 'ready'
  progress: number
  chapters: any[]
  activeChapterIndex: number
  activeModeId?: string
  book_trimmed_ids?: number[]
  sync_state?: number // 0: Local, 1: Synced, 2: CloudOnly
  cloud_id?: number
  full_trim_status?: string      // 'running' | 'completed' | 'failed'
  full_trim_progress?: number    // 0-100
  full_trim_task_id?: string
  full_trim_prompt_id?: number
  full_trim_available?: boolean
}

export const useBookStore = defineStore('book', () => {
  const books = ref<BookUI[]>([])
  const activeBook = ref<BookUI | null>(null)
  const prompts = ref<Prompt[]>([])
  const isLoading = ref(false)
  const uploadProgress = ref(0)
  const syncProgress = ref(0)

  // 初始化数据库
  const init = async () => {
    // #ifdef APP-PLUS
    await repo.init()
    // #endif
  }

  // 1. Fetch Books (Local First)
  const fetchBooks = async () => {
    isLoading.value = true
    try {
      // #ifdef APP-PLUS
      const userStore = useUserStore()
      if (!userStore.isLoggedIn()) {
        books.value = []
        uni.showToast({
          title: '请登录以恢复您的书籍',
          icon: 'none',
          duration: 3000
        })
        return
      }

      try {
        const res = await api.getBooks()
        if (res.code === 0) {
          const cloudBooks = res.data || []
          console.log('[FetchBooks] Syncing', cloudBooks.length, 'books from cloud')

          for (const cloudBook of cloudBooks) {
            await repo.syncBookFromCloud(cloudBook)
          }
        }
      } catch (e) {
        console.warn('[FetchBooks] Failed to sync books from cloud', e)
      }

      const localBooks = await repo.getBooks()
      books.value = localBooks.map(b => ({
        id: Number(b.id),
        title: b.title,
        total_chapters: b.totalChapters,
        created_at: new Date(b.createdAt).toISOString(),
        status: b.processStatus,
        progress: 0,
        activeChapterIndex: 0,
        chapters: [],
        sync_state: b.syncState || 0
      }))
      // #endif

      // #ifndef APP-PLUS
      const res = await api.getBooks()
      if (res.code === 0) {
        const data = res.data || []
        books.value = data.map(b => ({
          ...b,
          progress: 0,
          status: 'ready',
          activeChapterIndex: 0,
          chapters: []
        }))
      }
      // #endif
    } catch (e) {
      console.error('Failed to fetch books', e)
      // 容错：确保 books 始终是数组
      if (!books.value) books.value = []
    } finally {
      isLoading.value = false
    }
  }

  // 2. Add Book (此方法仅用于 H5/MP，App 端走 RenderJS -> createBookRecord)
  const createBookRecord = async (title: string, total: number, bookMD5: string) => {
    // #ifdef APP-PLUS
    try {
      return await repo.createBook(title, total, bookMD5)
    } catch (e: any) {
      if (e.message && e.message.includes('已存在')) {
        throw new Error(e.message)
      }
      throw new Error('创建书籍失败：' + e.message)
    }
    // #endif
    return 0
  }

  // 3. RenderJS 专用：批量插入章节
  const insertChapters = async (bookId: number, chapters: any[]) => {
    // #ifdef APP-PLUS
    await repo.insertChapters(bookId, chapters)
    // #endif
  }

  // 5. Fetch Book Detail
  const fetchBookDetail = async (bookId: number) => {
    // #ifdef APP-PLUS
    try {
      const book = await repo.getBook(bookId)
      console.log('[Store] Fetched book:', book)
      if (book) {
        const syncState = book.syncState || 0
        
        if (syncState === 2) {
          const cloudBookId = book.cloudId || Number(book.id)
          const res = await api.getBookDetail(cloudBookId)
          if (res.code === 0) {
            activeBook.value = {
              id: Number(book.id),
              cloud_id: cloudBookId,
              title: res.data.book.title,
              total_chapters: res.data.book.total_chapters,
              created_at: res.data.book.created_at,
              status: 'ready',
              progress: 0,
              activeChapterIndex: 0,
              chapters: res.data.chapters.map(c => ({
                id: Number(c.id),
                book_id: Number(c.book_id),
                index: c.index,
                title: c.title,
                md5: c.md5,
                cloud_id: Number(c.id),
                trimmed_prompt_ids: [],
                isLoaded: false,
                modes: {}
              })),
              sync_state: 2
            }

            if (activeBook.value && activeBook.value.chapters.length > 0) {
              try {
                const syncRes = await syncTrimmedStatusById(cloudBookId)
                if (syncRes.code === 0 && syncRes.data.trimmed_map) {
                  const tMap = syncRes.data.trimmed_map as Record<number, number[]>
                  activeBook.value.chapters.forEach(c => {
                    const id = c.cloud_id || c.id
                    if (tMap[id]) {
                      const remoteIds = tMap[id].map((id: number) => Number(id))
                      c.trimmed_prompt_ids = [...new Set([...c.trimmed_prompt_ids, ...remoteIds])]
                    }
                  })
                }
              } catch (e) { console.warn('[Store] Sync trim status by id failed', e) }
            }
          }
        } else {
          const chapters = await repo.getChapters(bookId)
          console.log('[Store] Fetched chapters count:', chapters.length)
          
          activeBook.value = {
            id: Number(book.id),
            cloud_id: book.cloudId,
            title: book.title,
            total_chapters: book.totalChapters,
            created_at: new Date(book.createdAt).toISOString(),
            status: book.processStatus,
            progress: 0,
            activeChapterIndex: 0,
            chapters: chapters.map(c => ({
              id: Number(c.id),
              book_id: Number(c.bookId),
              index: c.index,
              title: c.title,
              md5: c.md5,
              cloud_id: c.cloudId,
              trimmed_prompt_ids: [],
              isLoaded: false,
              modes: {}
            })),
            sync_state: syncState
          }

          if (syncState === 0 && chapters.length > 0) {
            const userStore = useUserStore()
            if (userStore.isLoggedIn()) {
              const md5s = chapters.map(c => c.md5).filter(m => !!m)
              try {
                const syncRes = await api.syncTrimmedStatus(md5s)
                if (syncRes.code === 0 && syncRes.data.trimmed_map) {
                  const tMap = syncRes.data.trimmed_map
                  activeBook.value.chapters.forEach(c => {
                    if (c.md5 && tMap[c.md5]) {
                      const remoteIds = tMap[c.md5].map(id => Number(id))
                      c.trimmed_prompt_ids = [...new Set([...c.trimmed_prompt_ids, ...remoteIds])]
                    }
                  })
                }
              } catch (e) { console.warn('[Store] Sync trim status by md5 failed', e) }
            }
          } else if (syncState === 1 && book.cloudId) {
            try {
              const syncRes = await syncTrimmedStatusById(book.cloudId)
              if (syncRes.code === 0 && syncRes.data.trimmed_map) {
                const tMap = syncRes.data.trimmed_map as Record<number, number[]>
                activeBook.value.chapters.forEach(c => {
                  const id = c.cloud_id || c.id
                  if (tMap[id]) {
                    const remoteIds = tMap[id].map((id: number) => Number(id))
                    c.trimmed_prompt_ids = [...new Set([...c.trimmed_prompt_ids, ...remoteIds])]
                  }
                })
              }
            } catch (e) { console.warn('[Store] Sync trim status by id failed', e) }
          }
        }
      }
    } catch (e) {
      console.error('[Store] Fetch book detail failed', e)
    }
    return
    // #endif
  }

  // 6. Fetch Chapter Content
  const fetchChapter = async (bookId: number, chapterId: number) => {
    if (!activeBook.value) return
    const chapter = activeBook.value.chapters.find(c => c.id === chapterId)
    if (!chapter) return

    const syncState = activeBook.value.sync_state || 0

    // #ifdef APP-PLUS
    if (syncState === 2) {
      const cloudChapterId = chapter.cloud_id || chapterId
      const res = await getBatchChapterContents([cloudChapterId])
      if (res.code === 0 && res.data[0] && res.data[0].content) {
        const content = res.data[0].content
        const lines = content.split('\n')
        chapter.modes['original'] = lines
        chapter.isLoaded = true
        const cacheKey = `chapter_${cloudChapterId}`
        uni.setStorageSync(cacheKey, lines)
      }
    } else {
      const content = await repo.getChapterContent(bookId, chapterId)
      const lines = content.split('\n')
      chapter.modes['original'] = lines
      chapter.isLoaded = true
      const cacheKey = `chapter_${chapterId}`
      uni.setStorageSync(cacheKey, lines)
    }
    // #endif
  }

  // 辅助 Actions
  const setActiveBook = (book: BookUI) => { activeBook.value = book }
  
  const setChapter = async (index: number) => {
    if (activeBook.value && activeBook.value.chapters[index]) {
      activeBook.value.activeChapterIndex = index
      const chapter = activeBook.value.chapters[index]
      if (!chapter.isLoaded) await fetchChapter(activeBook.value.id, chapter.id)
    }
  }

  const fetchPrompts = async () => {
    try {
      const res = await api.getPrompts()
      if (res.code === 0) prompts.value = res.data
    } catch (e) {
      console.warn('Failed to fetch prompts, using default')
      // 离线兜底 Prompt
      if (prompts.value.length === 0) {
        prompts.value = [
          { id: 1, name: '标准精简', description: '去除冗余，保留核心剧情', content: '', is_system: true, version: '1.0' },
          { id: 2, name: '极致浓缩', description: '仅保留主线脉络', content: '', is_system: true, version: '1.0' }
        ]
      }
    }
  }

  // 7. Fetch Trimmed Content (Local Cache Only)
  const fetchChapterTrim = async (bookId: number, chapterId: number, promptId: number): Promise<string[] | null> => {
    if (!activeBook.value) return null
    const chapter = activeBook.value.chapters.find(c => c.id === chapterId)
    if (!chapter) return null

    const modeKey = `mode_${promptId}`
    if (chapter.modes && chapter.modes[modeKey]) return chapter.modes[modeKey]
    if (!chapter.modes) chapter.modes = {}

    const cacheKey = `trim_${chapterId}_${promptId}`
    try {
      const cached = uni.getStorageSync(cacheKey)
      if (cached) {
        chapter.modes[modeKey] = cached
        return cached
      }
    } catch (e) { console.warn('[Store] Read storage cache failed', e) }

    // 从云端获取
    const cloudChapterId = chapter.cloud_id || chapterId
    try {
      const res = await getBatchTrimmedById([cloudChapterId], promptId)
      if (res.code === 0 && res.data[0] && res.data[0].trimmed_content) {
        const lines = res.data[0].trimmed_content.split('\n')
        chapter.modes[modeKey] = lines
        uni.setStorageSync(cacheKey, lines)
        return lines
      }
    } catch (e) { console.warn('[Store] Fetch trim from cloud failed', e) }

    return null
  }

  // 手动保存精简内容 (供 UI 层调用)
  const saveChapterTrim = async (bookId: number, chapterId: number, promptId: number, content: string) => {
    if (!activeBook.value) return
    const chapter = activeBook.value.chapters.find(c => c.id === chapterId)
    if (!chapter) return

    const lines = content.split('\n')
    chapter.modes[`mode_${promptId}`] = lines

    const cacheKey = `trim_${chapterId}_${promptId}`
    uni.setStorageSync(cacheKey, lines)
  }

  // 批量获取章节内容
  const fetchBatchChapters = async (ids: number[], promptId: number) => {
    if (!activeBook.value) return
    const syncState = activeBook.value.sync_state || 0

    // #ifdef APP-PLUS
    if (syncState === 2) {
      const cloudIds = ids.map(id => {
        const chapter = activeBook.value?.chapters.find(c => c.id === id)
        return chapter?.cloud_id || id
      })
      try {
        const res = await getBatchChapterContents(cloudIds)
        if (res.code === 0 && res.data && activeBook.value) {
          for (let idx = 0; idx < res.data.length; idx++) {
            const item = res.data[idx]
            if (item && item.content) {
              const chapter = activeBook.value.chapters.find(c => (c.cloud_id || c.id) === cloudIds[idx])
              if (chapter) {
                const lines = item.content.split('\n')
                chapter.modes['original'] = lines
                chapter.isLoaded = true
                const cacheKey = `chapter_${cloudIds[idx]}`
                uni.setStorageSync(cacheKey, lines)
              }
            }
          }
        }
      } catch (e) { console.warn('[Store] Fetch chapters from cloud failed', ids, e) }
    } else {
      for (const id of ids) {
        const chapter = activeBook.value.chapters.find(c => c.id === id)
        if (!chapter) continue
        if (!chapter.isLoaded) {
          try {
            const content = await repo.getChapterContent(activeBook.value!.id, id)
            const lines = content.split('\n')
            chapter.modes['original'] = lines
            chapter.isLoaded = true
            const cacheKey = `chapter_${id}`
            uni.setStorageSync(cacheKey, lines)
          } catch (e) { console.warn('[Store] Fetch chapter from local failed', id, e) }
        }
      }
    }
    // #endif
  }
  const updateProgress = async (bookId: number, chapterId: number, promptId: number) => {
    if (!activeBook.value) return

    // 1. 立即保存到本地 SQLite
    try {
      // #ifdef APP-PLUS
      await repo.updateProgress(bookId, chapterId, promptId)
      // #endif
    } catch (e) { console.warn('[Progress] Save to local failed', e) }

    // 2. 异步上报云端 (sync_state=1/2)
    const cloudBookId = activeBook.value.cloud_id
    if (cloudBookId) {
      setTimeout(async () => {
        try {
          await updateReadingProgress(cloudBookId, chapterId, promptId)
        } catch (e) { console.warn('[Progress] Sync to cloud failed', e) }
      }, 1000)
    }
  }

  // 上传书籍元数据建立同步 (App端)
  const syncBookToCloud = async (bookId: number) => {
    // #ifdef APP-PLUS
    const book = await repo.getBook(bookId);
    if (!book || book.platform !== 'app') return;

    const total = book.totalChapters;
    if (total === 0) return;

    syncProgress.value = 1;
    let cloudBookId = book.cloudId || 0;
    const BATCH_SIZE = 200;
    let syncedCount = book.syncedCount || 0;

    try {
      while (syncedCount < total) {
        const chunk = await repo.getChaptersBatch(bookId, syncedCount, BATCH_SIZE);
        if (chunk.length === 0) break;

        const payload = {
            book_name: book.title,
            book_md5: book.bookMD5,
            cloud_book_id: cloudBookId || undefined,
            total_chapters: total,
            chapters: chunk.map(c => ({
                local_id: Number(c.id),
                index: c.index,
                title: c.title,
                md5: c.md5,
                content: c.content || '',
                words_count: c.words_count || 0
            }))
        };

        const res = await api.syncLocalBook(payload);
        if (res.code === 0 && res.data) {
            const returnedBookId = res.data.book_id;

            if (!cloudBookId && returnedBookId) {
                cloudBookId = returnedBookId;
            }

            const mappings = res.data.chapter_mappings || [];

            for (const mapping of mappings) {
                await db.execute('UPDATE chapters SET cloud_id = ? WHERE id = ?', [mapping.cloud_id, mapping.local_id]);
            }

            await db.execute('UPDATE books SET cloud_id = ?, sync_state = ?, synced_count = ? WHERE id = ?', [cloudBookId, 1, syncedCount + chunk.length, bookId]);
        } else {
            throw new Error(res.msg || 'Sync failed');
        }

        syncedCount += chunk.length;
        syncProgress.value = Math.floor((syncedCount / total) * 100);
      }

      if (cloudBookId > 0) {
         console.log('[Sync] Book synced to cloud:', cloudBookId);
      }
    } catch (e) {
        console.error('[Sync] Failed:', e);
        throw e;
    } finally {
        setTimeout(() => { syncProgress.value = 0 }, 1000);
    }
    // #endif
  }

  const activeChapter = computed(() => {
    if (!activeBook.value) return null
    return activeBook.value.chapters[activeBook.value.activeChapterIndex]
  })

  // 全文精简任务轮询定时器
  let fullTrimPollTimer: ReturnType<typeof setInterval> | null = null

  // 启动全文精简任务
  const startFullTrimTask = async (bookId: number, promptId: number): Promise<boolean> => {
    const res = await taskApi.startFullTrim(bookId, promptId)
    if (res.code === 0) {
      const book = books.value.find(b => b.id === bookId)
      if (book) {
        book.full_trim_status = 'running'
        book.full_trim_progress = 0
        book.full_trim_task_id = res.data.task_id
        book.full_trim_prompt_id = promptId
      }
      // 开始监控进度
      monitorFullTrimTask(res.data.task_id, bookId)
    }
    return res.code === 0
  }

  // 监控任务进度（每5秒）
  const monitorFullTrimTask = (taskId: string, bookId: number) => {
    if (fullTrimPollTimer) clearInterval(fullTrimPollTimer)

    fullTrimPollTimer = setInterval(async () => {
      const res = await taskApi.getTaskProgress(taskId)
      if (res.code !== 0) return

      const book = books.value.find(b => b.id === bookId)
      if (!book) {
        clearInterval(fullTrimPollTimer)
        return
      }

      book.full_trim_status = res.data.status
      book.full_trim_progress = res.data.progress

      if (res.data.status === 'completed') {
        book.full_trim_available = true
        book.full_trim_status = undefined
        book.full_trim_progress = undefined
        clearInterval(fullTrimPollTimer)
        fullTrimPollTimer = null
      } else if (res.data.status === 'failed') {
        // 失败只记录error，不显示UI状态
        book.full_trim_status = undefined
        book.full_trim_progress = undefined
        clearInterval(fullTrimPollTimer)
        fullTrimPollTimer = null
      }
    }, 5000)
  }

  return {
    books, activeBook, prompts, isLoading, uploadProgress, syncProgress, activeChapter,
    init, fetchBooks, fetchBookDetail, fetchChapter,
    createBookRecord, insertChapters, saveChapterTrim,
    setActiveBook, setChapter, fetchPrompts,
    fetchChapterTrim, fetchBatchChapters, updateProgress, syncBookToCloud,
    startFullTrimTask, monitorFullTrimTask
  }
})