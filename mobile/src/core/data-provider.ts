import type { Book, Chapter } from './types';

/**
 * 数据获取结果（带缓存指示）
 */
export interface ContentResult {
  content: string;
  cached: boolean;  // true = 来自缓存，false = 来自网络
  source: 'memory' | 'storage' | 'sqlite' | 'network';
}

/**
 * 精简状态映射
 * - MD5寻址: Record<string, number[]>  => { "md5_xxx": [1, 2] }
 * - ID寻址: Record<number, number[]>     => { 5001: [1, 2] }
 */
export type TrimmedStatusMap = Record<string, number[]> | Record<number, number[]>;

/**
 * 数据提供者接口
 * 
 * 这是设计文档中的核心接口，用于抽象不同平台的数据获取逻辑。
 * 
 * 设计原则：
 * 1. 三级缓存：Memory -> Storage/SQLite -> Network
 * 2. 适配器模式：App端使用SQLite，小程序端使用API+Storage
 * 3. 批量预加载：支持批量获取提升性能
 * 4. 缓存指示：返回数据来源，便于调试和监控
 * 
 * @see DOCS_ARCHITECTURE.md 第6章：Uni-app 代码实现策略
 * @see DOCS_INTERACTION_SPEC.md 第7章：前端决策矩阵
 */
export interface IDataProvider {
  // --- 章节内容获取 ---
  
  /**
   * 获取单个章节内容（带三级缓存）
   * 
   * @param book 书籍信息
   * @param chapter 章节信息
   * @returns 章节内容，包含缓存指示
   */
  getChapterContent(book: Book, chapter: Chapter): Promise<ContentResult>;
  
  /**
   * 批量获取章节内容（预加载，最多10个）
   * 
   * @param book 书籍信息
   * @param chapters 章节列表
   * @returns 章节内容列表，包含缓存指示
   */
  getBatchChapterContents(book: Book, chapters: Chapter[]): Promise<ContentResult[]>;
    
  // --- 阅读进度 ---
  
  /**
   * 上报阅读进度
   * 
   * @param bookId 书籍ID
   * @param chapterId 章节ID
   * @param promptId 精简模式ID（可选）
   * 
   * @description
   * - App端: 保存到本地 SQLite + 异步上报云端
   * - 小程序: 直接上报云端
   */
  updateProgress(
    bookId: number, 
    chapterId: number, 
    promptId?: number
  ): Promise<void>;
}
