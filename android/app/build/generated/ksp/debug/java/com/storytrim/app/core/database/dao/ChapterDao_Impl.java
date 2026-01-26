package com.storytrim.app.core.database.dao;

import android.database.Cursor;
import android.os.CancellationSignal;
import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.room.CoroutinesRoom;
import androidx.room.EntityInsertionAdapter;
import androidx.room.RoomDatabase;
import androidx.room.RoomSQLiteQuery;
import androidx.room.SharedSQLiteStatement;
import androidx.room.util.CursorUtil;
import androidx.room.util.DBUtil;
import androidx.sqlite.db.SupportSQLiteStatement;
import com.storytrim.app.core.database.entity.ChapterEntity;
import java.lang.Class;
import java.lang.Exception;
import java.lang.Long;
import java.lang.Object;
import java.lang.Override;
import java.lang.String;
import java.lang.SuppressWarnings;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.concurrent.Callable;
import javax.annotation.processing.Generated;
import kotlin.Unit;
import kotlin.coroutines.Continuation;
import kotlinx.coroutines.flow.Flow;

@Generated("androidx.room.RoomProcessor")
@SuppressWarnings({"unchecked", "deprecation"})
public final class ChapterDao_Impl implements ChapterDao {
  private final RoomDatabase __db;

  private final EntityInsertionAdapter<ChapterEntity> __insertionAdapterOfChapterEntity;

  private final SharedSQLiteStatement __preparedStmtOfDeleteChaptersByBookId;

  public ChapterDao_Impl(@NonNull final RoomDatabase __db) {
    this.__db = __db;
    this.__insertionAdapterOfChapterEntity = new EntityInsertionAdapter<ChapterEntity>(__db) {
      @Override
      @NonNull
      protected String createQuery() {
        return "INSERT OR REPLACE INTO `chapters` (`id`,`book_id`,`cloud_id`,`chapter_index`,`title`,`md5`,`words_count`) VALUES (nullif(?, 0),?,?,?,?,?,?)";
      }

      @Override
      protected void bind(@NonNull final SupportSQLiteStatement statement,
          @NonNull final ChapterEntity entity) {
        statement.bindLong(1, entity.getId());
        statement.bindLong(2, entity.getBookId());
        statement.bindLong(3, entity.getCloudId());
        statement.bindLong(4, entity.getChapterIndex());
        statement.bindString(5, entity.getTitle());
        statement.bindString(6, entity.getMd5());
        statement.bindLong(7, entity.getWordsCount());
      }
    };
    this.__preparedStmtOfDeleteChaptersByBookId = new SharedSQLiteStatement(__db) {
      @Override
      @NonNull
      public String createQuery() {
        final String _query = "DELETE FROM chapters WHERE book_id = ?";
        return _query;
      }
    };
  }

  @Override
  public Object insertChapter(final ChapterEntity chapter,
      final Continuation<? super Long> $completion) {
    return CoroutinesRoom.execute(__db, true, new Callable<Long>() {
      @Override
      @NonNull
      public Long call() throws Exception {
        __db.beginTransaction();
        try {
          final Long _result = __insertionAdapterOfChapterEntity.insertAndReturnId(chapter);
          __db.setTransactionSuccessful();
          return _result;
        } finally {
          __db.endTransaction();
        }
      }
    }, $completion);
  }

  @Override
  public Object insertChapters(final List<ChapterEntity> chapters,
      final Continuation<? super Unit> $completion) {
    return CoroutinesRoom.execute(__db, true, new Callable<Unit>() {
      @Override
      @NonNull
      public Unit call() throws Exception {
        __db.beginTransaction();
        try {
          __insertionAdapterOfChapterEntity.insert(chapters);
          __db.setTransactionSuccessful();
          return Unit.INSTANCE;
        } finally {
          __db.endTransaction();
        }
      }
    }, $completion);
  }

