import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api, type Book as ApiBook, type Chapter as ApiChapter, type Prompt } from '../api'

export interface ChapterContent {
  [mode: string]: string[];
}

export interface Chapter extends ApiChapter {
  modes: ChapterContent; // Cached content split by lines
  isLoaded: boolean;
  trimmed_prompt_ids: number[]; // Local cache of available trims
}

export interface Book {
  id: number
  title: string
  status: 'new' | 'processing' | 'ready'
  activeModeId?: string
  book_trimmed_ids?: number[]
  chapters: Chapter[]
  activeChapterIndex: number
}

export const useBookStore = defineStore('book', () => {
  const books = ref<Book[]>([])
  const activeBook = ref<Book | null>(null)
  const prompts = ref<Prompt[]>([])
  const isLoading = ref(false)

  // --- Actions ---

  // 0. Fetch Prompts
  const fetchPrompts = async () => {
    try {
      const res = await api.getPrompts()
      if (res.code === 0) {
        prompts.value = res.data
      }
    } catch (e) {
      console.error('Failed to fetch prompts', e)
    }
  }

  // 1. Fetch Book List
  const fetchBooks = async () => {
    isLoading.value = true
    try {
      const res = await api.getBooks()
      if (res.code === 0) {
        // Map API Book to UI Book (Add null check for res.data)
        const data = res.data || []
        books.value = data.map(b => ({
          ...b,
          progress: 0,
          status: 'ready',
          activeChapterIndex: 0,
          chapters: []
        }))
      }
    } catch (e) {
      console.error('Failed to fetch books', e)
    } finally {
      isLoading.value = false
    }
  }

  // 2. Upload Book (Uni-app handled via Renderjs in shelf.vue, this is fallback)
  const uploadBook = async (filePath: string, fileName: string) => {
    isLoading.value = true
    try {
      const res = await api.upload(filePath, fileName)
      if (res.code === 0) {
        await fetchBooks()
      }
      return res
    } finally {
      isLoading.value = false
    }
  }

  // 3. Fetch Book Detail (Metadata + Chapter List)
  const fetchBookDetail = async (bookId: number, promptId: number = 2) => {
    isLoading.value = true
    try {
      const res = await api.getBookDetail(bookId, promptId)
      if (res.code === 0) {
        const data = res.data
        
        // Find existing book or create temp
        let book = books.value.find(b => b.id === bookId)
        if (!book) {
          book = {
            ...data.book,
            progress: 0,
            status: 'ready',
            activeChapterIndex: 0,
            chapters: []
          }
          books.value.push(book)
        }

        // Map Chapters
        book.chapters = data.chapters.map(c => ({
          ...c,
          modes: {},
          isLoaded: false,
          trimmed_prompt_ids: c.trimmed_prompt_ids || []
        }))
        
        book.book_trimmed_ids = data.book_trimmed_ids || []
        
        // Restore history
        if (data.reading_history) {
           const idx = book.chapters.findIndex(c => c.id === data.reading_history?.last_chapter_id)
           if (idx >= 0) book.activeChapterIndex = idx
           if (data.reading_history.last_prompt_id > 0) {
             book.activeModeId = data.reading_history.last_prompt_id.toString()
           }
        }

        activeBook.value = book
        
        if (book.chapters.length > 0) {
          await fetchChapter(bookId, book.chapters[book.activeChapterIndex].id)
        }
      }
    } catch (e) {
      console.error('Failed to fetch book detail', e)
    } finally {
      isLoading.value = false
    }
  }

  // 4. Fetch Specific Chapter Content (Raw Only)
  const fetchChapter = async (bookId: number, chapterId: number) => {
    if (!activeBook.value) return
    const chapter = activeBook.value.chapters.find(c => c.id === chapterId)
    if (!chapter) return

    // 先读本地缓存
    const cacheKey = `book_${bookId}_chap_${chapterId}`
    const cached = uni.getStorageSync(cacheKey)
    if (cached && cached.modes?.original) {
      chapter.modes = { ...chapter.modes, ...cached.modes }
      chapter.trimmed_prompt_ids = cached.trimmed_prompt_ids || []
      chapter.isLoaded = true
      // 如果已经有缓存了，可以选择直接返回，或者后台静默更新
      // return 
    }

    try {
      const res = await api.getChapter(chapterId)
      if (res.code === 0) {
        const data = res.data
        chapter.modes['original'] = data.content ? data.content.split('\n') : []
        chapter.trimmed_prompt_ids = data.trimmed_prompt_ids || []
        chapter.isLoaded = true
        
        // 存入本地
        uni.setStorageSync(cacheKey, {
          modes: chapter.modes,
          trimmed_prompt_ids: chapter.trimmed_prompt_ids
        })
      }
    } catch (e) {
      console.error('Failed to fetch chapter', e)
    }
  }

  // 5. Fetch Trimmed Content Only
  const fetchChapterTrim = async (chapterId: number, promptId: number): Promise<boolean> => {
    if (!activeBook.value) return false
    const chapter = activeBook.value.chapters.find(c => c.id === chapterId)
    if (!chapter) return false

    try {
      const res = await api.getChapterTrim(chapterId, promptId)
      if (res.code === 0 && res.data.trimmed_content) {
        chapter.modes[promptId.toString()] = res.data.trimmed_content.split('\n')
        
        // 更新本地存储
        const cacheKey = `book_${activeBook.value.id}_chap_${chapterId}`
        uni.setStorageSync(cacheKey, {
          modes: chapter.modes,
          trimmed_prompt_ids: chapter.trimmed_prompt_ids
        })
        return true
      }
    } catch (e) {
      console.error('Failed to fetch trim', e)
    }
    return false
  }

  // 6. Switch Chapter
  const setChapter = async (index: number) => {
    if (!activeBook.value) return
    if (index >= 0 && index < activeBook.value.chapters.length) {
      activeBook.value.activeChapterIndex = index
      activeBook.value.progress = Math.floor((index / activeBook.value.chapters.length) * 100)
      
      const chapter = activeBook.value.chapters[index]
      if (!chapter.isLoaded) {
        await fetchChapter(activeBook.value.id, chapter.id)
      }
    }
  }

  const setActiveBook = (bookId: number) => {
    activeBook.value = books.value.find(b => b.id === bookId) || null
  }

  const updateBookStatus = (bookId: number, status: 'new' | 'processing' | 'ready') => {
    const book = books.value.find(b => b.id === bookId)
    if (book) book.status = status
  }

  // 7. 更新进度到后端
  const updateProgress = async (bookId: number, chapterId: number, promptId: number) => {
    try {
      await api.updateProgress(bookId, chapterId, promptId)
    } catch (e) {
      console.error('Failed to sync progress', e)
    }
  }

  // 8. 批量获取章节并缓存
  const fetchBatchChapters = async (chapterIds: number[], promptId?: number) => {
    if (!activeBook.value) return
    const bookId = activeBook.value.id
    
    try {
      const res = await api.getBatchChapters(chapterIds, promptId)
      if (res.code === 0) {
        res.data.forEach(item => {
          const chapter = activeBook.value?.chapters.find(c => c.id === item.id)
          if (chapter) {
            chapter.modes['original'] = item.content ? item.content.split('\n') : []
            if (item.trimmed_content) {
              chapter.modes[promptId?.toString() || ''] = item.trimmed_content.split('\n')
              if (!chapter.trimmed_prompt_ids.includes(Number(promptId))) {
                chapter.trimmed_prompt_ids.push(Number(promptId))
              }
            }
            chapter.isLoaded = true
            
            // 写入本地持久化
            uni.setStorageSync(`book_${bookId}_chap_${item.id}`, {
              modes: chapter.modes,
              trimmed_prompt_ids: chapter.trimmed_prompt_ids
            })
          }
        })
      }
    } catch (e) {
      console.error('Batch fetch failed', e)
    }
  }
  
  const activeChapter = computed(() => {
    if (!activeBook.value) return null
    return activeBook.value.chapters[activeBook.value.activeChapterIndex]
  })

  return { 
    books, activeBook, activeChapter, prompts, isLoading,
    fetchPrompts, fetchBooks, uploadBook, fetchBookDetail, 
    fetchChapter, fetchChapterTrim, fetchBatchChapters, updateProgress,
    setActiveBook, setChapter, updateBookStatus 
  }
})