// 仅在 App 端生效
const DB_NAME = 'story_trim.db';
const DB_PATH = '_doc/story_trim.db';

// 简易 SQL 转义与拼接 (修复 Android 基座 args 绑定失效问题)
function escape(val: any): string {
  if (val === null || val === undefined) return 'NULL';
  if (typeof val === 'number') return val.toString();
  // 转义单引号：' -> ''
  return `'${String(val).replace(/'/g, "''")}'`;
}

function formatSql(sql: string, args: any[]): string {
  if (!args || args.length === 0) return sql;
  let i = 0;
  // 简单的 ? 替换，不处理字符串内的 ?
  return sql.replace(/\?/g, () => {
    return i < args.length ? escape(args[i++]) : 'NULL';
  });
}

export const db = {
  // 打开数据库
  open(): Promise<void> {
    return new Promise((resolve, reject) => {
      // #ifdef APP-PLUS
      if (plus.sqlite.isOpenDatabase({ name: DB_NAME, path: DB_PATH })) {
        resolve();
        return;
      }
      plus.sqlite.openDatabase({
        name: DB_NAME,
        path: DB_PATH,
        success: () => resolve(),
        fail: (e) => reject(e)
      });
      // #endif
      
      // #ifndef APP-PLUS
      console.warn('SQLite is not supported on this platform');
      resolve();
      // #endif
    });
  },

  // 执行非查询 SQL (INSERT, UPDATE, DELETE, CREATE)
  execute(sql: string, args: any[] = []): Promise<void> {
    const finalSql = formatSql(sql, args);
    return new Promise((resolve, reject) => {
      // #ifdef APP-PLUS
      // console.log('[SQLite Exec]', finalSql.substring(0, 200)) 
      plus.sqlite.executeSql({
        name: DB_NAME,
        sql: finalSql,
        success: () => resolve(),
        fail: (e) => {
          console.error('[SQL Error]', finalSql.substring(0, 500), e);
          reject(e);
        }
      });
      // #endif
      // #ifndef APP-PLUS
      resolve();
      // #endif
    });
  },

  // 执行查询 SQL (SELECT)
  select<T>(sql: string, args: any[] = []): Promise<T[]> {
    const finalSql = formatSql(sql, args);
    return new Promise((resolve, reject) => {
      // #ifdef APP-PLUS
      plus.sqlite.selectSql({
        name: DB_NAME,
        sql: finalSql,
        success: (res) => resolve(res as T[]),
        fail: (e) => {
          console.error('[SQL Select Error]', finalSql, e);
          reject(e);
        }
      });
      // #endif
      // #ifndef APP-PLUS
      resolve([] as T[]);
      // #endif
    });
  },

  // 事务封装
  async transaction(fn: () => Promise<void>) {
    // 事务不需要转义，直接执行
    await this.execute('BEGIN TRANSACTION');
    try {
      await fn();
      await this.execute('COMMIT');
    } catch (e) {
      await this.execute('ROLLBACK');
      throw e;
    }
  }
};