  @Override
  public Object deleteChaptersByBookId(final long bookId,
      final Continuation<? super Unit> $completion) {
    return CoroutinesRoom.execute(__db, true, new Callable<Unit>() {
      @Override
      @NonNull
      public Unit call() throws Exception {
        final SupportSQLiteStatement _stmt = __preparedStmtOfDeleteChaptersByBookId.acquire();
        int _argIndex = 1;
        _stmt.bindLong(_argIndex, bookId);
        try {
          __db.beginTransaction();
          try {
            _stmt.executeUpdateDelete();
            __db.setTransactionSuccessful();
            return Unit.INSTANCE;
          } finally {
            __db.endTransaction();
          }
        } finally {
          __preparedStmtOfDeleteChaptersByBookId.release(_stmt);
        }
      }
    }, $completion);
  }

  @Override
  public Flow<List<ChapterEntity>> getChaptersByBookId(final long bookId) {
    final String _sql = "SELECT * FROM chapters WHERE book_id = ? ORDER BY chapter_index ASC";
    final RoomSQLiteQuery _statement = RoomSQLiteQuery.acquire(_sql, 1);
    int _argIndex = 1;
    _statement.bindLong(_argIndex, bookId);
    return CoroutinesRoom.createFlow(__db, false, new String[] {"chapters"}, new Callable<List<ChapterEntity>>() {
      @Override
      @NonNull
      public List<ChapterEntity> call() throws Exception {
        final Cursor _cursor = DBUtil.query(__db, _statement, false, null);
        try {
          final int _cursorIndexOfId = CursorUtil.getColumnIndexOrThrow(_cursor, "id");
          final int _cursorIndexOfBookId = CursorUtil.getColumnIndexOrThrow(_cursor, "book_id");
          final int _cursorIndexOfCloudId = CursorUtil.getColumnIndexOrThrow(_cursor, "cloud_id");
          final int _cursorIndexOfChapterIndex = CursorUtil.getColumnIndexOrThrow(_cursor, "chapter_index");
          final int _cursorIndexOfTitle = CursorUtil.getColumnIndexOrThrow(_cursor, "title");
          final int _cursorIndexOfMd5 = CursorUtil.getColumnIndexOrThrow(_cursor, "md5");
          final int _cursorIndexOfWordsCount = CursorUtil.getColumnIndexOrThrow(_cursor, "words_count");
          final List<ChapterEntity> _result = new ArrayList<ChapterEntity>(_cursor.getCount());
          while (_cursor.moveToNext()) {
            final ChapterEntity _item;
            final long _tmpId;
            _tmpId = _cursor.getLong(_cursorIndexOfId);
            final long _tmpBookId;
            _tmpBookId = _cursor.getLong(_cursorIndexOfBookId);
            final long _tmpCloudId;
            _tmpCloudId = _cursor.getLong(_cursorIndexOfCloudId);
            final int _tmpChapterIndex;
            _tmpChapterIndex = _cursor.getInt(_cursorIndexOfChapterIndex);
            final String _tmpTitle;
            _tmpTitle = _cursor.getString(_cursorIndexOfTitle);
            final String _tmpMd5;
            _tmpMd5 = _cursor.getString(_cursorIndexOfMd5);
            final int _tmpWordsCount;
            _tmpWordsCount = _cursor.getInt(_cursorIndexOfWordsCount);
            _item = new ChapterEntity(_tmpId,_tmpBookId,_tmpCloudId,_tmpChapterIndex,_tmpTitle,_tmpMd5,_tmpWordsCount);
            _result.add(_item);
          }
          return _result;
        } finally {
          _cursor.close();
        }
      }

      @Override
      protected void finalize() {
        _statement.release();
      }
    });
  }

