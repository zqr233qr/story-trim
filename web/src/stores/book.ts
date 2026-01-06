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

export interface Book extends ApiBook {
  progress: number;
  status: 'new' | 'processing' | 'ready';
  activeModeId?: string;
  activeChapterIndex: number;
  chapters: Chapter[];
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
      if (res.data.code === 0) {
        prompts.value = res.data.data
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
      if (res.data.code === 0) {
        // Map API Book to UI Book
        books.value = res.data.data.map(b => ({
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

  // 2. Upload Book
  const uploadBook = async (file: File) => {
    isLoading.value = true
    try {
      const res = await api.upload(file)
      if (res.data.code === 0) {
        // Refresh list
        await fetchBooks()
      }
      return res.data
    } finally {
      isLoading.value = false
    }
  }

  // 3. Fetch Book Detail (Metadata + Chapter List)
  const fetchBookDetail = async (bookId: number, promptId: number = 2) => {
    isLoading.value = true
    try {
      const res = await api.getBookDetail(bookId, promptId)
      if (res.data.code === 0) {
        const data = res.data.data
        
        // Find existing book or create temp
        let book = books.value.find(b => b.id === bookId)
        if (!book) {
          // If not in list (direct link access), reconstruct basics
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
          modes: {}, // Content loaded on demand
          isLoaded: false
        }))
        
        // Restore history if exists
        if (data.reading_history) {
           const idx = book.chapters.findIndex(c => c.id === data.reading_history?.last_chapter_id)
           if (idx >= 0) book.activeChapterIndex = idx
           
           // Restore active mode from history (which now includes system default fallback)
           if (data.reading_history.last_prompt_id > 0) {
             book.activeModeId = data.reading_history.last_prompt_id.toString()
           }
        }

        activeBook.value = book
        
        // Fetch current chapter content immediately
        if (book.chapters.length > 0) {
          await fetchChapter(bookId, book.chapters[book.activeChapterIndex].id, promptId)
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

    try {
      const res = await api.getChapter(chapterId)
      if (res.data.code === 0) {
        const data = res.data.data
        // Parse raw content
        chapter.modes['original'] = data.content ? data.content.split('\n') : []
        chapter.trimmed_prompt_ids = data.trimmed_prompt_ids || []
        chapter.isLoaded = true
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
      if (res.data.code === 0 && res.data.data.trimmed_content) {
        chapter.modes[promptId.toString()] = res.data.data.trimmed_content.split('\n')
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
    // Just sets the pointer, View should trigger fetch
    activeBook.value = books.value.find(b => b.id === bookId) || null
  }

  const updateBookStatus = (bookId: number, status: 'new' | 'processing' | 'ready') => {
    const book = books.value.find(b => b.id === bookId)
    if (book) book.status = status
  }
  
  // Computed
  const activeChapter = computed(() => {
    if (!activeBook.value) return null
    return activeBook.value.chapters[activeBook.value.activeChapterIndex]
  })

  return { 
    books, 
    activeBook, 
    activeChapter, 
    prompts,
    isLoading,
    fetchPrompts,
    fetchBooks, 
    uploadBook, 
    fetchBookDetail, 
    fetchChapter,
    fetchChapterTrim,
    setActiveBook, 
    setChapter, 
    updateBookStatus 
  }
})
