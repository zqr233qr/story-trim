export type PlatformType = 'app' | 'mp' | 'h5';

export interface LocalBook {
  id: number | string;
  title: string;
  bookMD5?: string;
  cloudId?: number;
  syncedCount?: number;
  totalChapters: number;
  lastReadChapterId?: number | string;
  processStatus: 'new' | 'processing' | 'ready';
  platform: PlatformType;
  cover?: string;
  createdAt: number;
  syncState?: number;
}

export interface LocalChapter {
  id: number | string;
  bookId: number | string;
  index: number;
  title: string;
  content?: string;         // 列表查询时为空，详情查询时有值
  words_count: number;      // 字数（与服务端保持一致）
  length?: number;          // 原始长度（RenderJS传递）
  md5: string;              // 章节归一化MD5
  trimmedPromptIds?: number[];
  cloudId?: number;         // 云端章节ID（用于同步）
}

export interface TrimmedContent {
  id?: number;
  sourceMd5: string;        // 关联原文MD5
  promptId: number;
  content: string;
  createdAt: number;
}

// 扩展类型：用于云端API响应
export interface CloudBook {
  id: number;
  title: string;
  total_chapters: number;
  book_md5?: string;
  created_at: string;
}

export interface CloudChapter {
  id: number;
  book_id: number;
  index: number;
  title: string;
  chapter_md5?: string;
  md5?: string;
  cloud_id?: number;      // 云端章节ID
}

// 统一的 Book 和 Chapter 接口（兼容 Local 和 Cloud）
export interface Book extends CloudBook {
  syncState?: number;       // 0: Local, 1: Synced, 2: CloudOnly
  chapters?: Chapter[];
  activeChapterIndex?: number;
  activeModeId?: string;
}

export interface Chapter extends CloudChapter {
  content?: string;
  trimmed_content?: string;
  trimmed_prompt_ids?: number[];
  isLoaded?: boolean;
  modes?: Record<string, string[]>;
}