  @Override
  public Object getChapterById(final long id,
      final Continuation<? super ChapterEntity> $completion) {
    final String _sql = "SELECT * FROM chapters WHERE id = ? LIMIT 1";
    final RoomSQLiteQuery _statement = RoomSQLiteQuery.acquire(_sql, 1);
    int _argIndex = 1;
    _statement.bindLong(_argIndex, id);
    final CancellationSignal _cancellationSignal = DBUtil.createCancellationSignal();
    return CoroutinesRoom.execute(__db, false, _cancellationSignal, new Callable<ChapterEntity>() {
      @Override
      @Nullable
      public ChapterEntity call() throws Exception {
        final Cursor _cursor = DBUtil.query(__db, _statement, false, null);
        try {
          final int _cursorIndexOfId = CursorUtil.getColumnIndexOrThrow(_cursor, "id");
          final int _cursorIndexOfBookId = CursorUtil.getColumnIndexOrThrow(_cursor, "book_id");
          final int _cursorIndexOfCloudId = CursorUtil.getColumnIndexOrThrow(_cursor, "cloud_id");
          final int _cursorIndexOfChapterIndex = CursorUtil.getColumnIndexOrThrow(_cursor, "chapter_index");
          final int _cursorIndexOfTitle = CursorUtil.getColumnIndexOrThrow(_cursor, "title");
          final int _cursorIndexOfMd5 = CursorUtil.getColumnIndexOrThrow(_cursor, "md5");
          final int _cursorIndexOfWordsCount = CursorUtil.getColumnIndexOrThrow(_cursor, "words_count");
          final ChapterEntity _result;
          if (_cursor.moveToFirst()) {
            final long _tmpId;
            _tmpId = _cursor.getLong(_cursorIndexOfId);
            final long _tmpBookId;
            _tmpBookId = _cursor.getLong(_cursorIndexOfBookId);
            final long _tmpCloudId;
            _tmpCloudId = _cursor.getLong(_cursorIndexOfCloudId);
            final int _tmpChapterIndex;
            _tmpChapterIndex = _cursor.getInt(_cursorIndexOfChapterIndex);
            final String _tmpTitle;
            _tmpTitle = _cursor.getString(_cursorIndexOfTitle);
            final String _tmpMd5;
            _tmpMd5 = _cursor.getString(_cursorIndexOfMd5);
            final int _tmpWordsCount;
            _tmpWordsCount = _cursor.getInt(_cursorIndexOfWordsCount);
            _result = new ChapterEntity(_tmpId,_tmpBookId,_tmpCloudId,_tmpChapterIndex,_tmpTitle,_tmpMd5,_tmpWordsCount);
          } else {
            _result = null;
          }
          return _result;
        } finally {
          _cursor.close();
          _statement.release();
        }
      }
    }, $completion);
  }

