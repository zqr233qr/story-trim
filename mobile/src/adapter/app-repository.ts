import { db } from '../utils/sqlite';
import { parser } from '../utils/parser';
import type { IBookRepository } from '../core/repository';
import type { LocalBook, LocalChapter, TrimmedContent } from '../core/types';
import type { CloudBook } from '../core/repository';

export class AppRepository implements IBookRepository {
  async init(): Promise<void> {
    await db.open();
    
    // 1. Books Table
    // sync_state:0=Local,1=Synced,2=CloudOnly
    await db.execute(`
      CREATE TABLE IF NOT EXISTS books (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        cloud_id INTEGER DEFAULT 0,
        book_md5 TEXT,
        fingerprint TEXT,
        title TEXT,
        total_chapters INTEGER,
        process_status TEXT,
        sync_state INTEGER DEFAULT 0,
        synced_count INTEGER DEFAULT 0,
        created_at INTEGER
      )
    `);
 
    // 2. Chapters Table
    await db.execute(`
      CREATE TABLE IF NOT EXISTS chapters (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        book_id INTEGER,
        cloud_id INTEGER DEFAULT 0,
        chapter_index INTEGER,
        title TEXT,
        content TEXT,
        md5 TEXT,
        words_count INTEGER DEFAULT 0
      )
    `);
 
    // 3. Trimmed Content Table (Local Cache / Tier 3)
    await db.execute(`
      CREATE TABLE IF NOT EXISTS trimmed_content (
        source_md5 TEXT,
        prompt_id INTEGER,
        content TEXT,
        created_at INTEGER,
        PRIMARY KEY (source_md5, prompt_id)
      )
    `);
 
    // 4. Contents Table (内容去重表 - 设计文档要求）
    await db.execute(`
      CREATE TABLE IF NOT EXISTS contents (
        chapter_md5 TEXT PRIMARY KEY,
        raw_content TEXT
      )
    `);
 
    // 5. Reading History Table (阅读进度表 - 设计文档要求）
    await db.execute(`
      CREATE TABLE IF NOT EXISTS reading_history (
        book_id INTEGER PRIMARY KEY,
        last_chapter_id INTEGER,
        last_prompt_id INTEGER,
        scroll_offset REAL,
        updated_at INTEGER
      )
    `);
  }

  // --- Sync Helpers ---
  
  async updateBookCloudInfo(localId: number, cloudId: number, syncState: number, syncedCount?: number): Promise<void> {
    if (syncedCount !== undefined) {
      await db.execute('UPDATE books SET cloud_id = ?, sync_state = ?, synced_count = ? WHERE id = ?', [cloudId, syncState, syncedCount, localId]);
    } else {
      await db.execute('UPDATE books SET cloud_id = ?, sync_state = ? WHERE id = ?', [cloudId, syncState, localId]);
    }
  }
  
  async updateChapterCloudId(localId: number, cloudId: number): Promise<void> {
    await db.execute('UPDATE chapters SET cloud_id = ? WHERE id = ?', [cloudId, localId]);
  }

