import { defineStore } from 'pinia'
import { ref } from 'vue'
import { type Chapter } from '../api'

export const useBookStore = defineStore('book', () => {
  const currentBookId = ref<number>(0)
  const bookTitle = ref('')
  const chapters = ref<Chapter[]>([])
  
  function setBook(id: number, title: string, chaps: Chapter[]) {
    currentBookId.value = id
    bookTitle.value = title
    chapters.value = chaps
  }

  return { currentBookId, bookTitle, chapters, setBook }
})