  @Override
  public Object getChapterByIndex(final long bookId, final int index,
      final Continuation<? super ChapterEntity> $completion) {
    final String _sql = "SELECT * FROM chapters WHERE book_id = ? AND chapter_index = ? LIMIT 1";
    final RoomSQLiteQuery _statement = RoomSQLiteQuery.acquire(_sql, 2);
    int _argIndex = 1;
    _statement.bindLong(_argIndex, bookId);
    _argIndex = 2;
    _statement.bindLong(_argIndex, index);
    final CancellationSignal _cancellationSignal = DBUtil.createCancellationSignal();
    return CoroutinesRoom.execute(__db, false, _cancellationSignal, new Callable<ChapterEntity>() {
      @Override
      @Nullable
      public ChapterEntity call() throws Exception {
        final Cursor _cursor = DBUtil.query(__db, _statement, false, null);
        try {
          final int _cursorIndexOfId = CursorUtil.getColumnIndexOrThrow(_cursor, "id");
          final int _cursorIndexOfBookId = CursorUtil.getColumnIndexOrThrow(_cursor, "book_id");
          final int _cursorIndexOfCloudId = CursorUtil.getColumnIndexOrThrow(_cursor, "cloud_id");
          final int _cursorIndexOfChapterIndex = CursorUtil.getColumnIndexOrThrow(_cursor, "chapter_index");
          final int _cursorIndexOfTitle = CursorUtil.getColumnIndexOrThrow(_cursor, "title");
          final int _cursorIndexOfMd5 = CursorUtil.getColumnIndexOrThrow(_cursor, "md5");
          final int _cursorIndexOfWordsCount = CursorUtil.getColumnIndexOrThrow(_cursor, "words_count");
          final ChapterEntity _result;
          if (_cursor.moveToFirst()) {
            final long _tmpId;
            _tmpId = _cursor.getLong(_cursorIndexOfId);
            final long _tmpBookId;
            _tmpBookId = _cursor.getLong(_cursorIndexOfBookId);
            final long _tmpCloudId;
            _tmpCloudId = _cursor.getLong(_cursorIndexOfCloudId);
            final int _tmpChapterIndex;
            _tmpChapterIndex = _cursor.getInt(_cursorIndexOfChapterIndex);
            final String _tmpTitle;
            _tmpTitle = _cursor.getString(_cursorIndexOfTitle);
            final String _tmpMd5;
            _tmpMd5 = _cursor.getString(_cursorIndexOfMd5);
            final int _tmpWordsCount;
            _tmpWordsCount = _cursor.getInt(_cursorIndexOfWordsCount);
            _result = new ChapterEntity(_tmpId,_tmpBookId,_tmpCloudId,_tmpChapterIndex,_tmpTitle,_tmpMd5,_tmpWordsCount);
          } else {
            _result = null;
          }
          return _result;
        } finally {
          _cursor.close();
          _statement.release();
        }
      }
    }, $completion);
  }

  @Override
  public Object getChaptersByBookIdPaged(final long bookId, final int limit, final int offset,
      final Continuation<? super List<ChapterEntity>> $completion) {
    final String _sql = "SELECT * FROM chapters WHERE book_id = ? ORDER BY chapter_index ASC LIMIT ? OFFSET ?";
    final RoomSQLiteQuery _statement = RoomSQLiteQuery.acquire(_sql, 3);
    int _argIndex = 1;
    _statement.bindLong(_argIndex, bookId);
    _argIndex = 2;
    _statement.bindLong(_argIndex, limit);
    _argIndex = 3;
    _statement.bindLong(_argIndex, offset);
    final CancellationSignal _cancellationSignal = DBUtil.createCancellationSignal();
    return CoroutinesRoom.execute(__db, false, _cancellationSignal, new Callable<List<ChapterEntity>>() {
      @Override
      @NonNull
      public List<ChapterEntity> call() throws Exception {
        final Cursor _cursor = DBUtil.query(__db, _statement, false, null);
        try {
          final int _cursorIndexOfId = CursorUtil.getColumnIndexOrThrow(_cursor, "id");
          final int _cursorIndexOfBookId = CursorUtil.getColumnIndexOrThrow(_cursor, "book_id");
          final int _cursorIndexOfCloudId = CursorUtil.getColumnIndexOrThrow(_cursor, "cloud_id");
          final int _cursorIndexOfChapterIndex = CursorUtil.getColumnIndexOrThrow(_cursor, "chapter_index");
          final int _cursorIndexOfTitle = CursorUtil.getColumnIndexOrThrow(_cursor, "title");
          final int _cursorIndexOfMd5 = CursorUtil.getColumnIndexOrThrow(_cursor, "md5");
          final int _cursorIndexOfWordsCount = CursorUtil.getColumnIndexOrThrow(_cursor, "words_count");
          final List<ChapterEntity> _result = new ArrayList<ChapterEntity>(_cursor.getCount());
          while (_cursor.moveToNext()) {
            final ChapterEntity _item;
            final long _tmpId;
            _tmpId = _cursor.getLong(_cursorIndexOfId);
            final long _tmpBookId;
            _tmpBookId = _cursor.getLong(_cursorIndexOfBookId);
            final long _tmpCloudId;
            _tmpCloudId = _cursor.getLong(_cursorIndexOfCloudId);
            final int _tmpChapterIndex;
            _tmpChapterIndex = _cursor.getInt(_cursorIndexOfChapterIndex);
            final String _tmpTitle;
            _tmpTitle = _cursor.getString(_cursorIndexOfTitle);
            final String _tmpMd5;
            _tmpMd5 = _cursor.getString(_cursorIndexOfMd5);
            final int _tmpWordsCount;
            _tmpWordsCount = _cursor.getInt(_cursorIndexOfWordsCount);
            _item = new ChapterEntity(_tmpId,_tmpBookId,_tmpCloudId,_tmpChapterIndex,_tmpTitle,_tmpMd5,_tmpWordsCount);
            _result.add(_item);
          }
          return _result;
        } finally {
          _cursor.close();
          _statement.release();
        }
      }
    }, $completion);
  }

