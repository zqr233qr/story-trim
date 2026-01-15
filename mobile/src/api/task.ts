import { request } from './index'
import type { Response } from './index'

export interface TaskProgress {
  status: 'pending' | 'running' | 'completed' | 'failed'
  progress: number
  error?: string
}

export interface FullTrimStatus {
  has_full_trim: boolean
  task_id?: string
  status?: string
  progress?: number
  prompt_id?: number
}

export interface ActiveTasksCount {
  has_active: boolean
}

export const taskApi = {
  // 启动全文精简任务
  startFullTrim: (bookId: number, promptId: number): Promise<Response<{ task_id: string }>> => {
    return request({
      url: '/tasks/full-trim',
      method: 'POST',
      data: { book_id: bookId, prompt_id: promptId }
    })
  },

  // 获取任务进度
  getTaskProgress: (taskId: string): Promise<Response<TaskProgress>> => {
    return request({
      url: `/tasks/${taskId}/progress`,
      method: 'GET'
    })
  },

  // 获取书籍全文精简状态
  getBookFullTrimStatus: (bookId: number): Promise<Response<FullTrimStatus>> => {
    return request({
      url: `/books/${bookId}/full-trim-status`,
      method: 'GET'
    })
  },

  // 获取用户所有活跃任务
  getActiveTasks: (): Promise<Response<any[]>> => {
    return request({
      url: '/tasks/active',
      method: 'GET'
    })
  },

  // 检查用户是否有活跃任务
  getActiveTasksCount: (): Promise<Response<ActiveTasksCount>> => {
    return request({
      url: '/tasks/active/count',
      method: 'GET'
    })
  }
}
