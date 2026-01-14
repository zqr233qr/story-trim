import { IDataProvider, ContentResult } from '../core/data-provider';
import type { Book, Chapter } from '../core/types';
import * as ContentAPI from '../api/content';

/**
 * Tier 2 缓存键命名
 */
const STORAGE_KEYS = {
  CONTENT: (md5: string) => `st_content_${md5}`,
  TRIM: (md5: string, promptId: number) => `st_trim_${md5}_${promptId}`
};

/**
 * 小程序端数据提供者
 * 
 * 实现两级缓存架构：
 * - Tier 1: 内存缓存 (Pinia Store)
 * - Tier 2: 持久化缓存 (uni.setStorage)
 * 
 * 对应后端接口：
 * - 批量章节内容: POST /chapters/content
 * - 批量精简内容: POST /chapters/trim (ID寻址）
 * - 同步精简状态: POST /chapters/sync-status (ID寻址）
 * - 上报阅读进度: POST /books/:id/progress
 * 
 * @see DOCS_INTERACTION_SPEC.md 第7章：前端决策矩阵
 */
export class MiniDataProvider implements IDataProvider {
  
  async getChapterContent(book: Book, chapter: Chapter): Promise<ContentResult> {
    if (!chapter.cloud_id) {
      return { content: '', cached: false, source: 'network' as const };
    }

    const cached = uni.getStorageSync(STORAGE_KEYS.CONTENT(chapter.chapter_md5 || chapter.md5 || ''));
    if (cached) {
      return {
        content: cached,
        cached: true,
        source: 'storage'
      };
    }
    
    const results = await this.getBatchChapterContents(book, [chapter]);
    return results[0];
  }
  
  async getBatchChapterContents(book: Book, chapters: Chapter[]): Promise<ContentResult[]> {
    const results: ContentResult[] = [];
    const missingChapters: Chapter[] = [];
    
    for (const chapter of chapters) {
      const cacheKey = chapter.chapter_md5 || chapter.md5 || '';
      const cached = uni.getStorageSync(STORAGE_KEYS.CONTENT(cacheKey));
      if (cached) {
        results.push({
          content: cached,
          cached: true,
          source: 'storage'
        });
      } else {
        results.push({ content: '', cached: false, source: 'network' as const });
        missingChapters.push(chapter);
      }
    }
    
    if (missingChapters.length > 0) {
      const ids = missingChapters.map(ch => ch.cloud_id).filter((id): id is number => id > 0);
      if (ids.length === 0) return results;
      
      try {
        const remoteResults = await ContentAPI.getBatchChapterContents(ids);
        
        for (const remote of remoteResults.data) {
          const cacheKey = remote.chapter_md5;
          if (cacheKey) {
            uni.setStorage({
              key: STORAGE_KEYS.CONTENT(cacheKey),
              data: remote.content
            });
          }
        }
        
        const remoteMap = new Map(remoteResults.data.map(r => [r.chapter_md5 as string, r.content]));
        for (const chapter of missingChapters) {
          const idx = chapters.indexOf(chapter);
          const cacheKey = chapter.chapter_md5 || chapter.md5 || '';
          if (idx >= 0 && cacheKey) {
            results[idx] = {
              content: remoteMap.get(cacheKey) || '',
              cached: false,
              source: 'network' as const
            };
          }
        }
      } catch (e) {
        console.error('[MiniDataProvider] Failed to fetch remote content:', e);
      }
    }
    
    return results;
  }
    
  async updateProgress(
    bookId: number, 
    chapterId: number, 
    promptId?: number
  ): Promise<void> {
    try {
      await ContentAPI.updateReadingProgress(bookId, chapterId, promptId);
    } catch (e) {
      console.error('[MiniDataProvider] Failed to upload progress:', e);
    }
  }
}
