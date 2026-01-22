import { db } from "../utils/sqlite";
import type { IBookRepository } from "../core/repository";
import type { LocalBook, LocalChapter } from "../core/types";
import type { CloudBook } from "../core/repository";
import { BASE_URL } from "../api";

declare const plus: any;

// BookContentManifestChapter 表示清单中的单章信息。
interface BookContentManifestChapter {
  local_id?: number;
  chapter_id?: number;
  index: number;
  title: string;
  chapter_md5: string;
  size: number;
  words_count?: number;
  file_name?: string;
  offset?: number;
  length?: number;
}

// BookContentManifest 表示全量下载的清单结构。
interface BookContentManifest {
  book_id: number;
  book_name: string;
  total_chapters: number;
  chapters: BookContentManifestChapter[];
}

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
        user_id INTEGER DEFAULT 0,
        book_md5 TEXT,
        title TEXT,
        total_chapters INTEGER,
        sync_state INTEGER DEFAULT 0,
        created_at INTEGER
      )
    `);

    await this.ensureBookUserColumn();



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
      const hasContentColumn = columns.some(
        (col: any) => col.name === "content",
      );

      if (hasContentColumn) {
        console.log("[Migration] Old chapters table detected, migrating...");

        const chapters = await db.select<any>("SELECT * FROM chapters");

        await db.execute("DROP TABLE IF EXISTS chapters_backup");
        await db.execute(
          "CREATE TABLE chapters_backup AS SELECT * FROM chapters",
        );
        await db.execute("DROP TABLE chapters");

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

        for (const ch of chapters) {
          await db.execute(
            "INSERT INTO chapters (id, book_id, cloud_id, chapter_index, title, md5, words_count) VALUES (?, ?, ?, ?, ?, ?, ?)",
            [
              ch.id,
              ch.book_id,
              ch.cloud_id,
              ch.chapter_index,
              ch.title,
              ch.md5,
              ch.words_count,
            ],
          );

          // 迁移内容到 contents 表
          if (ch.md5 && ch.content) {
            await db.execute(
              "INSERT OR REPLACE INTO contents (chapter_md5, raw_content) VALUES (?, ?)",
              [ch.md5, ch.content],
            );
          }
        }

        await db.execute("DROP TABLE chapters_backup");
        console.log("[Migration] Database migration completed");
      }
    } catch (e) {
      console.error("[Migration] Migration failed:", e);
    }
  }

  private async ensureBookUserColumn(): Promise<void> {
    const columns = await db.select<any>("PRAGMA table_info(books)");
    const hasUserColumn = columns.some((col: any) => col.name === "user_id");
    if (!hasUserColumn) {
      await db.execute("ALTER TABLE books ADD COLUMN user_id INTEGER DEFAULT 0");
    }
  }

  async getBooks(): Promise<LocalBook[]> {
    const rows = await db.select<any>(
      "SELECT * FROM books ORDER BY created_at DESC",
    );
    return rows.map((r) => ({
      id: r.id,
      title: r.title,
      bookMD5: r.book_md5,
      cloudId: r.cloud_id,
      totalChapters: r.total_chapters,
      userId: r.user_id || 0,
      platform: "app",
      createdAt: r.created_at,
      syncState: r.sync_state || 0,
    }));

  }

  async getBook(id: number | string): Promise<LocalBook | null> {
    const rows = await db.select<any>("SELECT * FROM books WHERE id = ?", [id]);
    if (rows.length === 0) return null;
    const r = rows[0];
    return {
      id: r.id,
      title: r.title,
      bookMD5: r.book_md5,
      cloudId: r.cloud_id,
      totalChapters: r.total_chapters,
      userId: r.user_id || 0,
      platform: "app",
      createdAt: r.created_at,
      syncState: r.sync_state || 0,
    };
  }

  async syncBookFromCloud(cloudBook: CloudBook, userId: number): Promise<void> {
    const existing = await db.select<any>(
      "SELECT id, sync_state FROM books WHERE cloud_id = ?",
      [cloudBook.id],
    );

    if (existing.length > 0) {
      const syncState = existing[0].sync_state;
      if (syncState === 0) {
        await db.execute(
          "UPDATE books SET sync_state = 1, user_id = ? WHERE id = ?",
          [userId, existing[0].id],
        );
      } else {
        await db.execute("UPDATE books SET user_id = ? WHERE id = ?", [
          userId,
          existing[0].id,
        ]);
      }
      return;
    }

    await db.execute(
      "INSERT INTO books (cloud_id, user_id, book_md5, title, total_chapters, sync_state, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
      [
        cloudBook.id,
        userId,
        cloudBook.book_md5 || "",
        cloudBook.title,
        cloudBook.total_chapters,
        2,
        Date.now(),
      ],
    );
  }

  // downloadBookContent 下载整本书籍内容并落库。
  async downloadBookContent(
    bookId: number,
    cloudBookId: number,
    userId: number,
    onProgress?: (progress: number) => void,
  ): Promise<void> {
    if (!cloudBookId) {
      throw new Error("云端书籍ID为空");
    }

    console.log("[DownloadZip] Start", { bookId, cloudBookId });
    const token = uni.getStorageSync("token");
    const url = `${BASE_URL}/books/${cloudBookId}/content-db`;
    let currentProgress = 0;
    const setProgress = (value: number) => {
      const next = Math.min(100, Math.max(currentProgress, value));
      currentProgress = next;
      if (onProgress) {
        onProgress(next);
      }
    };

    let downloadTimer: number | null = null;
    const startDownloadTimer = () => {
      if (downloadTimer !== null) return;
      downloadTimer = setInterval(() => {
        if (currentProgress >= 8) return;
        setProgress(currentProgress + 1);
      }, 300) as unknown as number;
    };
    const stopDownloadTimer = () => {
      if (downloadTimer !== null) {
        clearInterval(downloadTimer as unknown as number);
        downloadTimer = null;
      }
    };

    startDownloadTimer();
    const tempFilePath = await new Promise<string>((resolve, reject) => {
      const task = uni.downloadFile({
        url,
        header: token ? { Authorization: `Bearer ${token}` } : {},
        success: (res) => {
          stopDownloadTimer();
          if (res.statusCode === 200 && res.tempFilePath) {
            console.log("[DownloadZip] Download complete", {
              tempFilePath: res.tempFilePath,
            });
            setProgress(40);
            resolve(res.tempFilePath);
          } else {
            reject(new Error(`下载失败: ${res.statusCode}`));
          }
        },
        fail: (err) => {
          stopDownloadTimer();
          reject(err);
        },
      });

      task.onProgressUpdate((res) => {
        const mapped = Math.floor(res.progress * 0.4);
        setProgress(mapped);
        if (res.progress % 20 === 0) {
          console.log("[DownloadZip] Download progress", {
            progress: res.progress,
            totalBytes: res.totalBytesWritten,
          });
        }
      });
    });

    const baseDir = `_doc/book_content/${cloudBookId}_${Date.now()}`;
    console.log("[DownloadZip] Decompress start", { baseDir });
    await this.ensureDir(baseDir);

    let decompressTimer: number | null = null;
    const startDecompressTimer = () => {
      if (decompressTimer !== null) return;
      decompressTimer = setInterval(() => {
        if (currentProgress >= 58) return;
        setProgress(currentProgress + 1);
      }, 200) as unknown as number;
    };
    const stopDecompressTimer = () => {
      if (decompressTimer !== null) {
        clearInterval(decompressTimer as unknown as number);
        decompressTimer = null;
      }
    };

    startDecompressTimer();
    try {
      await this.decompressZip(tempFilePath, baseDir);
    } finally {
      stopDecompressTimer();
    }
    setProgress(60);
    console.log("[DownloadZip] Decompress done", { baseDir });

    try {
      const dbPath = `${baseDir}/book.db`;
      console.log("[DownloadZip] Merge database start", {
        dbPath,
      });
      await this.mergeDownloadDatabase(bookId, cloudBookId, userId, dbPath, (value: number) => {
        setProgress(value);
      });
      console.log("[DownloadZip] Merge database done");
    } finally {
      await this.cleanupDir(baseDir);
    }
  }

  async getChapters(bookId: number | string): Promise<LocalChapter[]> {
    const rows = await db.select<any>(
      "SELECT id, chapter_index, title, md5, cloud_id, words_count FROM chapters WHERE book_id = ? ORDER BY chapter_index ASC",
      [bookId],
    );
    return rows.map((r) => ({
      id: r.id,
      bookId: bookId,
      index: r.chapter_index,
      title: r.title,
      md5: r.md5,
      cloudId: r.cloud_id,
      words_count: r.words_count || 0,
    }));
  }

  async getChaptersBatch(
    bookId: number | string,
    offset: number,
    limit: number,
  ): Promise<LocalChapter[]> {
    const rows = await db.select<any>(
      "SELECT c.id, c.chapter_index, c.title, c.md5, c.cloud_id, cnt.raw_content as content, c.words_count FROM chapters c LEFT JOIN contents cnt ON c.md5 = cnt.chapter_md5 WHERE c.book_id = ? ORDER BY c.chapter_index ASC LIMIT ? OFFSET ?",
      [bookId, limit, offset],
    );

    return rows.map((r) => ({
      id: r.id,
      bookId: bookId,
      index: r.chapter_index,
      title: r.title,
      md5: r.md5,
      content: r.content || "",
      words_count: r.words_count || 0,
    }));
  }

  async getChapterContent(
    bookId: number | string,
    chapterId: number | string,
  ): Promise<string> {
    const rows = await db.select<any>("SELECT md5 FROM chapters WHERE id = ?", [
      chapterId,
    ]);
    if (rows.length === 0) return "";

    const md5 = rows[0].md5;
    const contentRows = await db.select<any>(
      "SELECT raw_content FROM contents WHERE chapter_md5 = ?",
      [md5],
    );
    return contentRows.length > 0 ? contentRows[0].raw_content : "";
  }

  // uploadBookZip 打包并上传本地书籍。
  async uploadBookZip(
    bookId: number,
    onProgress?: (progress: number) => void,
  ): Promise<{
    book_id: number;
    chapter_mappings: Array<{ local_id: number; cloud_id: number }>;
  }> {
    console.log("[UploadZip] Start", { bookId });
    const startAt = Date.now();

    const book = await this.getBook(bookId);
    if (!book) {
      throw new Error("书籍不存在");
    }

    const total = book.totalChapters || 0;
    if (total === 0) {
      throw new Error("章节为空");
    }

    const contentParts: string[] = [];
    let offsetBytes = 0;

    const manifest: BookContentManifest = {
      book_id: book.cloudId || 0,
      book_name: book.title,
      total_chapters: total,
      chapters: [],
    };

    const batchSize = 200;
    let offset = 0;
    while (offset < total) {
      const batchStart = Date.now();
      const batch = await this.getChaptersBatch(bookId, offset, batchSize);
      const batchCost = Date.now() - batchStart;
      console.log("[UploadZip] Read chapters batch", {
        offset,
        size: batch.length,
        costMs: batchCost,
      });
      if (batch.length === 0) break;

      for (const chapter of batch) {
        const content = chapter.content || "";
        const byteLength = this.getUtf8ByteLength(content);
        manifest.chapters.push({
          local_id: Number(chapter.id),
          index: chapter.index,
          title: chapter.title,
          chapter_md5: chapter.md5 || "",
          size: byteLength,
          words_count: chapter.words_count || 0,
          offset: offsetBytes,
          length: byteLength,
        });
        contentParts.push(content);
        offsetBytes += byteLength;
      }

      offset += batch.length;
      if (onProgress) {
        onProgress(Math.floor((offset / total) * 60));
      }
    }

    const baseDir = `_doc/upload_books/${bookId}_${Date.now()}`;
    await this.ensureDir(baseDir);

    const bookContent = contentParts.join("");
    const bookPath = `${baseDir}/book.txt`;
    const manifestPath = `${baseDir}/manifest.json`;
    console.log("[UploadZip] Write book file", { bookPath });
    await this.writeTextFile(bookPath, bookContent);
    await this.writeTextFile(manifestPath, JSON.stringify(manifest));

    const bookStat = await this.getFileStat(bookPath);
    const manifestStat = await this.getFileStat(manifestPath);
    console.log("[UploadZip] Book files ready", {
      bookSize: bookStat.size,
      manifestSize: manifestStat.size,
    });
    if (bookStat.size === 0 || manifestStat.size === 0) {
      throw new Error("book.txt 或 manifest.json 写入失败");
    }

    await new Promise((resolve) => setTimeout(resolve, 200));

    const zipPath = `${baseDir}.zip`;
    console.log("[UploadZip] Compress zip", { zipPath, baseDir });
    const zipStart = Date.now();
    try {
      await this.compressZip(baseDir, zipPath);
      console.log("[UploadZip] Compress zip done", {
        costMs: Date.now() - zipStart,
      });

      const zipStat = await this.getFileStat(zipPath);
      console.log("[UploadZip] Generate zip done", {
        costMs: Date.now() - zipStart,
        size: zipStat.size,
        sizeMB: Number((zipStat.size / 1024 / 1024).toFixed(2)),
      });

      if (onProgress) {
        onProgress(85);
      }

      console.log("[UploadZip] Upload zip start", {
        totalCostMs: Date.now() - startAt,
      });
      const response = await this.uploadZipFilePath(
        zipPath,
        book,
        total,
        onProgress,
      );
      console.log("[UploadZip] Upload success", {
        response,
        totalCostMs: Date.now() - startAt,
      });
      return response;
    } finally {
      await this.cleanupDir(baseDir);
    }
  }

  async createBook(
    title: string,
    total: number,
    bookMD5: string,
  ): Promise<number> {
    console.log("[Repo] createBook called:", { title, total, bookMD5 });
    const existing = await db.select<any>(
      "SELECT id, title, sync_state, book_md5 FROM books WHERE book_md5 = ?",
      [bookMD5],
    );

    if (existing.length > 0) {
      console.log("[Repo] Book already exists:", existing[0]);
      const existingBook = existing[0];
      if (existingBook.sync_state === 2) {
        throw new Error(`该书籍已存在于云端，无需重复导入`);
      } else {
        throw new Error(`该书籍已存在于本地，无需重复导入`);
      }
    }

    await db.execute(
      "INSERT INTO books (user_id, title, book_md5, total_chapters, created_at) VALUES (?, ?, ?, ?, ?)",
      [0, title, bookMD5, total, Date.now()],
    );
    const res = await db.select<any>("SELECT last_insert_rowid() as id");
    console.log("[Repo] Book created with ID:", res[0].id);
    return res[0].id;
  }

  async insertChapters(bookId: number, chapters: any[]): Promise<void> {
    await db.transaction(async () => {
      await db.execute(
        `INSERT INTO chapters (book_id, chapter_index, title, md5, words_count) VALUES ${chapters.map(() => "(?, ?, ?, ?, ?)").join(",")}`,
        chapters.flatMap((c) => [
          bookId,
          c.index,
          c.title,
          c.md5,
          c.length || 0,
        ]),
      );

      await db.execute(
        `INSERT OR REPLACE INTO contents (chapter_md5, raw_content) VALUES ${chapters.map(() => "(?, ?)").join(",")}`,
        chapters.flatMap((c) => [c.md5, c.content]),
      );
    });
  }

  // resetBookChaptersForDownload 清理下载前章节数据。
  async resetBookChaptersForDownload(bookId: number): Promise<void> {
    await db.execute("DELETE FROM chapters WHERE book_id = ?", [bookId]);
  }

  // insertDownloadedChapters 批量写入下载章节。
  async insertDownloadedChapters(
    bookId: number,
    chapters: Array<{
      chapter_id: number;
      index: number;
      title: string;
      chapter_md5: string;
      content: string;
      length?: number;
    }>,
  ): Promise<void> {
    if (chapters.length === 0) {
      return;
    }

    await db.transaction(async () => {
      await db.execute(
        `INSERT INTO chapters (book_id, cloud_id, chapter_index, title, md5, words_count) VALUES ${chapters.map(() => "(?, ?, ?, ?, ?, ?)").join(",")}`,
        chapters.flatMap((c) => [
          bookId,
          c.chapter_id,
          c.index,
          c.title,
          c.chapter_md5,
          c.length || 0,
        ]),
      );

      await db.execute(
        `INSERT OR REPLACE INTO contents (chapter_md5, raw_content) VALUES ${chapters.map(() => "(?, ?)").join(",")}`,
        chapters.flatMap((c) => [c.chapter_md5, c.content]),
      );
    });
  }

  // finalizeDownload 更新下载完成状态。
  async finalizeDownload(
    bookId: number,
    cloudBookId: number,
    totalChapters: number,
  ): Promise<void> {
    await db.execute(
      "UPDATE books SET cloud_id = ?, sync_state = 1, total_chapters = ? WHERE id = ?",
      [cloudBookId, totalChapters, bookId],
    );
  }

  // cleanupTempDir 清理下载临时目录。
  async cleanupTempDir(dirPath: string): Promise<void> {
    await this.cleanupDir(dirPath);
  }

  async deleteBook(id: number | string): Promise<void> {
    const chapters = await db.select<any>(
      "SELECT md5 FROM chapters WHERE book_id = ?",
      [id],
    );
    const md5s = chapters.map((c) => c.md5);

    await db.transaction(async () => {
      await db.execute("DELETE FROM chapters WHERE book_id = ?", [id]);
      await db.execute("DELETE FROM books WHERE id = ?", [id]);
      await db.execute("DELETE FROM reading_history WHERE book_id = ?", [id]);

      if (md5s.length > 0) {
        await db.execute(
          `DELETE FROM contents WHERE chapter_md5 IN (${md5s.map(() => "?").join(",")})`,
          md5s,
        );
      }
    });
  }

  async updateProgress(
    bookId: number | string,
    chapterId: number | string,
    promptId: number = 0,
  ): Promise<void> {
    await db.execute(
      "INSERT OR REPLACE INTO reading_history (book_id, last_chapter_id, last_prompt_id, updated_at) VALUES (?, ?, ?, ?)",
      [bookId, chapterId, promptId, Date.now()],
    );
  }

  async getReadingHistory(bookId: number | string): Promise<{
    last_chapter_id: number;
    last_prompt_id: number;
    updated_at: number;
  } | null> {
    const rows = await db.select<any>(
      "SELECT last_chapter_id, last_prompt_id, updated_at FROM reading_history WHERE book_id = ?",
      [bookId],
    );
    if (rows.length === 0) return null;
    return {
      last_chapter_id: rows[0].last_chapter_id,
      last_prompt_id: rows[0].last_prompt_id,
      updated_at: rows[0].updated_at,
    };
  }

  // ensureDir 确保目录存在。
  private async ensureDir(dirPath: string): Promise<void> {
    await this.getDirectoryEntry(dirPath);
  }

  // decompressZip 解压压缩包到指定目录。
  private async decompressZip(
    zipPath: string,
    targetDir: string,
  ): Promise<void> {
    await new Promise<void>((resolve, reject) => {
      if (!plus?.zip) {
        reject(new Error("当前环境不支持解压"));
        return;
      }
      plus.zip.decompress(
        zipPath,
        targetDir,
        () => resolve(),
        (err: any) => reject(err),
      );
    });
  }

  // readTextFile 读取文本文件内容。
  private async readTextFile(filePath: string): Promise<string> {
    const entry = await this.resolveFileEntry(filePath);
    return new Promise<string>((resolve, reject) => {
      entry.file((file: any) => {
        const reader = new plus.io.FileReader();
        reader.onloadend = () => resolve(reader.result as string);
        reader.onerror = (err: any) => reject(err);
        reader.readAsText(file, "utf-8");
      }, reject);
    });
  }

  // readTextRange 读取指定范围的文本内容。
  private async readTextRange(
    filePath: string,
    start: number,
    end: number,
  ): Promise<string> {
    const entry = await this.resolveFileEntry(filePath);
    return new Promise<string>((resolve, reject) => {
      entry.file((file: any) => {
        const slice = file.slice(start, end);
        const startAt = Date.now();
        const timeout = setTimeout(() => {
          reject(new Error(`读取范围超时: ${filePath}`));
        }, 10000);

        const reader = new plus.io.FileReader();
        reader.onloadend = () => {
          clearTimeout(timeout);
          resolve(reader.result as string);
        };
        reader.onerror = (err: any) => {
          clearTimeout(timeout);
          reject(err);
        };
        reader.readAsText(slice, "utf-8");
        console.log("[DownloadZip] Read range text", {
          filePath,
          start,
          end,
          costMs: Date.now() - startAt,
        });
      }, reject);
    });
  }

  // getFileStat 获取文件大小。
  private async getFileStat(filePath: string): Promise<{ size: number }> {
    const entry = await this.resolveFileEntry(filePath);
    return new Promise((resolve, reject) => {
      entry.file((file: any) => {
        resolve({ size: file.size || 0 });
      }, reject);
    });
  }

  // uploadZipFilePath 上传压缩包文件路径。
  private async uploadZipFilePath(
    zipPath: string,
    book: LocalBook,
    totalChapters: number,
    onProgress?: (progress: number) => void,
  ): Promise<{
    book_id: number;
    chapter_mappings: Array<{ local_id: number; cloud_id: number }>;
  }> {
    const query = `book_name=${encodeURIComponent(book.title)}&book_md5=${encodeURIComponent(book.bookMD5 || "")}&total_chapters=${totalChapters}`;
    const url = `${BASE_URL}/books/upload-zip?${query}`;
    const token = uni.getStorageSync("token");
    const requestStart = Date.now();

    return new Promise((resolve, reject) => {
      const task = uni.uploadFile({
        url,
        filePath: zipPath,
        name: "file",
        header: {
          Authorization: token ? `Bearer ${token}` : "",
        },
        success: (res) => {
          let data: { code: number; msg: string; data: any } | null = null;
          try {
            data =
              typeof res.data === "string"
                ? JSON.parse(res.data)
                : (res.data as any);
          } catch (e) {
            reject(new Error("上传响应解析失败"));
            return;
          }
          console.log("[UploadZip] Upload response", {
            status: res.statusCode,
            costMs: Date.now() - requestStart,
          });
          if (res.statusCode === 200 && data && data.code === 0) {
            resolve(data.data);
          } else {
            reject(new Error((data && data.msg) || "上传失败"));
          }
        },
        fail: (err) => {
          console.warn("[UploadZip] Upload failed", {
            costMs: Date.now() - requestStart,
            err,
          });
          reject(err);
        },
      });

      task.onProgressUpdate?.((progress) => {
        if (onProgress) {
          onProgress(85 + Math.floor((progress.progress / 100) * 15));
        }
      });
    });
  }

  // writeTextFile 写入文本内容。
  private async writeTextFile(
    filePath: string,
    content: string,
  ): Promise<void> {
    const fileEntry = await this.createFileEntry(filePath);
    await new Promise<void>((resolve, reject) => {
      fileEntry.createWriter((writer: any) => {
        writer.onwriteend = () => resolve();
        writer.onerror = (err: any) => reject(err);
        writer.write(content);
      }, reject);
    });
  }

  // compressZip 压缩目录为 zip 文件。
  private async compressZip(
    sourceDir: string,
    targetZip: string,
  ): Promise<void> {
    await new Promise<void>((resolve, reject) => {
      if (!plus?.zip) {
        reject(new Error("当前环境不支持压缩"));
        return;
      }
      plus.zip.compress(
        sourceDir,
        targetZip,
        () => resolve(),
        (err: any) => reject(err),
      );
    });
  }

  // createFileEntry 创建文件入口。
  private async createFileEntry(filePath: string): Promise<any> {
    const normalized = filePath.replace(/\\/g, "/");
    const parts = normalized.split("/").filter(Boolean);
    const fileName = parts.pop();
    if (!fileName) {
      throw new Error("文件路径非法");
    }
    const dirPath = parts.join("/");
    const directory = await this.getDirectoryEntry(dirPath || "_doc");
    return new Promise((resolve, reject) => {
      directory.getFile(fileName, { create: true }, resolve, reject);
    });
  }

  // getDirectoryEntry 获取目录入口并确保存在。
  private async getDirectoryEntry(dirPath: string): Promise<any> {
    const normalized = dirPath.replace(/\\/g, "/");
    const parts = normalized.split("/").filter(Boolean);
    const rootDir = await this.resolveDirectoryRoot(normalized);
    const childParts = parts[0] === "_doc" ? parts.slice(1) : parts;
    let current = rootDir;

    for (const part of childParts) {
      current = await new Promise((resolve, reject) => {
        current.getDirectory(part, { create: true }, resolve, reject);
      });
    }

    return current;
  }

  // uploadZipBuffer 上传压缩包并返回服务端结果。
  private async uploadZipBuffer(
    zipData: ArrayBuffer,
    book: LocalBook,
    totalChapters: number,
    onProgress?: (progress: number) => void,
  ): Promise<{
    book_id: number;
    chapter_mappings: Array<{ local_id: number; cloud_id: number }>;
  }> {
    const query = `book_name=${encodeURIComponent(book.title)}&book_md5=${encodeURIComponent(book.bookMD5 || "")}&total_chapters=${totalChapters}`;
    const url = `${BASE_URL}/books/upload-zip?${query}`;
    const token = uni.getStorageSync("token");
    const requestStart = Date.now();

    console.log("[UploadZip] Upload request", {
      size: zipData.byteLength,
      sizeMB: Number((zipData.byteLength / 1024 / 1024).toFixed(2)),
    });

    return new Promise((resolve, reject) => {
      uni.request({
        url,
        method: "POST",
        data: zipData,
        header: {
          "Content-Type": "application/zip",
          Authorization: token ? `Bearer ${token}` : "",
        },
        responseType: "json",
        timeout: 120000,
        success: (res) => {
          const data = res.data as { code: number; msg: string; data: any };
          console.log("[UploadZip] Upload response", {
            status: res.statusCode,
            costMs: Date.now() - requestStart,
          });
          if (res.statusCode === 200 && data.code === 0) {
            resolve(data.data);
          } else {
            reject(new Error(data.msg || "上传失败"));
          }
        },
        fail: (err) => {
          console.warn("[UploadZip] Upload failed", {
            costMs: Date.now() - requestStart,
            err,
          });
          reject(err);
        },
      });

      if (onProgress) {
        onProgress(90);
      }
    });
  }

  // readManifest 读取并解析下载清单。
  private async readManifest(filePath: string): Promise<BookContentManifest> {
    const content = await this.readTextFile(filePath);
    const data = JSON.parse(content) as BookContentManifest;
    if (!data || !Array.isArray(data.chapters)) {
      throw new Error("清单解析失败");
    }
    return data;
  }

  // mergeDownloadDatabase 合并下载的 SQLite 数据库。
  private async mergeDownloadDatabase(
    bookId: number,
    cloudBookId: number,
    userId: number,
    dbPath: string,
    onProgress?: (progress: number) => void,
  ): Promise<void> {
    const updateProgress = (value: number) => {
      if (onProgress) {
        onProgress(Math.min(100, Math.max(60, value)));
      }
    };

    if (!plus?.io) {
      throw new Error("当前环境不支持文件系统");
    }

    const alias = "download";
    const mergeStart = Date.now();
    const tempName = `story_trim_download_${Date.now()}.db`;
    const tempPath = `_doc/${tempName}`;
    const tempNativePath = await this.copyDownloadDbToDoc(dbPath, tempPath);
    const safePath = tempNativePath.replace(/'/g, "''");
    const attachSQL = `ATTACH DATABASE '${safePath}' AS ${alias}`;
    const detachSQL = `DETACH DATABASE ${alias}`;
    let attached = false;

    await db.execute(attachSQL);
    attached = true;
    try {
      const totals = await db.select<{ total: number }>(
        `SELECT COUNT(*) as total FROM ${alias}.chapters`,
      );
      const total = totals[0]?.total || 0;
      if (total === 0) {
        throw new Error("下载章节为空");
      }

      await db.transaction(async () => {
        await db.execute("DELETE FROM chapters WHERE book_id = ?", [bookId]);

        const batchSize = 200;
        let processed = 0;
        let offset = 0;
        let batchIndex = 0;
        const logEvery = 5;

        while (offset < total) {
          const rows = await db.select<{
            chapter_id: number;
            chapter_index: number;
            title: string;
            chapter_md5: string;
            words_count: number;
            raw_content: string;
          }>(
            `SELECT c.chapter_id as chapter_id, c.chapter_index as chapter_index, c.title as title, c.chapter_md5 as chapter_md5, c.words_count as words_count, cnt.raw_content as raw_content FROM ${alias}.chapters c LEFT JOIN ${alias}.contents cnt ON c.chapter_md5 = cnt.chapter_md5 ORDER BY c.chapter_index ASC LIMIT ? OFFSET ?`,
            [batchSize, offset],
          );
          if (rows.length === 0) {
            break;
          }

          const chapterBatch: any[] = [];
          const contentBatch: any[] = [];
          for (const row of rows) {
            const rawContent = row.raw_content || "";
            const wordsCount = row.words_count || rawContent.length;
            chapterBatch.push([
              bookId,
              row.chapter_id,
              row.chapter_index,
              row.title,
              row.chapter_md5,
              wordsCount,
            ]);
            contentBatch.push([row.chapter_md5, rawContent]);
          }

          const writeStart = Date.now();
          await db.execute(
            `INSERT INTO chapters (book_id, cloud_id, chapter_index, title, md5, words_count) VALUES ${chapterBatch
              .map(() => "(?, ?, ?, ?, ?, ?)")
              .join(",")}`,
            chapterBatch.flat(),
          );
          await db.execute(
            `INSERT OR REPLACE INTO contents (chapter_md5, raw_content) VALUES ${contentBatch
              .map(() => "(?, ?)")
              .join(",")}`,
            contentBatch.flat(),
          );
          if (batchIndex % logEvery === 0) {
            console.log("[DownloadZip] Merge batch", {
              count: rows.length,
              costMs: Date.now() - writeStart,
            });
          }

          processed += rows.length;
          offset += rows.length;
          batchIndex += 1;
          updateProgress(60 + Math.floor((processed / total) * 40));
        }

        await db.execute(
          "UPDATE books SET cloud_id = ?, sync_state = 1, total_chapters = ?, user_id = ? WHERE id = ?",
          [cloudBookId, total, userId, bookId],
        );
      });

      updateProgress(100);
      console.log("[DownloadZip] Merge database cost", {
        costMs: Date.now() - mergeStart,
        chapters: total,
      });
    } finally {
      if (attached) {
        await db.execute(detachSQL);
      }
      await this.removeFile(tempPath);
    }
  }

  // copyDownloadDbToDoc 将下载数据库复制到可附加路径。
  private async copyDownloadDbToDoc(
    sourcePath: string,
    targetPath: string,
  ): Promise<string> {
    const sourceEntry = await this.resolveFileEntry(sourcePath);
    const targetDir = await this.resolveDirectoryEntry("_doc");

    await new Promise<void>((resolve, reject) => {
      sourceEntry.copyTo(targetDir, targetPath.replace(/^_doc\//, ""), () => {
        resolve();
      }, (err: any) => reject(err));
    });

    const targetEntry = await this.resolveFileEntry(targetPath);
    const nativeURL = targetEntry.nativeURL || targetEntry.fullPath || targetPath;
    return nativeURL.startsWith("file://") ? nativeURL.replace("file://", "") : nativeURL;
  }

  // removeFile 删除临时文件。
  private async removeFile(filePath: string): Promise<void> {
    try {
      const entry = await this.resolveFileEntry(filePath);
      await new Promise<void>((resolve) => {
        entry.remove(() => resolve(), () => resolve());
      });
    } catch (e) {
      console.warn("[Repo] 删除临时文件失败", e);
    }
  }

  // getUtf8ByteLength 计算 UTF-8 字节长度。
  private getUtf8ByteLength(content: string): number {
    return unescape(encodeURIComponent(content)).length;
  }

  // decodeUtf8 解码 UTF-8 字节内容。
  private decodeUtf8(bytes: Uint8Array): string {
    let binary = "";
    const chunkSize = 8192;
    for (let i = 0; i < bytes.length; i += chunkSize) {
      const chunk = bytes.subarray(i, i + chunkSize);
      binary += String.fromCharCode(...chunk);
    }
    return decodeURIComponent(escape(binary));
  }

  // extractChunkContents 从合并文件中提取章节内容。
  private extractChunkContents(
    text: string,
    chapters: BookContentManifestChapter[],
  ): Map<number, string> {
    const sorted = [...chapters].sort(
      (a, b) => (a.offset || 0) - (b.offset || 0),
    );
    const result = new Map<number, string>();
    if (sorted.length === 0) {
      return result;
    }

    const byteString = unescape(encodeURIComponent(text));
    for (const chapter of sorted) {
      if (!chapter.chapter_id) {
        continue;
      }
      const offset = chapter.offset ?? 0;
      const length = chapter.length ?? 0;
      if (length <= 0) {
        continue;
      }
      const slice = byteString.slice(offset, offset + length);
      const content = decodeURIComponent(escape(slice));
      result.set(chapter.chapter_id, content);
    }
    return result;
  }

  // cleanupDir 清理临时目录。
  private async cleanupDir(dirPath: string): Promise<void> {
    try {
      const entry = await this.resolveDirectoryEntry(dirPath);
      await new Promise<void>((resolve) => {
        entry.removeRecursively(
          () => resolve(),
          () => resolve(),
        );
      });
    } catch (e) {
      console.warn("[Repo] 清理临时目录失败", e);
    }
  }

  // resolveDirectoryRoot 获取根目录入口。
  private async resolveDirectoryRoot(dirPath: string): Promise<any> {
    if (!plus?.io) {
      throw new Error("当前环境不支持文件系统");
    }
    const normalized = dirPath.replace(/\\/g, "/");
    if (normalized.startsWith("_doc")) {
      return this.resolveDirectoryEntry("_doc");
    }
    return this.resolveDirectoryEntry(normalized);
  }

  // resolveDirectoryEntry 解析目录入口。
  private async resolveDirectoryEntry(dirPath: string): Promise<any> {
    return new Promise((resolve, reject) => {
      plus.io.resolveLocalFileSystemURL(dirPath, resolve, reject);
    });
  }

  // resolveFileEntry 解析文件入口。
  private async resolveFileEntry(filePath: string): Promise<any> {
    return new Promise((resolve, reject) => {
      plus.io.resolveLocalFileSystemURL(filePath, resolve, reject);
    });
  }
}
