package com.storytrim.app.core.database;

import androidx.annotation.NonNull;
import androidx.room.DatabaseConfiguration;
import androidx.room.InvalidationTracker;
import androidx.room.RoomDatabase;
import androidx.room.RoomOpenHelper;
import androidx.room.migration.AutoMigrationSpec;
import androidx.room.migration.Migration;
import androidx.room.util.DBUtil;
import androidx.room.util.TableInfo;
import androidx.sqlite.db.SupportSQLiteDatabase;
import androidx.sqlite.db.SupportSQLiteOpenHelper;
import com.storytrim.app.core.database.dao.BookDao;
import com.storytrim.app.core.database.dao.BookDao_Impl;
import com.storytrim.app.core.database.dao.ChapterDao;
import com.storytrim.app.core.database.dao.ChapterDao_Impl;
import com.storytrim.app.core.database.dao.ContentDao;
import com.storytrim.app.core.database.dao.ContentDao_Impl;
import com.storytrim.app.core.database.dao.ReadingHistoryDao;
import com.storytrim.app.core.database.dao.ReadingHistoryDao_Impl;
import java.lang.Class;
import java.lang.Override;
import java.lang.String;
import java.lang.SuppressWarnings;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import javax.annotation.processing.Generated;

@Generated("androidx.room.RoomProcessor")
@SuppressWarnings({"unchecked", "deprecation"})
public final class AppDatabase_Impl extends AppDatabase {
  private volatile BookDao _bookDao;

  private volatile ChapterDao _chapterDao;

  private volatile ContentDao _contentDao;

  private volatile ReadingHistoryDao _readingHistoryDao;

