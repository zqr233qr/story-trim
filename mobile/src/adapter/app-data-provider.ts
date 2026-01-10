import { IDataProvider, ContentResult, TrimmedStatusMap } from '../core/data-provider';
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
  
  async getTrimmedStatus(book: Book, chapters: Chapter[]): Promise<TrimmedStatusMap> {
    if (book.syncState === 1) {
      const validMd5s = chapters
        .map(ch => ch.md5)
        .filter((m): m is string => !!m);
      
      if (validMd5s.length === 0) return {};
      
      try {
        const response = await ContentAPI.syncTrimmedStatusByMd5(validMd5s);
        return response.data.trimmed_map;
      } catch (e) {
        console.error('[AppDataProvider] Failed to sync trimmed status:', e);
        return {};
      }
    }
    
    const cloudIds = chapters
      .map(ch => ch.cloud_id)
      .filter((id): id is number => id > 0);
    
    if (cloudIds.length === 0) return {};
    
    try {
      const response = await ContentAPI.syncTrimmedStatusById(book.id);
      return response.data.trimmed_map;
    } catch (e) {
      console.error('[AppDataProvider] Failed to sync trimmed status:', e);
      return {};
    }
  }
  
  async getTrimmedContent(
    book: Book, 
    chapter: Chapter, 
    promptId: number
  ): Promise<ContentResult | null> {
    if (!chapter.md5) return null;
    
    const rows = await db.select<any>(
      'SELECT content FROM trimmed_content WHERE source_md5 = ? AND prompt_id = ?',
      [chapter.md5, promptId]
    );
    
    if (rows.length > 0 && rows[0].content) {
      return {
        content: rows[0].content,
        cached: true,
        source: 'sqlite'
      };
    }
    
    let remoteResults;
    if (book.syncState === 1) {
      remoteResults = await ContentAPI.getBatchTrimmedByMd5([chapter.md5], promptId);
    } else {
      if (!chapter.cloud_id || chapter.cloud_id === 0) return null;
      remoteResults = await ContentAPI.getBatchTrimmedById([chapter.cloud_id], promptId);
    }
    
    if (remoteResults.data.length > 0 && remoteResults.data[0].trimmed_content) {
      const content = remoteResults.data[0].trimmed_content;
      
      await this.saveTrimmedContent(chapter.md5, promptId, content);
      
      return {
        content,
        cached: false,
        source: 'network'
      };
    }
    
    return null;
  }
  
  async getBatchTrimmedContents(
    book: Book, 
    chapters: Chapter[], 
    promptId: number
  ): Promise<ContentResult[]> {
    const validMd5s = chapters
      .map(ch => ch.md5)
      .filter((m): m is string => !!m);
    
    if (validMd5s.length === 0) {
      return chapters.map(() => ({ content: '', cached: false, source: 'network' as const }));
    }
    
    const placeholders = validMd5s.map(() => '?').join(',');
    const rows = await db.select<any>(
      `SELECT source_md5, content FROM trimmed_content 
       WHERE source_md5 IN (${placeholders}) AND prompt_id = ?`,
      [...validMd5s, promptId]
    );
    
    const md5ToContent = new Map(rows.map(r => [r.source_md5, r.content]));
    
      const results = chapters.map(chapter => ({
        content: chapter.md5 ? md5ToContent.get(chapter.md5) : '',
        cached: chapter.md5 ? !!md5ToContent.get(chapter.md5) : false,
        source: (chapter.md5 && md5ToContent.get(chapter.md5) ? 'sqlite' : 'network') as 'sqlite' | 'network'
      }));
    
    const missingChapters = chapters.filter((ch, idx) => !results[idx].content);
    if (missingChapters.length > 0) {
      let remoteResults;
      if (book.syncState === 1) {
        const missingMd5s = missingChapters
          .map(ch => ch.md5)
          .filter((m): m is string => !!m);
        remoteResults = await ContentAPI.getBatchTrimmedByMd5(missingMd5s, promptId);
      } else {
        const missingIds = missingChapters
          .map(ch => ch.cloud_id)
          .filter((id): id is number => id > 0);
        
        if (missingIds.length === 0) return results;
        remoteResults = await ContentAPI.getBatchTrimmedById(missingIds, promptId);
      }
      
       for (const remote of remoteResults.data) {
         if (remote.chapter_md5) {
           await this.saveTrimmedContent(remote.chapter_md5, promptId, remote.trimmed_content);
         }
       }
       
       const remoteMap: Map<string, string> = new Map(
         remoteResults.data.map(r => [r.chapter_md5 || '', r.trimmed_content])
       );
       
       for (let i = 0; i < chapters.length; i++) {
         if (!results[i].content && chapters[i].md5) {
           const content = remoteMap.get(chapters[i].md5) || '';
           results[i] = {
             content,
             cached: false,
             source: 'network' as const
           };
         }
       }
    }
    
    return results;
  }
  
  async saveTrimmedContent(
    chapterMd5: string, 
    promptId: number, 
    content: string
  ): Promise<void> {
    await db.execute(
      'INSERT OR REPLACE INTO trimmed_content (source_md5, prompt_id, content, created_at) VALUES (?, ?, ?, ?)',
      [chapterMd5, promptId, content, Date.now()]
    );
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
