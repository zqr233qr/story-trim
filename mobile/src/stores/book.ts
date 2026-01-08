import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api'
import type { Book, Chapter, Prompt } from '@/api'
// #ifdef APP-PLUS
import { AppRepository } from '@/adapter/app-repository'
const repo = new AppRepository()
// #endif

export interface BookUI extends Book {
  status: 'new' | 'processing' | 'ready'
  progress: number
  chapters: any[]
  activeChapterIndex: number
  activeModeId?: string
  book_trimmed_ids?: number[]
}

export const useBookStore = defineStore('book', () => {
  const books = ref<BookUI[]>([])
  const activeBook = ref<BookUI | null>(null)
  const prompts = ref<Prompt[]>([])
  const isLoading = ref(false)
  const uploadProgress = ref(0)

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
      const localBooks = await repo.getBooks()
      books.value = localBooks.map(b => ({
        id: Number(b.id),
        title: b.title,
        total_chapters: b.totalChapters,
        fingerprint: b.fingerprint,
        created_at: new Date(b.createdAt).toISOString(),
        status: b.processStatus,
        progress: 0,
        activeChapterIndex: 0,
        chapters: []
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
  const addBook = async (filePath: string, fileName: string) => {
    isLoading.value = true
    try {
      // #ifndef APP-PLUS
      const res = await api.upload(filePath, fileName)
      if (res.code === 0) await fetchBooks()
      return res
      // #endif
    } finally {
      isLoading.value = false
    }
  }

  // 3. RenderJS 专用：创建书籍记录
  const createBookRecord = async (title: string, fingerprint: string, total: number) => {
    // #ifdef APP-PLUS
    return await repo.createBook(title, fingerprint, total)
    // #endif
    return 0
  }

  // 4. RenderJS 专用：批量插入章节
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
        const chapters = await repo.getChapters(bookId)
        console.log('[Store] Fetched chapters count:', chapters.length)
        
        activeBook.value = {
          id: Number(book.id),
          title: book.title,
          total_chapters: book.totalChapters,
          fingerprint: book.fingerprint,
          created_at: new Date(book.createdAt).toISOString(),
          status: book.processStatus,
          progress: 0,
          activeChapterIndex: 0,
          chapters: chapters.map(c => ({
            id: Number(c.id),
            book_id: Number(c.bookId),
            index: c.index,
            title: c.title,
            md5: c.md5, // 关键：用于缓存查找
            trimmed_prompt_ids: [],
            isLoaded: false,
            modes: {}
          }))
        }

        // --- 同步云端精简状态 ---
        if (activeBook.value && chapters.length > 0) {
          const md5s = chapters.map(c => c.md5).filter(m => !!m);
          try {
            const syncRes = await api.syncTrimmedStatus(md5s);
            if (syncRes.code === 0 && syncRes.data.trimmed_map) {
              const tMap = syncRes.data.trimmed_map;
              activeBook.value.chapters.forEach(c => {
                if (c.md5 && tMap[c.md5]) {
                  // 合并云端状态
                  const remoteIds = tMap[c.md5].map(id => Number(id));
                  c.trimmed_prompt_ids = [...new Set([...c.trimmed_prompt_ids, ...remoteIds])];
                }
              });
            }
          } catch (e) { console.warn('[Store] Sync trim status failed', e) }
        }

        if (chapters.length > 0) {
          // 预加载第一章
          await fetchChapter(bookId, Number(chapters[0].id))
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

    // #ifdef APP-PLUS
    const content = await repo.getChapterContent(bookId, chapterId)
    chapter.modes['original'] = content.split('\n')
    chapter.isLoaded = true
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

    // #ifdef APP-PLUS
    try {
      if (chapter.md5) {
        const cached = await repo.getTrimmedContent(chapter.md5, promptId)
        if (cached && cached.content) {
          const lines = cached.content.split('\n')
          chapter.modes[modeKey] = lines
          return lines
        }
      }
    } catch (e) { console.warn(e) }
    // #endif

    return null
  }

  // 手动保存精简内容 (供 UI 层调用)
  const saveChapterTrim = async (bookId: number, chapterId: number, promptId: number, content: string) => {
    if (!activeBook.value) return
    const chapter = activeBook.value.chapters.find(c => c.id === chapterId)
    if (!chapter) return

    const lines = content.split('\n')
    chapter.modes[`mode_${promptId}`] = lines
    
    // #ifdef APP-PLUS
    if (chapter.md5) {
      await repo.saveTrimmedContent(chapter.md5, promptId, content)
    }
    // #endif
  }

  // 占位函数 (修复 ReferenceError)
  const fetchBatchChapters = async (ids: number[], promptId: number) => {}
  const updateProgress = async (bookId: number, chapterId: number, promptId: number) => {}

  const activeChapter = computed(() => {
    if (!activeBook.value) return null
    return activeBook.value.chapters[activeBook.value.activeChapterIndex]
  })

  return {
    books, activeBook, prompts, isLoading, uploadProgress, activeChapter,
    init, fetchBooks, addBook, fetchBookDetail, fetchChapter,
    createBookRecord, insertChapters, saveChapterTrim, // 导出新方法
    setActiveBook, setChapter, fetchPrompts,
    fetchChapterTrim, fetchBatchChapters, updateProgress
  }
})