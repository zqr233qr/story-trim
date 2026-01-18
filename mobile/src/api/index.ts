// 基础配置
// 统一使用云端服务器地址
export const BASE_URL = "http://110.41.38.214:8088/api/v1";

const WS_BASE_URL = BASE_URL.replace("http://", "ws://").replace(
  "https://",
  "wss://",
);

// 统一响应结构
export interface Response<T> {
  code: number;
  msg: string;
  data: T;
}

// 核心请求方法封装 (适配 Uni-app)
export const request = <T>(
  options: UniApp.RequestOptions,
): Promise<Response<T>> => {
  const token = uni.getStorageSync("token");
  const finalUrl = BASE_URL + options.url;
  console.log("[API Request]", options.method || "GET", finalUrl);

  return new Promise((resolve, reject) => {
    uni.request({
      ...options,
      url: finalUrl,
      header: {
        ...options.header,
        Authorization: token ? `Bearer ${token}` : "",
      },
      success: (res) => {
        // console.log('[API Response]', res.statusCode);
        if (res.statusCode === 401) {
          uni.removeStorageSync("token");
          reject(new Error("Unauthorized"));
        } else {
          resolve(res.data as Response<T>);
        }
      },
      fail: (err) => {
        console.error("[API Fail]", err);
        reject(err);
      },
    });
  });
};

// --- 类型定义 ---
export interface User {
  id: number;
  username: string;
  token?: string;
}
export interface Book {
  id: number;
  title: string;
  total_chapters: number;
  created_at: string;
  book_md5?: string;
  full_trim_status?: string;
  full_trim_progress?: number;
}
export interface Chapter {
  id: number;
  book_id: number;
  index: number;
  title: string;
  content?: string;
  trimmed_content?: string;
  trimmed_prompt_ids?: number[];
  md5?: string;
}
export interface ReadingHistory {
  last_chapter_id: number;
  last_prompt_id: number;
}
export interface BookDetail {
  book: Book;
  chapters: Chapter[];
}
export interface Prompt {
  id: number;
  name: string;
  description?: string;
  is_default?: boolean;
  version: string;
  content: string;
  is_system: boolean;
}
export interface Task {
  id: string;
  type: string;
  status: string;
  progress: number;
  error?: string;
}
export interface ParserRule {
  name: string;
  pattern: string;
  weight: number;
}
export interface ParserConfig {
  version: number;
  rules: ParserRule[];
}

// --- API 方法 ---
export const api = {
  login: (data: any) =>
    request<{ token: string }>({ url: "/auth/login", method: "POST", data }),
  register: (data: any) =>
    request<void>({ url: "/auth/register", method: "POST", data }),

  getBooks: () => request<Book[]>({ url: "/books", method: "GET" }),
  getParserRules: () =>
    request<ParserConfig>({ url: "/common/parser-rules", method: "GET" }),
  getBookDetail: (id: number) =>
    request<BookDetail>({
      url: `/books/${id}`,
      method: "GET",
    }),
  getBookProgress: (id: number) =>
    request<ReadingHistory>({
      url: `/books/${id}/progress`,
      method: "GET",
    }),
  getPrompts: () =>
    request<Prompt[]>({ url: "/common/prompts", method: "GET" }),

  syncTrimmedStatus: (md5s: string[]) =>
    request<{ trimmed_map: Record<string, number[]> }>({
      url: "/contents/sync-status",
      method: "POST",
      data: { md5s },
    }),

  syncLocalBook: (data: {
    book_name: string;
    book_md5: string;
    cloud_book_id?: number;
    chapters: any[];
  }) =>
    request<{
      book_id: number;
      chapter_mappings: Array<{ local_id: number; cloud_id: number }>;
    }>({
      url: "/books/sync-local",
      method: "POST",
      data,
    }),

  // 1. 基于 ChapterID 的流式 (SSE/WS)
  trimStream: async (
    chapterId: number,
    promptId: number,
    onData: (text: string) => void,
    onError: (err: string) => void,
    onDone: () => void,
  ) => {
    const token = uni.getStorageSync("token");

    // App 端使用 WebSocket
    const wsBase = BASE_URL.replace("http", "ws");
    const wsUrl = `${wsBase}/trim/stream/by-id?token=${token}&chapter_id=${chapterId}&prompt_id=${promptId}`;
    const socketTask = uni.connectSocket({ url: wsUrl, complete: () => {} });
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
    socketTask.onError(() => {
      onError("WebSocket Error");
    });
  },

  // 2. 基于 RawContent 的流式 (无状态, 支持离线混合模式)
  trimStreamRaw(
    content: string,
    promptId: number,
    md5: string | undefined,
    bookMD5: string,
    chapterIndex: number,
    onData: (chunk: string) => void,
    onError: (err: string) => void,
    onDone: () => void,
  ) {
    const rawToken = uni.getStorageSync("token");
    const token = rawToken || "";

    // 使用 WebSocket 连接到 /trim/stream/by-md5
    const socketTask = uni.connectSocket({
      url: `${WS_BASE_URL}/trim/stream/by-md5?token=${token}`,
      complete: () => {},
    });

    socketTask.onOpen(() => {
      console.log("[WS Open] Sending payload...");
      socketTask.send({
        data: JSON.stringify({
          content,
          prompt_id: promptId,
          md5: md5 || "",
          book_md5: bookMD5,
          chapter_index: chapterIndex,
        }),
      });
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
      console.log("[WS Close]");
      onDone();
    });

    socketTask.onError((err) => {
      console.error("[WS Error]", err);
      onError("WebSocket Error");
    });
  },

  // 删除书籍
  deleteBook: async (id: number) => {
    return request({
      url: `/books/${id}`,
      method: "DELETE",
    });
  },
};
