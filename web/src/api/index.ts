import axios from 'axios';

// 基础配置
const BASE_URL = import.meta.env.PROD ? '/api/v1' : 'http://localhost:8080/api/v1';

const client = axios.create({
  baseURL: BASE_URL,
  timeout: 300000, // 5分钟超时 (适应大文件上传)
});

// 请求拦截器：注入 JWT
client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 响应拦截器：统一处理错误码
client.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// --- 类型定义 ---

export interface Response<T> {
  code: number;
  msg: string;
  data: T;
}

export interface User {
  id: number;
  username: string;
  token?: string;
}

export interface Book {
  id: number;
  title: string;
  total_chapters: number;
  fingerprint: string;
  created_at: string;
}

export interface Chapter {
  id: number;
  book_id: number;
  index: number;
  title: string;
  content?: string; // 详情时才有
  trimmed_content?: string; // 详情时才有
}

export interface ReadingHistory {
  last_chapter_id: number;
  last_prompt_id: number;
}

export interface BookDetail {
  book: Book;
  chapters: Chapter[];
  trimmed_ids: number[]; // 用户已精简的章节ID列表
  reading_history?: ReadingHistory;
}

export interface Prompt {
  id: number;
  name: string;
  version: string;
  content: string;
  is_system: boolean;
}

export interface Task {
  id: string;
  type: string;
  status: string; // pending, running, completed
  progress: number;
  error?: string;
}

// --- API 方法 ---

export const api = {
  // Auth
  register: (data: any) => client.post<Response<void>>('/auth/register', data),
  login: (data: any) => client.post<Response<{ token: string }>>('/auth/login', data),

  // Story
  upload: (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return client.post<Response<Book>>('/upload', formData);
  },

  getBooks: () => client.get<Response<Book[]>>('/books'),
  
  getBookDetail: (id: number, promptId?: number) => 
    client.get<Response<BookDetail>>(`/books/${id}`, { params: { prompt_id: promptId } }),
  
  getChapter: (id: number, promptId?: number) => 
    client.get<Response<Chapter>>(`/chapters/${id}`, { params: { prompt_id: promptId } }),
  
  getPrompts: () => client.get<Response<Prompt[]>>('/prompts'),

  // Tasks
  startBatchTrim: (bookId: number, promptId: number) => 
    client.post<Response<{ task_id: string }>>('/tasks/batch-trim', { book_id: bookId, prompt_id: promptId }),
  
  getTaskStatus: (taskId: string) => 
    client.get<Response<Task>>(`/tasks/${taskId}`),

  // SSE Stream: Trim
  trimStream: async (
    chapterId: number,
    promptId: number,
    onData: (text: string) => void,
    onError: (err: string) => void,
    onDone: () => void
  ) => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${BASE_URL}/trim/stream`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : '',
        },
        body: JSON.stringify({
          chapter_id: chapterId,
          prompt_id: promptId,
        }),
      });

      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
      
      const reader = response.body?.getReader();
      if (!reader) throw new Error('No response body');

      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() ?? ''; 

        for (const line of lines) {
          if (line.startsWith('data:')) {
            let data = line.substring(5);
            if (data.startsWith(' ')) data = data.substring(1); // 去除第一个空格
            onData(data);
          }
        }
      }
      onDone();
    } catch (err: any) {
      onError(err.message || 'Network Error');
    }
  }
};
