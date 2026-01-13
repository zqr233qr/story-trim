import { db } from '../utils/sqlite';
import { parser } from '../utils/parser';
import type { IBookRepository } from '../core/repository';
import type { LocalBook, LocalChapter, TrimmedContent } from '../core/types';
import type { CloudBook } from '../core/repository';

export class AppRepository implements IBookRepository {
  async init(): Promise<void> {
    await db.open();

    // 数据库迁移：将旧表的数据迁移到新结构
    await this.migrateDatabase();

    // 1. Books Table
    await db.execute(`
      CREATE TABLE IF NOT EXISTS books (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        cloud_id INTEGER DEFAULT 0,
        book_md5 TEXT,
        title TEXT,
        total_chapters INTEGER,
        process_status TEXT,
        sync_state INTEGER DEFAULT 0,
        synced_count INTEGER DEFAULT 0,
        created_at INTEGER
      )
    `);

    // 2. Chapters Table (只存索引，内容存 contents 表)
    await db.execute(`
      CREATE TABLE IF NOT EXISTS chapters (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        book_id INTEGER,
        cloud_id INTEGER DEFAULT 0,
        chapter_index INTEGER,
        title TEXT,
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

  private async migrateDatabase(): Promise<void> {
    try {
      const columns = await db.select<any>("PRAGMA table_info(chapters)");
      const hasContentColumn = columns.some((col: any) => col.name === 'content');

      if (hasContentColumn) {
        console.log('[Migration] Old chapters table detected, migrating...');

        const chapters = await db.select<any>('SELECT * FROM chapters');

        await db.execute('DROP TABLE IF EXISTS chapters_backup');
        await db.execute('CREATE TABLE chapters_backup AS SELECT * FROM chapters');
        await db.execute('DROP TABLE chapters');

        // 创建新表
        await db.execute(`
          CREATE TABLE chapters (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            book_id INTEGER,
            cloud_id INTEGER DEFAULT 0,
            chapter_index INTEGER,
            title TEXT,
            md5 TEXT,
            words_count INTEGER DEFAULT 0
          )
        `);

        // 迁移数据到新表
        for (const ch of chapters) {
          await db.execute(
            'INSERT INTO chapters (id, book_id, cloud_id, chapter_index, title, md5, words_count) VALUES (?, ?, ?, ?, ?, ?, ?)',
            [ch.id, ch.book_id, ch.cloud_id, ch.chapter_index, ch.title, ch.md5, ch.words_count]
          );

          // 迁移内容到 contents 表
          if (ch.md5 && ch.content) {
            await db.execute(
              'INSERT OR REPLACE INTO contents (chapter_md5, raw_content) VALUES (?, ?)',
              [ch.md5, ch.content]
            );
          }
        }

        await db.execute('DROP TABLE chapters_backup');
        console.log('[Migration] Database migration completed');
      }
    } catch (e) {
      console.error('[Migration] Migration failed:', e);
    }
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
    const result = await parser.parseFile(filePath, fileName, onProgress);

    let bookId = 0;
    await db.transaction(async () => {
      await db.execute('PRAGMA synchronous = OFF');
      await db.execute('PRAGMA journal_mode = MEMORY');

      await db.execute(
        'INSERT INTO books (title, book_md5, total_chapters, process_status, synced_count, created_at) VALUES (?, ?, ?, ?, ?, ?)',
        [result.title, result.bookMD5, result.totalChapters, 'ready', 0, Date.now()]
      );

      const res = await db.select<any>('SELECT last_insert_rowid() as id');
      bookId = res[0].id;

      const total = result.chapters.length;
      const BATCH_SIZE = 200;

      for (let i = 0; i < total; i += BATCH_SIZE) {
        const chunk = result.chapters.slice(i, i + BATCH_SIZE);
        const placeholders = chunk.map(() => '(?, ?, ?, ?, ?)').join(',');

        await db.execute(
          `INSERT INTO chapters (book_id, chapter_index, title, md5, words_count) VALUES ${placeholders}`,
          chunk.flatMap(c => [bookId, c.index, c.title, c.md5, c.length || 0])
        );

        await db.execute(
          `INSERT OR REPLACE INTO contents (chapter_md5, raw_content) VALUES ${chunk.map(() => '(?, ?)').join(',')}`,
          chunk.flatMap(c => [c.md5, c.content])
        );

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
      const chapters = await db.select<any>('SELECT md5 FROM chapters WHERE book_id = ?', [id]);
      const md5s = chapters.map(c => c.md5);

      await db.execute('DELETE FROM chapters WHERE book_id = ?', [id]);
      await db.execute('DELETE FROM books WHERE id = ?', [id]);
      await db.execute('DELETE FROM reading_history WHERE book_id = ?', [id]);

      if (md5s.length > 0) {
        await db.execute(`DELETE FROM contents WHERE chapter_md5 IN (${md5s.map(() => '?').join(',')})`, md5s);
      }
    });
  }

  async syncBookFromCloud(cloudBook: CloudBook): Promise<void> {
    const existing = await db.select<any>('SELECT id, cloud_id, sync_state FROM books WHERE book_md5 = ?', [cloudBook.book_md5]);

    if (existing.length > 0) {
      const existingBook = existing[0];
      console.log('[Sync] Book already exists locally:', cloudBook.title, 'sync_state:', existingBook.sync_state);

      if (existingBook.sync_state === 0 || existingBook.sync_state === undefined) {
        await db.execute(
          'UPDATE books SET cloud_id = ?, sync_state = 1, synced_count = total_chapters WHERE id = ?',
          [cloudBook.id, existingBook.id]
        );
        console.log('[Sync] Updated local book to synced state:', cloudBook.title);
      }
      return;
    }

    await db.execute(
      'INSERT INTO books (cloud_id, book_md5, title, total_chapters, process_status, sync_state, synced_count, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)',
      [cloudBook.id, cloudBook.book_md5 || '', cloudBook.title, cloudBook.total_chapters, 'ready', 2, 0, Date.now()]
    );

    console.log('[Sync] Created cloud-only book:', cloudBook.title);
  }

  async syncChaptersFromCloud(localBookId: number, cloudChapters: any[]): Promise<void> {
    await db.transaction(async () => {
      for (const cloudCh of cloudChapters) {
        await db.execute(
          'INSERT OR REPLACE INTO chapters (book_id, cloud_id, chapter_index, title, md5, words_count) VALUES (?, ?, ?, ?, ?, ?)',
          [localBookId, cloudCh.id, cloudCh.index, cloudCh.title, cloudCh.chapter_md5, cloudCh.words_count || 0]
        );
      }
    });
  }

  async getChapters(bookId: number | string): Promise<LocalChapter[]> {
    const rows = await db.select<any>('SELECT id, chapter_index, title, md5, cloud_id, words_count FROM chapters WHERE book_id = ? ORDER BY chapter_index ASC', [bookId]);
    return rows.map(r => ({
      id: r.id,
      bookId: bookId,
      index: r.chapter_index,
      title: r.title,
      md5: r.md5,
      cloudId: r.cloud_id,
      words_count: r.words_count || 0
    }));
  }

  async getChaptersBatch(bookId: number | string, offset: number, limit: number): Promise<LocalChapter[]> {
    const rows = await db.select<any>(
      'SELECT c.id, c.chapter_index, c.title, c.md5, c.cloud_id, cnt.raw_content as content, c.words_count FROM chapters c LEFT JOIN contents cnt ON c.md5 = cnt.chapter_md5 WHERE c.book_id = ? ORDER BY c.chapter_index ASC LIMIT ? OFFSET ?',
      [bookId, limit, offset]
    );

    return rows.map(r => ({
      id: r.id,
      bookId: bookId,
      index: r.chapter_index,
      title: r.title,
      md5: r.md5,
      content: r.content || '',
      words_count: r.words_count || 0
    }));
  }

  async getChapterContent(bookId: number | string, chapterId: number | string): Promise<string> {
    const rows = await db.select<any>('SELECT md5 FROM chapters WHERE id = ?', [chapterId]);
    if (rows.length === 0) return '';

    const md5 = rows[0].md5;
    const contentRows = await db.select<any>('SELECT raw_content FROM contents WHERE chapter_md5 = ?', [md5]);
    return contentRows.length > 0 ? contentRows[0].raw_content : '';
  }

  async saveChapterContent(md5: string, content: string): Promise<void> {
    await db.execute(
      'INSERT OR REPLACE INTO contents (chapter_md5, raw_content) VALUES (?, ?)',
      [md5, content]
    );
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

  async createBook(title: string, total: number, bookMD5: string): Promise<number> {
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
      'INSERT INTO books (title, book_md5, total_chapters, process_status, created_at) VALUES (?, ?, ?, ?, ?)',
      [title, bookMD5, total, 'ready', Date.now()]
    );
    const res = await db.select<any>('SELECT last_insert_rowid() as id');
    return res[0].id;
  }

  async insertChapters(bookId: number, chapters: any[]): Promise<void> {
    await db.transaction(async () => {
      await db.execute(
        `INSERT INTO chapters (book_id, chapter_index, title, md5, words_count) VALUES ${chapters.map(() => '(?, ?, ?, ?, ?)').join(',')}`,
        chapters.flatMap(c => [bookId, c.index, c.title, c.md5, c.length || 0])
      );

      await db.execute(
        `INSERT OR REPLACE INTO contents (chapter_md5, raw_content) VALUES ${chapters.map(() => '(?, ?)').join(',')}`,
        chapters.flatMap(c => [c.md5, c.content])
      );
    });
  }

  async updateProgress(bookId: number | string, chapterId: number | string, promptId: number = 0): Promise<void> {
    await db.execute(
      'INSERT OR REPLACE INTO reading_history (book_id, last_chapter_id, last_prompt_id, updated_at) VALUES (?, ?, ?, ?)',
      [bookId, chapterId, promptId, Date.now()]
    );
  }

  async getReadingHistory(bookId: number | string): Promise<{
    last_chapter_id: number;
    last_prompt_id: number;
    updated_at: number;
  } | null> {
    const rows = await db.select<any>(
      'SELECT last_chapter_id, last_prompt_id, updated_at FROM reading_history WHERE book_id = ?',
      [bookId]
    );
    if (rows.length === 0) return null;
    return {
      last_chapter_id: rows[0].last_chapter_id,
      last_prompt_id: rows[0].last_prompt_id,
      updated_at: rows[0].updated_at
    };
  }
}
