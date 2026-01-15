import { request } from './index';
import type { Response } from './index';

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
 * @param bookId 书籍ID (云端ID)
 * @param chapterId 章节ID
 * @param promptId 精简模式ID（可选，0表示原文）
 */
export function updateReadingProgress(bookId: number, chapterId: number, promptId?: number): Promise<Response<void>> {
  return request({
    url: `/books/${bookId}/progress`,
    method: 'POST',
    data: { chapter_id: chapterId, prompt_id: promptId || 0 }
  });
}

/**
 * 根据章节ID查询已精简状态
 * 对应后端: POST /chapters/status
 */
export function getChapterTrimStatusById(chapterId: number): Promise<Response<{ prompt_ids: number[] }>> {
  return request({
    url: '/chapters/status',
    method: 'POST',
    data: { chapter_id: chapterId }
  });
}

/**
 * 根据MD5查询已精简状态
 * 对应后端: POST /contents/status
 */
export function getChapterTrimStatusByMd5(md5: string): Promise<Response<{ prompt_ids: number[] }>> {
  return request({
    url: '/contents/status',
    method: 'POST',
    data: { chapter_md5: md5 }
  });
}
