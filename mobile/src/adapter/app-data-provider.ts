import { IDataProvider, ContentResult } from '../core/data-provider';
import type { Book, Chapter } from '../core/types';
import { db } from '../utils/sqlite';
import * as ContentAPI from '../api/content';

/**
 * App端数据提供者
 */
export class AppDataProvider implements IDataProvider {
  
  async getChapterContent(book: Book, chapter: Chapter): Promise<ContentResult> {
    if (book.syncState === 1) {
      const rows = await db.select<any>(
        `SELECT c.raw_content 
         FROM contents c 
         JOIN chapters ch ON c.chapter_md5 = ch.md5 
         WHERE ch.id = ?`,
        [chapter.id]
      );
      
      if (rows.length > 0 && rows[0].raw_content) {
        return {
          content: rows[0].raw_content,
          cached: true,
          source: 'sqlite'
        };
      }
    }
    
    const results = await this.getBatchChapterContents(book, [chapter]);
    return results[0] || { content: '', cached: false, source: 'network' as const };
  }
  
  async getBatchChapterContents(book: Book, chapters: Chapter[]): Promise<ContentResult[]> {
    if (book.syncState === 1) {
      const validMd5s = chapters
        .map(ch => ch.md5)
        .filter((m): m is string => !!m);
      
      if (validMd5s.length === 0) {
        return chapters.map(() => ({ content: '', cached: false, source: 'network' as const }));
      }
      
      const placeholders = validMd5s.map(() => '?').join(',');
      const rows = await db.select<any>(
        `SELECT chapter_md5, raw_content 
         FROM contents 
         WHERE chapter_md5 IN (${placeholders})`,
        validMd5s
      );
      
      const md5ToContent = new Map(rows.map(r => [r.chapter_md5, r.raw_content]));
      
      const results = chapters.map(chapter => {
        const content = chapter.md5 ? md5ToContent.get(chapter.md5) : undefined;
        return {
          content: content || '',
          cached: !!content,
          source: (content ? 'sqlite' : 'network') as 'sqlite' | 'network'
        };
      });
      
      const missingChapters = chapters.filter((ch, idx) => !results[idx].content);
      if (missingChapters.length > 0) {
        const missingIds = missingChapters
          .map(ch => ch.cloud_id)
          .filter((id): id is number => id > 0);
        
        if (missingIds.length > 0) {
          try {
            const remoteResults = await ContentAPI.getBatchChapterContents(missingIds);
            
            for (const remote of remoteResults.data) {
              if (remote.chapter_md5) {
                await db.execute(
                  'INSERT OR REPLACE INTO contents (chapter_md5, raw_content) VALUES (?, ?)',
                  [remote.chapter_md5, remote.content]
                );
              }
            }
            
            const remoteMap: Map<string, string> = new Map(remoteResults.data.map(r => [r.chapter_md5 as string, r.content]));
            for (let i = 0; i < chapters.length; i++) {
              if (!results[i].content && chapters[i].md5) {
                const content = remoteMap.get(chapters[i].md5);
                results[i] = {
                  content: content || '',
                  cached: false,
                  source: 'network' as const
                };
              }
            }
          } catch (e) {
            console.error('[AppDataProvider] Failed to fetch remote content:', e);
          }
        }
      }
      
    return results;
    }
    
    const cloudIds = chapters
      .map(ch => ch.cloud_id)
      .filter((id): id is number => id > 0);
    
    if (cloudIds.length === 0) {
      return chapters.map(() => ({ content: '', cached: false, source: 'network' as const }));
    }
    
    try {
      const remoteResults = await ContentAPI.getBatchChapterContents(cloudIds);
      const remoteMap = new Map(remoteResults.data.map(r => [r.chapter_md5 as string, r.content]));
      
      return chapters.map(chapter => ({
        content: chapter.md5 ? (remoteMap.get(chapter.md5) || '') : '',
        cached: false,
        source: 'network' as const
      }));
    } catch (e) {
      console.error('[AppDataProvider] Failed to fetch remote content:', e);
      return chapters.map(() => ({ content: '', cached: false, source: 'network' as const }));
    }
  }
   
  async updateProgress(
    bookId: number, 
    chapterId: number, 
    promptId?: number
  ): Promise<void> {
    await db.execute(
      `INSERT OR REPLACE INTO reading_history 
       (book_id, last_chapter_id, last_prompt_id, updated_at) 
       VALUES (?, ?, ?, ?)`,
      [bookId, chapterId, promptId || 0, Date.now()]
    );
    
    ContentAPI.updateReadingProgress(bookId, chapterId, promptId).catch(err => {
      console.warn('[AppDataProvider] Failed to upload progress:', err);
    });
  }
}