  async getBooks(): Promise<LocalBook[]> {
    const rows = await db.select<any>('SELECT * FROM books ORDER BY created_at DESC');
    return rows.map(r => ({
      id: r.id,
      title: r.title,
      fingerprint: r.fingerprint,
      bookMD5: r.book_md5,
      cloudId: r.cloud_id,
      syncedCount: r.synced_count || 0,
      totalChapters: r.total_chapters,
      processStatus: r.process_status,
      platform: 'app',
      createdAt: r.created_at,
      syncState: r.sync_state || 0
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
      bookMD5: r.book_md5,
      cloudId: r.cloud_id,
      syncedCount: r.synced_count || 0,
      totalChapters: r.total_chapters,
      processStatus: r.process_status,
      platform: 'app',
      createdAt: r.created_at,
      syncState: r.sync_state || 0
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
        'INSERT INTO books (title, book_md5, fingerprint, total_chapters, process_status, synced_count, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)',
        [result.title, result.bookMD5, result.fingerprint, result.totalChapters, 'ready', 0, Date.now()]
      );
      
      const res = await db.select<any>('SELECT last_insert_rowid() as id');
      bookId = res[0].id;

      // Insert Chapters (批量插入优化)
      const total = result.chapters.length;
      const BATCH_SIZE = 200; // 激进一点，但不要超过 SQL 变量限制 (999)
      
      for (let i = 0; i < total; i += BATCH_SIZE) {
        const chunkStart = Date.now();
        const chunk = result.chapters.slice(i, i + BATCH_SIZE);
        const placeholders = chunk.map(() => '(?, ?, ?, ?, ?, ?)').join(',');
        const values = chunk.flatMap(c => [bookId, c.index, c.title, c.content, c.md5, c.length || 0]);
        
        await db.execute(
          `INSERT INTO chapters (book_id, chapter_index, title, content, md5, words_count) VALUES ${placeholders}`,
          values
        );
        
        // 更新入库进度
        if (onProgress) {
           const p = 80 + Math.floor(((i + chunk.length) / total) * 20);
           onProgress(p);
        }
      }
    });
    
    if (onProgress) onProgress(100);

    return {
      id: bookId,
      title: result.title,
      fingerprint: result.fingerprint,
      bookMD5: result.bookMD5,
      syncedCount: 0,
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

  async syncBookFromCloud(cloudBook: CloudBook): Promise<void> {
    const existing = await db.select<any>('SELECT id, cloud_id, sync_state FROM books WHERE book_md5 = ?', [cloudBook.book_md5]);

    if (existing.length > 0) {
      const existingBook = existing[0];
      console.log('[Sync] Book already exists locally:', cloudBook.title, 'sync_state:', existingBook.sync_state);

      // 如果是本地书籍（sync_state=0），更新为已同步状态
      if (existingBook.sync_state === 0 || existingBook.sync_state === undefined) {
        await db.execute(
          'UPDATE books SET cloud_id = ?, sync_state = 1, synced_count = total_chapters WHERE id = ?',
          [cloudBook.id, existingBook.id]
        );
        console.log('[Sync] Updated local book to synced state:', cloudBook.title);
      }
      // 如果已经是云端书籍或已同步书籍，不需要更新
      return;
    }

    // 创建新的云端书籍记录（降级模式，sync_state=2）
    await db.execute(
      'INSERT INTO books (cloud_id, book_md5, fingerprint, title, total_chapters, process_status, sync_state, synced_count, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)',
      [cloudBook.id, cloudBook.book_md5 || '', cloudBook.fingerprint, cloudBook.title, cloudBook.total_chapters, 'ready', 2, cloudBook.total_chapters, Date.now()]
    );

    console.log('[Sync] Created cloud-only book:', cloudBook.title);
  }

  async getChapters(bookId: number | string): Promise<LocalChapter[]> {
    const rows = await db.select<any>('SELECT id, chapter_index, title, md5, words_count FROM chapters WHERE book_id = ? ORDER BY chapter_index ASC', [bookId]);
    return rows.map(r => ({
      id: r.id,
      bookId: bookId,
      index: r.chapter_index,
      title: r.title,
      md5: r.md5,
      words_count: r.words_count || 0
    }));
  }

  async getChaptersBatch(bookId: number | string, offset: number, limit: number): Promise<LocalChapter[]> {
    const rows = await db.select<any>(
        'SELECT id, chapter_index, title, md5, content, words_count FROM chapters WHERE book_id = ? ORDER BY chapter_index ASC LIMIT ? OFFSET ?', 
        [bookId, limit, offset]
    );
    
    return rows.map(r => ({
      id: r.id,
      bookId: bookId,
      index: r.chapter_index,
      title: r.title,
      md5: r.md5,
      content: r.content,
      words_count: r.words_count || 0
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
  async createBook(title: string, fingerprint: string, total: number, bookMD5: string): Promise<number> {
    const existing = await db.select<any>('SELECT id, title, sync_state FROM books WHERE book_md5 = ?', [bookMD5]);

    if (existing.length > 0) {
      const existingBook = existing[0];
      if (existingBook.sync_state === 2) {
        throw new Error(`该书籍已存在于云端，无需重复导入`);
      } else {
        throw new Error(`该书籍已存在于本地，无需重复导入`);
      }
    }

    await db.execute(
      'INSERT INTO books (title, book_md5, fingerprint, total_chapters, process_status, created_at) VALUES (?, ?, ?, ?, ?, ?)',
      [title, bookMD5, fingerprint, total, 'ready', Date.now()]
    );
    const res = await db.select<any>('SELECT last_insert_rowid() as id');
    return res[0].id;
  }

  async insertChapters(bookId: number, chapters: any[]): Promise<void> {
    const placeholders = chapters.map(() => '(?, ?, ?, ?, ?, ?)').join(',');
    const values = chapters.flatMap(c => [bookId, c.index, c.title, c.content, c.md5, c.length || 0]);
    
    await db.execute(
      `INSERT INTO chapters (book_id, chapter_index, title, content, md5, words_count) VALUES ${placeholders}`,
      values
    );
  }

  async updateProgress(bookId: number | string, chapterId: number | string): Promise<void> {
    // 暂未实现
  }
}
