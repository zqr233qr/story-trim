export type PlatformType = 'app' | 'mp' | 'h5';

export interface LocalBook {
  id: number | string;      // 统一ID (App端是number, MP端是string)
  title: string;
  fingerprint: string;      // 书籍指纹 (第一章归一化MD5)
  totalChapters: number;
  lastReadChapterId?: number | string;
  processStatus: 'new' | 'processing' | 'ready';
  platform: PlatformType;
  cover?: string;           // 封面图片(可选)
  createdAt: number;
  syncState?: number;       // 0: Local, 1: Synced, 2: CloudOnly
}

export interface LocalChapter {
  id: number | string;
  bookId: number | string;
  index: number;
  title: string;
  content?: string;         // 列表查询时为空，详情查询时有值
  wordCount: number;
  md5: string;              // 章节归一化MD5
  trimmedPromptIds?: number[];
}

export interface TrimmedContent {
  id?: number;
  sourceMd5: string;        // 关联原文MD5
  promptId: number;
  content: string;
  createdAt: number;
}
