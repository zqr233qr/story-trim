// 基础配置
// 重要：真机调试时，请将下方 IP 替换为你电脑的局域网 IP
const LOCAL_IP = '192.168.3.178'; 

export let BASE_URL = '/api/v1';

// #ifndef H5
BASE_URL = `http://${LOCAL_IP}:8080/api/v1`;
// #endif

const WS_BASE_URL = BASE_URL.replace('http://', 'ws://').replace('https://', 'wss://');

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
export const request = <T>(options: UniApp.RequestOptions): Promise<Response<T>> => {
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
        // console.log('[API Response]', res.statusCode);
        if (res.statusCode === 401) {
          uni.removeStorageSync('token');
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

// --- 类型定义 ---
export interface User { id: number; username: string; token?: string; }
export interface Book { id: number; title: string; total_chapters: number; fingerprint: string; created_at: string; book_md5?: string; }
export interface Chapter { 
  id: number; book_id: number; index: number; title: string; 
  content?: string; trimmed_content?: string; trimmed_prompt_ids?: number[]; md5?: string;
}
export interface ReadingHistory { last_chapter_id: number; last_prompt_id: number; }
export interface BookDetail { book: Book; chapters: Chapter[]; trimmed_ids?: number[]; trimmed_map?: Record<number, number[]>; reading_history?: ReadingHistory; }
export interface Prompt { id: number; name: string; description?: string; is_default?: boolean; version: string; content: string; is_system: boolean; }
export interface Task { id: string; type: string; status: string; progress: number; error?: string; }

// --- API 方法 ---
export const api = {
  login: (data: any) => request<{ token: string }>({ url: '/auth/login', method: 'POST', data }),
  register: (data: any) => request<void>({ url: '/auth/register', method: 'POST', data }),
  
  getBooks: () => request<Book[]>({ url: '/books', method: 'GET' }),
  getBookDetail: (id: number, promptId?: number) => 
    request<BookDetail>({ url: `/books/${id}`, method: 'GET', data: { prompt_id: promptId } }),
  getPrompts: () => request<Prompt[]>({ url: '/common/prompts', method: 'GET' }),

  syncTrimmedStatus: (md5s: string[]) => request<{ trimmed_map: Record<string, number[]> }>({ 
    url: '/contents/sync-status', 
    method: 'POST', 
    data: { md5s } 
  }),

  syncLocalBook: (data: { book_name: string; book_md5: string; cloud_book_id?: number; chapters: any[] }) => request<{ book_id: number; chapter_mappings: Array<{ local_id: number; cloud_id: number }> }>({
    url: '/books/sync-local',
    method: 'POST',
    data
  }),

  // 1. 基于 ChapterID 的流式 (SSE/WS)
  trimStream: async (
    chapterId: number,
    promptId: number,
    onData: (text: string) => void,
    onError: (err: string) => void,
    onDone: () => void
  ) => {
    const token = uni.getStorageSync('token');
    
    // #ifdef H5
    // H5 直接用 Fetch SSE
    try {
      const response = await fetch(`${BASE_URL}/trim/stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
        body: JSON.stringify({ chapter_id: chapterId, prompt_id: promptId })
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
    // App 端使用 WebSocket
    const wsBase = BASE_URL.replace('http', 'ws');
    const wsUrl = `${wsBase}/trim/stream/by-id?token=${token}&chapter_id=${chapterId}&prompt_id=${promptId}`;
    const socketTask = uni.connectSocket({ url: wsUrl, complete: () => {} });
    socketTask.onMessage((res) => {
      try {
        const p = JSON.parse(res.data as string);
        if (p.error) { onError(p.error); socketTask.close({}); } 
        else if (p.c) { onData(p.c); }
      } catch (e) {}
    });
    socketTask.onClose(() => { onDone(); });
    socketTask.onError(() => { onError('WebSocket Error'); });
    // #endif
  },

  // 2. 基于 RawContent 的流式 (无状态, 支持离线混合模式)
  trimStreamRaw(content: string, promptId: number, md5: string | undefined, bookFingerprint: string, chapterIndex: number, onData: (chunk: string) => void, onError: (err: string) => void, onDone: () => void) {
    const rawToken = uni.getStorageSync('token')
    const token = rawToken || ''

    // 使用 WebSocket 连接到 /trim/stream/by-md5
    const socketTask = uni.connectSocket({
      url: `${WS_BASE_URL}/trim/stream/by-md5?token=${token}`,
      complete: ()=> {}
    });

    socketTask.onOpen(() => {
      console.log('[WS Open] Sending payload...')
      socketTask.send({
        data: JSON.stringify({
          content,
          prompt_id: promptId,
          md5: md5 || '',
          book_fingerprint: bookFingerprint,
          chapter_index: chapterIndex
        })
      })
    })

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
      console.log('[WS Close]');
      onDone();
    });

    socketTask.onError((err) => {
      console.error('[WS Error]', err);
      onError('WebSocket Error');
    });
  }
};