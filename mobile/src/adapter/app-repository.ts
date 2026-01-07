import { db } from '../utils/sqlite';
import { parser } from '../utils/parser';
import type { IBookRepository } from '../core/repository';
import type { LocalBook, LocalChapter, TrimmedContent } from '../core/types';

export class AppRepository implements IBookRepository {
  async init(): Promise<void> {
    await db.open();
    await db.execute(`
      CREATE TABLE IF NOT EXISTS books (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT,
        fingerprint TEXT,
        total_chapters INTEGER,
        process_status TEXT,
        created_at INTEGER
      )
    `);
    await db.execute(`
      CREATE TABLE IF NOT EXISTS chapters (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        book_id INTEGER,
        chapter_index INTEGER,
        title TEXT,
        content TEXT,
        md5 TEXT
      )
    `);
    await db.execute(`
      CREATE TABLE IF NOT EXISTS trimmed_content (
        source_md5 TEXT,
        prompt_id INTEGER,
        content TEXT,
        created_at INTEGER,
        PRIMARY KEY (source_md5, prompt_id)
      )
    `);
  }

  async getBooks(): Promise<LocalBook[]> {
    const rows = await db.select<any>('SELECT * FROM books ORDER BY created_at DESC');
    console.log('[Repo] getBooks raw rows:', JSON.stringify(rows))
    return rows.map(r => ({
      id: r.id,
      title: r.title,
      fingerprint: r.fingerprint,
      totalChapters: r.total_chapters,
      processStatus: r.process_status,
      platform: 'app',
      createdAt: r.created_at
    }));
  }

  async getBook(id: number | string): Promise<LocalBook | null> {
    const rows = await db.select<any>('SELECT * FROM books WHERE id = ?', [id]);
    if (rows.length === 0) return null;
    const r = rows[0];
    return {
      id: r.id,
      title: r.title,
      fingerprint: r.fingerprint,
      totalChapters: r.total_chapters,
      processStatus: r.process_status,
      platform: 'app',
      createdAt: r.created_at
    };
  }

  async addBook(filePath: string, fileName: string, onProgress?: (p: number) => void): Promise<LocalBook> {
    // 1. 解析文件 (0-80%)
    const result = await parser.parseFile(filePath, fileName, onProgress);
    
    // 2. 存入数据库 (事务)
    let bookId = 0;
    const startInsert = Date.now();
    
    await db.transaction(async () => {
      // 极速模式：关闭磁盘同步等待
      await db.execute('PRAGMA synchronous = OFF');
      await db.execute('PRAGMA journal_mode = MEMORY');

      // Insert Book
      await db.execute(
        'INSERT INTO books (title, fingerprint, total_chapters, process_status, created_at) VALUES (?, ?, ?, ?, ?)',
        [result.title, result.fingerprint, result.totalChapters, 'ready', Date.now()]
      );
      
      const res = await db.select<any>('SELECT last_insert_rowid() as id');
      bookId = res[0].id;

      // Insert Chapters (批量插入优化)
      const total = result.chapters.length;
      const BATCH_SIZE = 200; // 激进一点，但不要超过 SQL 变量限制 (999)
      
      for (let i = 0; i < total; i += BATCH_SIZE) {
        const chunkStart = Date.now();
        const chunk = result.chapters.slice(i, i + BATCH_SIZE);
        const placeholders = chunk.map(() => '(?, ?, ?, ?, ?)').join(',');
        const values = chunk.flatMap(c => [bookId, c.index, c.title, c.content, c.md5]);
        
        await db.execute(
          `INSERT INTO chapters (book_id, chapter_index, title, content, md5) VALUES ${placeholders}`,
          values
        );
        
        console.log(`[SQL] Batch insert ${chunk.length} chapters took ${Date.now() - chunkStart}ms`);
        
        // 更新入库进度
        if (onProgress) {
           const p = 80 + Math.floor(((i + chunk.length) / total) * 20);
           onProgress(p);
        }
      }
    });
    
    console.log(`[SQL] Total insert took ${Date.now() - startInsert}ms`);
    if (onProgress) onProgress(100);

    return {
      id: bookId,
      title: result.title,
      fingerprint: result.fingerprint,
      totalChapters: result.totalChapters,
      processStatus: 'ready',
      platform: 'app',
      createdAt: Date.now()
    };
  }

  async deleteBook(id: number | string): Promise<void> {
    await db.transaction(async () => {
      await db.execute('DELETE FROM chapters WHERE book_id = ?', [id]);
      await db.execute('DELETE FROM books WHERE id = ?', [id]);
    });
  }

  async getChapters(bookId: number | string): Promise<LocalChapter[]> {
    const rows = await db.select<any>('SELECT id, chapter_index, title, md5 FROM chapters WHERE book_id = ? ORDER BY chapter_index ASC', [bookId]);
    return rows.map(r => ({
      id: r.id,
      bookId: bookId,
      index: r.chapter_index,
      title: r.title,
      md5: r.md5,
      wordCount: 0 // TODO
    }));
  }

  async getChapterContent(bookId: number | string, chapterId: number | string): Promise<string> {
    const rows = await db.select<any>('SELECT content FROM chapters WHERE id = ?', [chapterId]);
    return rows.length > 0 ? rows[0].content : '';
  }

  async getTrimmedContent(chapterMd5: string, promptId: number): Promise<TrimmedContent | null> {
    const rows = await db.select<any>('SELECT * FROM trimmed_content WHERE source_md5 = ? AND prompt_id = ?', [chapterMd5, promptId]);
    if (rows.length === 0) return null;
    return {
      sourceMd5: rows[0].source_md5,
      promptId: rows[0].prompt_id,
      content: rows[0].content,
      createdAt: rows[0].created_at
    };
  }

  async saveTrimmedContent(chapterMd5: string, promptId: number, content: string): Promise<void> {
    await db.execute(
      'INSERT OR REPLACE INTO trimmed_content (source_md5, prompt_id, content, created_at) VALUES (?, ?, ?, ?)',
      [chapterMd5, promptId, content, Date.now()]
    );
  }

  // RenderJS 专用：分步插入
  async createBook(title: string, fingerprint: string, total: number): Promise<number> {
    await db.execute(
      'INSERT INTO books (title, fingerprint, total_chapters, process_status, created_at) VALUES (?, ?, ?, ?, ?)',
      [title, fingerprint, total, 'ready', Date.now()]
    );
    const res = await db.select<any>('SELECT last_insert_rowid() as id');
    return res[0].id;
  }

  async insertChapters(bookId: number, chapters: any[]): Promise<void> {
    // 移除 PRAGMA 优化，优先保证数据一致性
    
    const placeholders = chapters.map(() => '(?, ?, ?, ?, ?)').join(',');
    const values = chapters.flatMap(c => [bookId, c.index, c.title, c.content, c.md5]);
    
    await db.execute(
      `INSERT INTO chapters (book_id, chapter_index, title, content, md5) VALUES ${placeholders}`,
      values
    );
  }

  async updateProgress(bookId: number | string, chapterId: number | string): Promise<void> {
    // 暂未实现
  }
}