  @Override
  public Object getChapterByCloudId(final long cloudId,
      final Continuation<? super ChapterEntity> $completion) {
    final String _sql = "SELECT * FROM chapters WHERE cloud_id = ? LIMIT 1";
    final RoomSQLiteQuery _statement = RoomSQLiteQuery.acquire(_sql, 1);
    int _argIndex = 1;
    _statement.bindLong(_argIndex, cloudId);
    final CancellationSignal _cancellationSignal = DBUtil.createCancellationSignal();
    return CoroutinesRoom.execute(__db, false, _cancellationSignal, new Callable<ChapterEntity>() {
      @Override
      @Nullable
      public ChapterEntity call() throws Exception {
        final Cursor _cursor = DBUtil.query(__db, _statement, false, null);
        try {
          final int _cursorIndexOfId = CursorUtil.getColumnIndexOrThrow(_cursor, "id");
          final int _cursorIndexOfBookId = CursorUtil.getColumnIndexOrThrow(_cursor, "book_id");
          final int _cursorIndexOfCloudId = CursorUtil.getColumnIndexOrThrow(_cursor, "cloud_id");
          final int _cursorIndexOfChapterIndex = CursorUtil.getColumnIndexOrThrow(_cursor, "chapter_index");
          final int _cursorIndexOfTitle = CursorUtil.getColumnIndexOrThrow(_cursor, "title");
          final int _cursorIndexOfMd5 = CursorUtil.getColumnIndexOrThrow(_cursor, "md5");
          final int _cursorIndexOfWordsCount = CursorUtil.getColumnIndexOrThrow(_cursor, "words_count");
          final ChapterEntity _result;
          if (_cursor.moveToFirst()) {
            final long _tmpId;
            _tmpId = _cursor.getLong(_cursorIndexOfId);
            final long _tmpBookId;
            _tmpBookId = _cursor.getLong(_cursorIndexOfBookId);
            final long _tmpCloudId;
            _tmpCloudId = _cursor.getLong(_cursorIndexOfCloudId);
            final int _tmpChapterIndex;
            _tmpChapterIndex = _cursor.getInt(_cursorIndexOfChapterIndex);
            final String _tmpTitle;
            _tmpTitle = _cursor.getString(_cursorIndexOfTitle);
            final String _tmpMd5;
            _tmpMd5 = _cursor.getString(_cursorIndexOfMd5);
            final int _tmpWordsCount;
            _tmpWordsCount = _cursor.getInt(_cursorIndexOfWordsCount);
            _result = new ChapterEntity(_tmpId,_tmpBookId,_tmpCloudId,_tmpChapterIndex,_tmpTitle,_tmpMd5,_tmpWordsCount);
          } else {
            _result = null;
          }
          return _result;
        } finally {
          _cursor.close();
          _statement.release();
        }
      }
    }, $completion);
  }

  @NonNull
  public static List<Class<?>> getRequiredConverters() {
    return Collections.emptyList();
  }
}
