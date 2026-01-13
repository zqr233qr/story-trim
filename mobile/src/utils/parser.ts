// #ifdef APP-PLUS
import SparkMD5 from 'spark-md5';

// 匹配章节标题：(行首或换行符) + 空白(可选) + 第xxx章 + (非换行符的内容)
const CHAPTER_HEADER_REGEX = /(?:^|\n)\s*(第[0-9一二三四五六七八九十百千万]+[章回节][^\r\n]*)/g;
const CHUNK_SIZE = 1024 * 1024; // 1MB

export interface ParsedChapter {
  index: number;
  title: string;
  content: string;
  md5: string;
  length: number;
}

export interface ParseResult {
  title: string;
  totalChapters: number;
  bookMD5: string;
  chapters: ParsedChapter[];
}

export const parser = {
  async parseFile(filePath: string, fileName: string, onProgress?: (p: number) => void): Promise<ParseResult> {
    console.log('[Parser] V5 Loaded - Standard MD5 Mode');
    return new Promise((resolve, reject) => {
      plus.io.resolveLocalFileSystemURL(filePath, (entry) => {
        entry.file(async (file) => {
          const reader = new plus.io.FileReader();
          const fileSize = file.size;
          let offset = 0;
          
          let chapters: ParsedChapter[] = [];
          let currentTitle = '序章';
          let currentBuffer: string[] = []; 
          let chapterIndex = 0;
          const spark = new SparkMD5.ArrayBuffer(); // 用于计算全文 MD5
          
          const commitChapter = (title: string, buffer: string[]) => {
            const content = buffer.join(''); 
            // 允许空章，但至少要有内容
            if (content.length === 0 && buffer.length === 0) return;

            const md5 = SparkMD5.hash(content); 
            
            chapters.push({
              index: chapterIndex++,
              title: title,
              content: content,
              md5: md5,
              length: content.length
            });
          };

          const readNextChunk = () => {
            const startRead = Date.now();
            if (onProgress) {
              onProgress(Math.floor((offset / fileSize) * 80));
            }

            if (offset >= fileSize) {
              console.log(`[Parser] Finished. Total chapters: ${chapters.length}`);
              commitChapter(currentTitle, currentBuffer);
              if (onProgress) onProgress(80);
              const bookMD5 = spark.end();
              console.log('[Parser] Book MD5:', bookMD5);
              resolve({
                title: fileName.replace(/\.txt$/i, ''),
                totalChapters: chapters.length,
                bookMD5: bookMD5,
                chapters: chapters
              });
              return;
            }

            const end = Math.min(offset + CHUNK_SIZE, fileSize);
            const slice = file.slice(offset, end);
            
            reader.onload = (e) => {
              const readTime = Date.now() - startRead;
              
              const arrayBuffer = e.target.result as ArrayBuffer;
              
              spark.append(arrayBuffer);
              
              const decoder = new TextDecoder('utf-8');
              const text = decoder.decode(arrayBuffer);
              
              let validEnd = text.length;
              if (end < fileSize) {
                validEnd = text.lastIndexOf('\n');
                if (validEnd <= 0) validEnd = text.length; 
              }

              const validChunk = text.substring(0, validEnd);
              
              const matchStart = Date.now();
              const matches = [...validChunk.matchAll(CHAPTER_HEADER_REGEX)];
              const matchTime = Date.now() - matchStart;
              
              console.log(`[Parser] Chunk: ${offset/1024}KB. Read: ${readTime}ms. Regex: ${matchTime}ms. Matches: ${matches.length}`);

              if (matches.length === 0) {
                currentBuffer.push(validChunk);
              } else {
                const firstMatch = matches[0];
                if (firstMatch.index! > 0) {
                  currentBuffer.push(validChunk.substring(0, firstMatch.index));
                }
                
                commitChapter(currentTitle, currentBuffer);
                
                for (let i = 0; i < matches.length - 1; i++) {
                  const m = matches[i];
                  const nextM = matches[i + 1];
                  const title = m[1].trim();
                  const start = m.index! + m[0].length;
                  const end = nextM.index!;
                  const content = validChunk.substring(start, end);
                  commitChapter(title, [content]);
                }
                
                const lastMatch = matches[matches.length - 1];
                currentTitle = lastMatch[1].trim();
                const lastContentStart = lastMatch.index! + lastMatch.length;
                currentBuffer = [validChunk.substring(lastContentStart)];
              }

              offset += validEnd;
              if (validEnd === 0 && end < fileSize) offset += CHUNK_SIZE;

              setTimeout(readNextChunk, 0); 
            };
            
            reader.onerror = (e) => {
              console.error('[Parser] Error reading file:', e);
              reject(e);
            };
            
            reader.readAsArrayBuffer(slice);
          };

          readNextChunk();

        }, (e) => reject(e));
      }, (e) => reject(e));
    });
  }
};
// #endif

// #ifndef APP-PLUS
export const parser = {
  async parseFile(filePath: string, fileName: string): Promise<any> {
    return { title: fileName, chapters: [] };
  }
}
// #endif
