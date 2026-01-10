import { request, Response } from './index';

// 重新导出类型，保持兼容性
export interface ChapterContentResponse {
  chapter_id: number;
  chapter_md5: string;
  content: string;
}

export interface TrimmedContentResponse {
  chapter_id?: number;
  chapter_md5?: string;
  prompt_id: number;
  trimmed_content: string;
}

export interface TrimmedStatusResponse {
  trimmed_map: Record<string, number[]> | Record<number, number[]>;
}

/**
 * 章节内容响应
 */
export interface ChapterContentResponse {
  chapter_id: number;
  chapter_md5: string;
  content: string;
}

/**
 * 精简内容响应
 */
export interface TrimmedContentResponse {
  chapter_id?: number;
  chapter_md5?: string;
  prompt_id: number;
  trimmed_content: string;
}

/**
 * 精简状态响应
 * - MD5寻址: Record<string, number[]>  => { "md5_xxx": [1, 2] }
 * - ID寻址: Record<number, number[]>     => { 5001: [1, 2] }
 */
export interface TrimmedStatusResponse {
  trimmed_map: Record<string, number[]> | Record<number, number[]>;
}

/**
 * 批量获取章节原文
 * 对应后端: POST /chapters/content
 * 
 * @param ids 章节ID数组（上限10）
 */
export function getBatchChapterContents(ids: number[]): Promise<Response<ChapterContentResponse[]>> {
  return request({
    url: '/chapters/content',
    method: 'POST',
    data: { ids }
  });
}

/**
 * 批量获取精简内容（ID寻址）
 * 对应后端: POST /chapters/trim
 * 
 * @param ids 章节ID数组（上限10）
 * @param promptId 精简模式ID
 */
export function getBatchTrimmedById(ids: number[], promptId: number): Promise<Response<TrimmedContentResponse[]>> {
  return request({
    url: '/chapters/trim',
    method: 'POST',
    data: { ids, prompt_id: promptId }
  });
}

/**
 * 批量获取精简内容（MD5寻址）
 * 对应后端: POST /contents/trim
 * 
 * @param md5s 章节MD5数组（上限10）
 * @param promptId 精简模式ID
 */
export function getBatchTrimmedByMd5(md5s: string[], promptId: number): Promise<Response<TrimmedContentResponse[]>> {
  return request({
    url: '/contents/trim',
    method: 'POST',
    data: { md5s, prompt_id: promptId }
  });
}

/**
 * 按MD5探测精简足迹
 * 对应后端: POST /contents/sync-status
 * 
 * @param md5s 章节MD5数组
 */
export function syncTrimmedStatusByMd5(md5s: string[]): Promise<Response<TrimmedStatusResponse>> {
  return request({
    url: '/contents/sync-status',
    method: 'POST',
    data: { md5s }
  });
}

/**
 * 按章节ID刷新精简足迹
 * 对应后端: POST /chapters/sync-status
 * 
 * @param bookId 书籍ID
 */
export function syncTrimmedStatusById(bookId: number): Promise<Response<TrimmedStatusResponse>> {
  return request({
    url: '/chapters/sync-status',
    method: 'POST',
    data: { book_id: bookId }
  });
}

/**
 * 上报阅读进度
 * 对应后端: POST /books/:id/progress
 * 
 * @param bookId 书籍ID
 * @param chapterId 章节ID
 * @param promptId 精简模式ID（可选）
 */
export function updateReadingProgress(bookId: number, chapterId: number, promptId?: number): Promise<Response<void>> {
  return request({
    url: `/books/${bookId}/progress`,
    method: 'POST',
    data: { chapter_id: chapterId, prompt_id: promptId }
  });
}
