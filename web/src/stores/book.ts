import { defineStore } from 'pinia'
import { ref } from 'vue'
import { type Chapter } from '../api'

export const useBookStore = defineStore('book', () => {
  const currentBookId = ref<number>(0)
  const bookTitle = ref('')
  const chapters = ref<Chapter[]>([])
  const trimmedIDs = ref<number[]>([]) // 已精简的章节 ID 列表
  const lastReadChapterID = ref<number | null>(null)
  const lastPromptID = ref<number | null>(null)
  
  function setBook(id: number, title: string, chaps: Chapter[], trimmed?: number[], history?: any) {
    currentBookId.value = id
    bookTitle.value = title
    chapters.value = chaps
    trimmedIDs.value = trimmed || []
    if (history) {
      lastReadChapterID.value = history.last_chapter_id
      lastPromptID.value = history.last_prompt_id
    } else {
      lastReadChapterID.value = null
      lastPromptID.value = null
    }
  }

  function updateChapterContent(index: number, content: string, trimmed?: string) {
    if (chapters.value[index]) {
      chapters.value[index].content = content
      if (trimmed) {
        chapters.value[index].trimmed_content = trimmed
      }
    }
  }

  // 标记某一章已精简
  function markChapterTrimmed(chapterID: number) {
    if (!trimmedIDs.value.includes(chapterID)) {
      trimmedIDs.value.push(chapterID)
    }
  }

  return { 
    currentBookId, bookTitle, chapters, trimmedIDs, 
    lastReadChapterID, lastPromptID,
    setBook, updateChapterContent, markChapterTrimmed 
  }
})