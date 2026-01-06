import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type Chapter, type ReadingHistory } from '../api'

export const useBookStore = defineStore('book', () => {
  const currentBookId = ref<number>(0)
  const bookTitle = ref('')
  const chapters = ref<Chapter[]>([])
  const trimmedIDs = ref<number[]>([])
  const lastReadInfo = ref<ReadingHistory | undefined>(undefined)
  
  // 任务状态
  const activeTaskId = ref<string>('')
  const taskProgress = ref<number>(0)
  const isTaskRunning = ref(false)
  let pollTimer: number | null = null

  function setBook(id: number, title: string, chaps: Chapter[], trimmed: number[], history?: ReadingHistory) {
    currentBookId.value = id
    bookTitle.value = title
    chapters.value = chaps
    trimmedIDs.value = trimmed || []
    lastReadInfo.value = history
  }

  function markChapterTrimmed(chapterID: number) {
    if (!trimmedIDs.value.includes(chapterID)) {
      trimmedIDs.value.push(chapterID)
    }
  }

  // 启动后台任务监控
  async function startTaskPolling(taskId: string) {
    activeTaskId.value = taskId
    isTaskRunning.value = true
    
    if (pollTimer) clearInterval(pollTimer)
    
    pollTimer = window.setInterval(async () => {
      try {
        const res = await api.getTaskStatus(taskId)
        if (res.data.code === 0) {
          const task = res.data.data
          taskProgress.value = task.progress
          
          if (task.status === 'completed' || task.status === 'failed' || task.status === 'cancelled') {
            stopTaskPolling()
            // 任务完成后，如果还在当前书，刷新一下已精简列表
            refreshTrimmedStatus()
          }
        }
      } catch (e) {
        stopTaskPolling()
      }
    }, 2000)
  }

  function stopTaskPolling() {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
    isTaskRunning.value = false
  }

  async function refreshTrimmedStatus() {
    if (currentBookId.value === 0) return
    const res = await api.getBookDetail(currentBookId.value)
    if (res.data.code === 0) {
      trimmedIDs.value = res.data.data.trimmed_ids
    }
  }

  return { 
    currentBookId, bookTitle, chapters, trimmedIDs, lastReadInfo,
    activeTaskId, taskProgress, isTaskRunning,
    setBook, markChapterTrimmed, startTaskPolling, stopTaskPolling 
  }
})