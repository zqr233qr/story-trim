// 基础配置
// 重要：真机调试时，请将下方 IP 替换为你电脑的局域网 IP (可通过 ifconfig 查看)
const LOCAL_IP = '192.168.3.178'; 

let BASE_URL = '/api/v1';

// #ifndef H5
BASE_URL = `http://${LOCAL_IP}:8080/api/v1`;
// #endif

if (import.meta.env.PROD) {
  BASE_URL = '/api/v1';
}

// 统一响应结构
export interface Response<T> {
  code: number;
  msg: string;
  data: T;
}

// 核心请求方法封装 (适配 Uni-app)
const request = <T>(options: UniApp.RequestOptions): Promise<Response<T>> => {
  const token = uni.getStorageSync('token');
  const finalUrl = BASE_URL + options.url;
  console.log('[API Request]', options.method || 'GET', finalUrl);
  
  return new Promise((resolve, reject) => {
    uni.request({
      ...options,
      url: finalUrl,
      header: {
        ...options.header,
        'Authorization': token ? `Bearer ${token}` : '',
      },
      success: (res) => {
        console.log('[API Response]', res.statusCode, res.data);
        if (res.statusCode === 401) {
          uni.removeStorageSync('token');
          uni.reLaunch({ url: '/pages/login/login' });
          reject(new Error('Unauthorized'));
        } else {
          resolve(res.data as Response<T>);
        }
      },
      fail: (err) => {
        console.error('[API Fail]', err);
        reject(err);
      },
    });
  });
};

// --- 类型定义 (直接从 web 项目同步) ---

export interface User { id: number; username: string; token?: string; }
export interface Book { id: number; title: string; total_chapters: number; fingerprint: string; created_at: string; }
export interface Chapter { 
  id: number; book_id: number; index: number; title: string; 
  content?: string; trimmed_content?: string; trimmed_prompt_ids?: number[]; 
}
export interface ReadingHistory { last_chapter_id: number; last_prompt_id: number; }
export interface BookDetail { book: Book; chapters: Chapter[]; trimmed_ids: number[]; reading_history?: ReadingHistory; }
export interface Prompt { id: number; name: string; description?: string; is_default?: boolean; version: string; content: string; is_system: boolean; }
export interface Task { id: string; type: string; status: string; progress: number; error?: string; }

// --- API 方法 ---

export const api = {
  login: (data: any) => request<{ token: string }>({ url: '/auth/login', method: 'POST', data }),
  register: (data: any) => request<void>({ url: '/auth/register', method: 'POST', data }),
  
  getBooks: () => request<Book[]>({ url: '/books', method: 'GET' }),
  getBookDetail: (id: number, promptId?: number) => 
    request<BookDetail>({ url: `/books/${id}`, method: 'GET', data: { prompt_id: promptId } }),
  getChapter: (id: number) => 
    request<Chapter>({ url: `/chapters/${id}`, method: 'GET' }),
  getChapterTrim: (id: number, promptId: number) => 
    request<{ prompt_id: number, trimmed_content: string }>({ url: `/chapters/${id}/trim`, method: 'GET', data: { prompt_id: promptId } }),
  getPrompts: () => request<Prompt[]>({ url: '/prompts', method: 'GET' }),

  updateProgress: (bookId: number, chapterId: number, promptId: number) => 
    request<void>({ url: `/books/${bookId}/progress`, method: 'POST', data: { chapter_id: chapterId, prompt_id: promptId } }),

  getBatchChapters: (chapterIds: number[], promptId?: number) => 
    request<{ id: number, content: string, trimmed_content: string }[]>({ url: '/chapters/batch', method: 'POST', data: { chapter_ids: chapterIds, prompt_id: promptId } }),

  startBatchTrim: (bookId: number, promptId: number) => 
    request<{ task_id: string }>({ url: '/tasks/batch-trim', method: 'POST', data: { book_id: bookId, prompt_id: promptId } }),

  getTaskStatus: (taskId: string) => 
    request<Task>({ url: `/tasks/${taskId}`, method: 'GET' }),

  upload: (filePath: string, fileName: string) => {
    return new Promise<Response<Book>>((resolve, reject) => {
      // #ifdef APP-PLUS
      if (typeof plus === 'undefined') {
        reject(new Error('plus is not defined'));
        return;
      }
      // #endif
      uni.uploadFile({
        url: BASE_URL + '/upload',
        filePath: filePath,
        name: 'file',
        header: { 'Authorization': `Bearer ${uni.getStorageSync('token')}` },
        success: (res) => resolve(JSON.parse(res.data)),
        fail: (err) => reject(err)
      });
    });
  },

  // SSE/WS 流式解析适配器
  trimStream: async (
    chapterId: number,
    promptId: number,
    onData: (text: string) => void,
    onError: (err: string) => void,
    onDone: () => void
  ) => {
    const token = uni.getStorageSync('token');
    
    // #ifdef H5
    const url = `${BASE_URL}/trim/stream`;
    const body = JSON.stringify({ chapter_id: chapterId, prompt_id: promptId });
    try {
      const response = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
        body
      });
      const reader = response.body?.getReader();
      if (!reader) throw new Error('No body');
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
            const raw = line.substring(5).trim();
            try { const p = JSON.parse(raw); if (p.c) onData(p.c); } catch (e) {}
          }
        }
      }
      onDone();
    } catch (e: any) { onError(e.message); }
    // #endif

    // #ifndef H5
    // App 和 小程序使用 WebSocket
    // 注意：ws 协议需要根据 http/https 自动切换
    const wsBase = BASE_URL.replace('http', 'ws');
    const wsUrl = `${wsBase}/trim/ws?token=${token}&chapter_id=${chapterId}&prompt_id=${promptId}`;
    
    console.log('[WS Connect]', wsUrl);
    
    const socketTask = uni.connectSocket({
      url: wsUrl,
      complete: () => {}
    });

    socketTask.onMessage((res) => {
      try {
        const p = JSON.parse(res.data as string);
        if (p.error) {
          onError(p.error);
          socketTask.close({});
        } else if (p.c) {
          onData(p.c);
        }
      } catch (e) {}
    });

    socketTask.onClose(() => {
      onDone();
    });

    socketTask.onError((err) => {
      onError('WebSocket Error');
    });
    // #endif
  }
};
