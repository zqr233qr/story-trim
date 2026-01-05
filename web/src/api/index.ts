import axios from 'axios';

// 自动判断环境：生产环境使用相对路径，开发环境连接本地 8080
const BASE_URL = import.meta.env.PROD ? '/api' : 'http://localhost:8080/api';

const client = axios.create({
  baseURL: BASE_URL,
  timeout: 300000, 
});

// 请求拦截器注入 Token
client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export interface Chapter {
  id: number;
  book_id: number;
  index: number;
  title: string;
  content: string;
  trimmed_content?: string;
}

export interface Prompt {
  id: number;
  name: string;
  version: string;
  content: string;
  is_system: boolean;
}

export interface UploadResponse {
  book_id: number;
  filename: string;
  chapters: Chapter[];
  total: number;
}

export const api = {
  // Auth
  register: (data: any) => client.post('/auth/register', data),
  login: (data: any) => client.post('/auth/login', data),

  // Story
  upload: (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return client.post<any>('/upload', formData);
  },

  getBooks: () => client.get('/books'),
  getBookDetail: (id: number) => client.get(`/books/${id}`),
  getChapter: (id: number, promptId?: number, version?: string) => 
    client.get(`/chapters/${id}`, { params: { prompt_id: promptId, version } }),
  getPrompts: () => client.get('/prompts'),
  
  trim: (params: { chapter_id: number; prompt_id?: number; prompt_version?: string }) => {
    return client.post<any>('/trim', params);
  },

  // SSE Stream with Auth
  trimStream: async (
    content: string, 
    chapterId: number | undefined,
    onData: (text: string) => void,
    onError: (err: string) => void,
    onDone: () => void,
    promptId?: number,
    promptVersion?: string
  ) => {
    try {
      const token = localStorage.getItem('token');
      const headers: any = {
        'Content-Type': 'application/json',
      };
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const response = await fetch(`${BASE_URL}/trim/stream`, {
        method: 'POST',
        headers: headers,
        body: JSON.stringify({
          content, 
          chapter_id: chapterId,
          prompt_id: promptId,
          prompt_version: promptVersion
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      
      const reader = response.body?.getReader();
      if (!reader) throw new Error('Response body is null');

      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        
        const lines = buffer.split('\n');
        buffer = lines.pop() ?? ''; 

        let currentEvent = '';
        
        for (const line of lines) {
          if (line.startsWith('event:')) {
            currentEvent = line.substring(6).trim();
          } else if (line.startsWith('data:')) {
            const data = line.substring(5); 
            if (currentEvent === 'message' || currentEvent === '') {
               // 尝试去除 data: 后面的一个前置空格（如果存在）
               let text = data;
               if (text.startsWith(' ')) text = text.substring(1);
               onData(text);
            } else if (currentEvent === 'error') {
               onError(data.trim());
            }
          } else if (line.trim() === '') {
            currentEvent = '';
          }
        }
      }
      onDone();
    } catch (err: any) {
      onError(err.message || 'Network Error');
    }
  }
};