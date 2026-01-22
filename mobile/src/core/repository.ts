import type { LocalBook, LocalChapter } from "./types";

export interface IBookRepository {
  // --- 初始化 ---
  init(): Promise<void>;

  // --- 书籍 CRUD ---
  getBooks(): Promise<LocalBook[]>;
  getBook(id: number | string): Promise<LocalBook | null>;

  // --- 云端同步 ---
  syncBookFromCloud(cloudBook: CloudBook, userId: number): Promise<void>;
  // 全量下载书籍内容并落库
  downloadBookContent(
    bookId: number,
    cloudBookId: number,
    userId: number,
    onProgress?: (progress: number) => void,
  ): Promise<void>;
  // 全量上传书籍内容并同步
  uploadBookZip(
    bookId: number,
    onProgress?: (progress: number) => void,
  ): Promise<{
    book_id: number;
    chapter_mappings: Array<{ local_id: number; cloud_id: number }>;
  }>;

  // --- 章节 ---
  getChapters(bookId: number | string): Promise<LocalChapter[]>;
  getChapterContent(
    bookId: number | string,
    chapterId: number | string,
  ): Promise<string>;

  // --- 进度 ---
  updateProgress(
    bookId: number | string,
    chapterId: number | string,
  ): Promise<void>;
}

export interface CloudBook {
  id: number;
  book_md5?: string;
  title: string;
  total_chapters: number;
  created_at: string;
}