  @Override
  @NonNull
  protected SupportSQLiteOpenHelper createOpenHelper(@NonNull final DatabaseConfiguration config) {
    final SupportSQLiteOpenHelper.Callback _openCallback = new RoomOpenHelper(config, new RoomOpenHelper.Delegate(3) {
      @Override
      public void createAllTables(@NonNull final SupportSQLiteDatabase db) {
        db.execSQL("CREATE TABLE IF NOT EXISTS `books` (`id` INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, `cloud_id` INTEGER NOT NULL, `user_id` INTEGER NOT NULL, `book_md5` TEXT NOT NULL, `title` TEXT NOT NULL, `total_chapters` INTEGER NOT NULL, `sync_state` INTEGER NOT NULL, `created_at` INTEGER NOT NULL)");
        db.execSQL("CREATE TABLE IF NOT EXISTS `chapters` (`id` INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, `book_id` INTEGER NOT NULL, `cloud_id` INTEGER NOT NULL, `chapter_index` INTEGER NOT NULL, `title` TEXT NOT NULL, `md5` TEXT NOT NULL, `words_count` INTEGER NOT NULL)");
        db.execSQL("CREATE TABLE IF NOT EXISTS `contents` (`chapter_md5` TEXT NOT NULL, `raw_content` TEXT NOT NULL, PRIMARY KEY(`chapter_md5`))");
        db.execSQL("CREATE TABLE IF NOT EXISTS `reading_history` (`book_id` INTEGER NOT NULL, `last_chapter_id` INTEGER NOT NULL, `last_prompt_id` INTEGER NOT NULL, `updated_at` INTEGER NOT NULL, PRIMARY KEY(`book_id`))");
        db.execSQL("CREATE TABLE IF NOT EXISTS room_master_table (id INTEGER PRIMARY KEY,identity_hash TEXT)");
        db.execSQL("INSERT OR REPLACE INTO room_master_table (id,identity_hash) VALUES(42, '15fc214164ddc03252cd8a8cbd8493f0')");
      }

      @Override
      public void dropAllTables(@NonNull final SupportSQLiteDatabase db) {
        db.execSQL("DROP TABLE IF EXISTS `books`");
        db.execSQL("DROP TABLE IF EXISTS `chapters`");
        db.execSQL("DROP TABLE IF EXISTS `contents`");
        db.execSQL("DROP TABLE IF EXISTS `reading_history`");
        final List<? extends RoomDatabase.Callback> _callbacks = mCallbacks;
        if (_callbacks != null) {
          for (RoomDatabase.Callback _callback : _callbacks) {
            _callback.onDestructiveMigration(db);
          }
        }
      }

      @Override
      public void onCreate(@NonNull final SupportSQLiteDatabase db) {
        final List<? extends RoomDatabase.Callback> _callbacks = mCallbacks;
        if (_callbacks != null) {
          for (RoomDatabase.Callback _callback : _callbacks) {
            _callback.onCreate(db);
          }
        }
      }

      @Override
      public void onOpen(@NonNull final SupportSQLiteDatabase db) {
        mDatabase = db;
        internalInitInvalidationTracker(db);
        final List<? extends RoomDatabase.Callback> _callbacks = mCallbacks;
        if (_callbacks != null) {
          for (RoomDatabase.Callback _callback : _callbacks) {
            _callback.onOpen(db);
          }
        }
      }

      @Override
      public void onPreMigrate(@NonNull final SupportSQLiteDatabase db) {
        DBUtil.dropFtsSyncTriggers(db);
      }

      @Override
      public void onPostMigrate(@NonNull final SupportSQLiteDatabase db) {
      }

      @Override
      @NonNull
      public RoomOpenHelper.ValidationResult onValidateSchema(
          @NonNull final SupportSQLiteDatabase db) {
        final HashMap<String, TableInfo.Column> _columnsBooks = new HashMap<String, TableInfo.Column>(8);
        _columnsBooks.put("id", new TableInfo.Column("id", "INTEGER", true, 1, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsBooks.put("cloud_id", new TableInfo.Column("cloud_id", "INTEGER", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsBooks.put("user_id", new TableInfo.Column("user_id", "INTEGER", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsBooks.put("book_md5", new TableInfo.Column("book_md5", "TEXT", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsBooks.put("title", new TableInfo.Column("title", "TEXT", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsBooks.put("total_chapters", new TableInfo.Column("total_chapters", "INTEGER", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsBooks.put("sync_state", new TableInfo.Column("sync_state", "INTEGER", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsBooks.put("created_at", new TableInfo.Column("created_at", "INTEGER", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        final HashSet<TableInfo.ForeignKey> _foreignKeysBooks = new HashSet<TableInfo.ForeignKey>(0);
        final HashSet<TableInfo.Index> _indicesBooks = new HashSet<TableInfo.Index>(0);
        final TableInfo _infoBooks = new TableInfo("books", _columnsBooks, _foreignKeysBooks, _indicesBooks);
        final TableInfo _existingBooks = TableInfo.read(db, "books");
        if (!_infoBooks.equals(_existingBooks)) {
          return new RoomOpenHelper.ValidationResult(false, "books(com.storytrim.app.core.database.entity.BookEntity).\n"
                  + " Expected:\n" + _infoBooks + "\n"
                  + " Found:\n" + _existingBooks);
        }
        final HashMap<String, TableInfo.Column> _columnsChapters = new HashMap<String, TableInfo.Column>(7);
        _columnsChapters.put("id", new TableInfo.Column("id", "INTEGER", true, 1, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsChapters.put("book_id", new TableInfo.Column("book_id", "INTEGER", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsChapters.put("cloud_id", new TableInfo.Column("cloud_id", "INTEGER", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsChapters.put("chapter_index", new TableInfo.Column("chapter_index", "INTEGER", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsChapters.put("title", new TableInfo.Column("title", "TEXT", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsChapters.put("md5", new TableInfo.Column("md5", "TEXT", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsChapters.put("words_count", new TableInfo.Column("words_count", "INTEGER", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        final HashSet<TableInfo.ForeignKey> _foreignKeysChapters = new HashSet<TableInfo.ForeignKey>(0);
        final HashSet<TableInfo.Index> _indicesChapters = new HashSet<TableInfo.Index>(0);
        final TableInfo _infoChapters = new TableInfo("chapters", _columnsChapters, _foreignKeysChapters, _indicesChapters);
        final TableInfo _existingChapters = TableInfo.read(db, "chapters");
        if (!_infoChapters.equals(_existingChapters)) {
          return new RoomOpenHelper.ValidationResult(false, "chapters(com.storytrim.app.core.database.entity.ChapterEntity).\n"
                  + " Expected:\n" + _infoChapters + "\n"
                  + " Found:\n" + _existingChapters);
        }
        final HashMap<String, TableInfo.Column> _columnsContents = new HashMap<String, TableInfo.Column>(2);
        _columnsContents.put("chapter_md5", new TableInfo.Column("chapter_md5", "TEXT", true, 1, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsContents.put("raw_content", new TableInfo.Column("raw_content", "TEXT", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        final HashSet<TableInfo.ForeignKey> _foreignKeysContents = new HashSet<TableInfo.ForeignKey>(0);
        final HashSet<TableInfo.Index> _indicesContents = new HashSet<TableInfo.Index>(0);
        final TableInfo _infoContents = new TableInfo("contents", _columnsContents, _foreignKeysContents, _indicesContents);
        final TableInfo _existingContents = TableInfo.read(db, "contents");
        if (!_infoContents.equals(_existingContents)) {
          return new RoomOpenHelper.ValidationResult(false, "contents(com.storytrim.app.core.database.entity.ContentEntity).\n"
                  + " Expected:\n" + _infoContents + "\n"
                  + " Found:\n" + _existingContents);
        }
        final HashMap<String, TableInfo.Column> _columnsReadingHistory = new HashMap<String, TableInfo.Column>(4);
        _columnsReadingHistory.put("book_id", new TableInfo.Column("book_id", "INTEGER", true, 1, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsReadingHistory.put("last_chapter_id", new TableInfo.Column("last_chapter_id", "INTEGER", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsReadingHistory.put("last_prompt_id", new TableInfo.Column("last_prompt_id", "INTEGER", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        _columnsReadingHistory.put("updated_at", new TableInfo.Column("updated_at", "INTEGER", true, 0, null, TableInfo.CREATED_FROM_ENTITY));
        final HashSet<TableInfo.ForeignKey> _foreignKeysReadingHistory = new HashSet<TableInfo.ForeignKey>(0);
        final HashSet<TableInfo.Index> _indicesReadingHistory = new HashSet<TableInfo.Index>(0);
        final TableInfo _infoReadingHistory = new TableInfo("reading_history", _columnsReadingHistory, _foreignKeysReadingHistory, _indicesReadingHistory);
        final TableInfo _existingReadingHistory = TableInfo.read(db, "reading_history");
        if (!_infoReadingHistory.equals(_existingReadingHistory)) {
          return new RoomOpenHelper.ValidationResult(false, "reading_history(com.storytrim.app.core.database.entity.ReadingHistoryEntity).\n"
                  + " Expected:\n" + _infoReadingHistory + "\n"
                  + " Found:\n" + _existingReadingHistory);
        }
        return new RoomOpenHelper.ValidationResult(true, null);
      }
    }, "15fc214164ddc03252cd8a8cbd8493f0", "5d7a0c6270d2f5a0b2eef64e6cffcc99");
    final SupportSQLiteOpenHelper.Configuration _sqliteConfig = SupportSQLiteOpenHelper.Configuration.builder(config.context).name(config.name).callback(_openCallback).build();
    final SupportSQLiteOpenHelper _helper = config.sqliteOpenHelperFactory.create(_sqliteConfig);
    return _helper;
  }

  @Override
  @NonNull
  protected InvalidationTracker createInvalidationTracker() {
    final HashMap<String, String> _shadowTablesMap = new HashMap<String, String>(0);
    final HashMap<String, Set<String>> _viewTables = new HashMap<String, Set<String>>(0);
    return new InvalidationTracker(this, _shadowTablesMap, _viewTables, "books","chapters","contents","reading_history");
  }

  @Override
  public void clearAllTables() {
    super.assertNotMainThread();
    final SupportSQLiteDatabase _db = super.getOpenHelper().getWritableDatabase();
    try {
      super.beginTransaction();
      _db.execSQL("DELETE FROM `books`");
      _db.execSQL("DELETE FROM `chapters`");
      _db.execSQL("DELETE FROM `contents`");
      _db.execSQL("DELETE FROM `reading_history`");
      super.setTransactionSuccessful();
    } finally {
      super.endTransaction();
      _db.query("PRAGMA wal_checkpoint(FULL)").close();
      if (!_db.inTransaction()) {
        _db.execSQL("VACUUM");
      }
    }
  }

  @Override
  @NonNull
  protected Map<Class<?>, List<Class<?>>> getRequiredTypeConverters() {
    final HashMap<Class<?>, List<Class<?>>> _typeConvertersMap = new HashMap<Class<?>, List<Class<?>>>();
    _typeConvertersMap.put(BookDao.class, BookDao_Impl.getRequiredConverters());
    _typeConvertersMap.put(ChapterDao.class, ChapterDao_Impl.getRequiredConverters());
    _typeConvertersMap.put(ContentDao.class, ContentDao_Impl.getRequiredConverters());
    _typeConvertersMap.put(ReadingHistoryDao.class, ReadingHistoryDao_Impl.getRequiredConverters());
    return _typeConvertersMap;
  }

  @Override
  @NonNull
  public Set<Class<? extends AutoMigrationSpec>> getRequiredAutoMigrationSpecs() {
    final HashSet<Class<? extends AutoMigrationSpec>> _autoMigrationSpecsSet = new HashSet<Class<? extends AutoMigrationSpec>>();
    return _autoMigrationSpecsSet;
  }

  @Override
  @NonNull
  public List<Migration> getAutoMigrations(
      @NonNull final Map<Class<? extends AutoMigrationSpec>, AutoMigrationSpec> autoMigrationSpecs) {
    final List<Migration> _autoMigrations = new ArrayList<Migration>();
    return _autoMigrations;
  }

  @Override
  public BookDao bookDao() {
    if (_bookDao != null) {
      return _bookDao;
    } else {
      synchronized(this) {
        if(_bookDao == null) {
          _bookDao = new BookDao_Impl(this);
        }
        return _bookDao;
      }
    }
  }

  @Override
  public ChapterDao chapterDao() {
    if (_chapterDao != null) {
      return _chapterDao;
    } else {
      synchronized(this) {
        if(_chapterDao == null) {
          _chapterDao = new ChapterDao_Impl(this);
        }
        return _chapterDao;
      }
    }
  }

  @Override
  public ContentDao contentDao() {
    if (_contentDao != null) {
      return _contentDao;
    } else {
      synchronized(this) {
        if(_contentDao == null) {
          _contentDao = new ContentDao_Impl(this);
        }
        return _contentDao;
      }
    }
  }

  @Override
  public ReadingHistoryDao readingHistoryDao() {
    if (_readingHistoryDao != null) {
      return _readingHistoryDao;
    } else {
      synchronized(this) {
        if(_readingHistoryDao == null) {
          _readingHistoryDao = new ReadingHistoryDao_Impl(this);
        }
        return _readingHistoryDao;
      }
    }
  }
}
