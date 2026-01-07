import type { LocalBook, LocalChapter, TrimmedContent } from './types';

export interface IBookRepository {
  // --- 初始化 ---
  init(): Promise<void>;

  // --- 书籍 CRUD ---
  getBooks(): Promise<LocalBook[]>;
  getBook(id: number | string): Promise<LocalBook | null>;
  // App端: filePath是本地路径; MP端: filePath是临时路径, fileObj是文件对象
  addBook(filePath: string, fileName: string, onProgress?: (p: number) => void): Promise<LocalBook>;
  deleteBook(id: number | string): Promise<void>;

  // --- 章节 ---
  getChapters(bookId: number | string): Promise<LocalChapter[]>;
  getChapterContent(bookId: number | string, chapterId: number | string): Promise<string>;

  // --- 精简缓存 ---
  // 获取某章节指定模式的精简内容
  getTrimmedContent(chapterMd5: string, promptId: number): Promise<TrimmedContent | null>;
  // 保存精简内容 (App存SQLite, MP存云端/本地Storage)
  saveTrimmedContent(chapterMd5: string, promptId: number, content: string): Promise<void>;
  
  // --- 进度 ---
  updateProgress(bookId: number | string, chapterId: number | string): Promise<void>;
}
