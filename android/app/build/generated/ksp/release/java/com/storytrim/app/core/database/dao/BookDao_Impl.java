package com.storytrim.app.core.database.dao;

import android.database.Cursor;
import android.os.CancellationSignal;
import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.room.CoroutinesRoom;
import androidx.room.EntityDeletionOrUpdateAdapter;
import androidx.room.EntityInsertionAdapter;
import androidx.room.RoomDatabase;
import androidx.room.RoomSQLiteQuery;
import androidx.room.SharedSQLiteStatement;
import androidx.room.util.CursorUtil;
import androidx.room.util.DBUtil;
import androidx.sqlite.db.SupportSQLiteStatement;
import com.storytrim.app.core.database.entity.BookEntity;
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
public final class BookDao_Impl implements BookDao {
  private final RoomDatabase __db;

  private final EntityInsertionAdapter<BookEntity> __insertionAdapterOfBookEntity;

  private final EntityDeletionOrUpdateAdapter<BookEntity> __updateAdapterOfBookEntity;

  private final SharedSQLiteStatement __preparedStmtOfDeleteBookById;

  private final SharedSQLiteStatement __preparedStmtOfDeleteBookByCloudId;

  public BookDao_Impl(@NonNull final RoomDatabase __db) {
    this.__db = __db;
    this.__insertionAdapterOfBookEntity = new EntityInsertionAdapter<BookEntity>(__db) {
      @Override
      @NonNull
      protected String createQuery() {
        return "INSERT OR REPLACE INTO `books` (`id`,`cloud_id`,`user_id`,`book_md5`,`title`,`total_chapters`,`sync_state`,`created_at`) VALUES (nullif(?, 0),?,?,?,?,?,?,?)";
      }

      @Override
      protected void bind(@NonNull final SupportSQLiteStatement statement,
          @NonNull final BookEntity entity) {
        statement.bindLong(1, entity.getId());
        statement.bindLong(2, entity.getCloudId());
        statement.bindLong(3, entity.getUserId());
        statement.bindString(4, entity.getBookMd5());
        statement.bindString(5, entity.getTitle());
        statement.bindLong(6, entity.getTotalChapters());
        statement.bindLong(7, entity.getSyncState());
        statement.bindLong(8, entity.getCreatedAt());
      }
    };
    this.__updateAdapterOfBookEntity = new EntityDeletionOrUpdateAdapter<BookEntity>(__db) {
      @Override
      @NonNull
      protected String createQuery() {
        return "UPDATE OR ABORT `books` SET `id` = ?,`cloud_id` = ?,`user_id` = ?,`book_md5` = ?,`title` = ?,`total_chapters` = ?,`sync_state` = ?,`created_at` = ? WHERE `id` = ?";
      }

      @Override
      protected void bind(@NonNull final SupportSQLiteStatement statement,
          @NonNull final BookEntity entity) {
        statement.bindLong(1, entity.getId());
        statement.bindLong(2, entity.getCloudId());
        statement.bindLong(3, entity.getUserId());
        statement.bindString(4, entity.getBookMd5());
        statement.bindString(5, entity.getTitle());
        statement.bindLong(6, entity.getTotalChapters());
        statement.bindLong(7, entity.getSyncState());
        statement.bindLong(8, entity.getCreatedAt());
        statement.bindLong(9, entity.getId());
      }
    };
    this.__preparedStmtOfDeleteBookById = new SharedSQLiteStatement(__db) {
      @Override
      @NonNull
      public String createQuery() {
        final String _query = "DELETE FROM books WHERE id = ?";
        return _query;
      }
    };
    this.__preparedStmtOfDeleteBookByCloudId = new SharedSQLiteStatement(__db) {
      @Override
      @NonNull
      public String createQuery() {
        final String _query = "DELETE FROM books WHERE cloud_id = ?";
        return _query;
      }
    };
  }

  @Override
  public Object insertBook(final BookEntity book, final Continuation<? super Long> $completion) {
    return CoroutinesRoom.execute(__db, true, new Callable<Long>() {
      @Override
      @NonNull
      public Long call() throws Exception {
        __db.beginTransaction();
        try {
          final Long _result = __insertionAdapterOfBookEntity.insertAndReturnId(book);
          __db.setTransactionSuccessful();
          return _result;
        } finally {
          __db.endTransaction();
        }
      }
    }, $completion);
  }

  @Override
  public Object updateBook(final BookEntity book, final Continuation<? super Unit> $completion) {
    return CoroutinesRoom.execute(__db, true, new Callable<Unit>() {
      @Override
      @NonNull
      public Unit call() throws Exception {
        __db.beginTransaction();
        try {
          __updateAdapterOfBookEntity.handle(book);
          __db.setTransactionSuccessful();
          return Unit.INSTANCE;
        } finally {
          __db.endTransaction();
        }
      }
    }, $completion);
  }

  @Override
  public Object deleteBookById(final long id, final Continuation<? super Unit> $completion) {
    return CoroutinesRoom.execute(__db, true, new Callable<Unit>() {
      @Override
      @NonNull
      public Unit call() throws Exception {
        final SupportSQLiteStatement _stmt = __preparedStmtOfDeleteBookById.acquire();
        int _argIndex = 1;
        _stmt.bindLong(_argIndex, id);
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
          __preparedStmtOfDeleteBookById.release(_stmt);
        }
      }
    }, $completion);
  }

  @Override
  public Object deleteBookByCloudId(final long cloudId,
      final Continuation<? super Unit> $completion) {
    return CoroutinesRoom.execute(__db, true, new Callable<Unit>() {
      @Override
      @NonNull
      public Unit call() throws Exception {
        final SupportSQLiteStatement _stmt = __preparedStmtOfDeleteBookByCloudId.acquire();
        int _argIndex = 1;
        _stmt.bindLong(_argIndex, cloudId);
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
          __preparedStmtOfDeleteBookByCloudId.release(_stmt);
        }
      }
    }, $completion);
  }

  @Override
  public Flow<List<BookEntity>> getAllBooks() {
    final String _sql = "SELECT * FROM books ORDER BY created_at DESC";
    final RoomSQLiteQuery _statement = RoomSQLiteQuery.acquire(_sql, 0);
    return CoroutinesRoom.createFlow(__db, false, new String[] {"books"}, new Callable<List<BookEntity>>() {
      @Override
      @NonNull
      public List<BookEntity> call() throws Exception {
        final Cursor _cursor = DBUtil.query(__db, _statement, false, null);
        try {
          final int _cursorIndexOfId = CursorUtil.getColumnIndexOrThrow(_cursor, "id");
          final int _cursorIndexOfCloudId = CursorUtil.getColumnIndexOrThrow(_cursor, "cloud_id");
          final int _cursorIndexOfUserId = CursorUtil.getColumnIndexOrThrow(_cursor, "user_id");
          final int _cursorIndexOfBookMd5 = CursorUtil.getColumnIndexOrThrow(_cursor, "book_md5");
          final int _cursorIndexOfTitle = CursorUtil.getColumnIndexOrThrow(_cursor, "title");
          final int _cursorIndexOfTotalChapters = CursorUtil.getColumnIndexOrThrow(_cursor, "total_chapters");
          final int _cursorIndexOfSyncState = CursorUtil.getColumnIndexOrThrow(_cursor, "sync_state");
          final int _cursorIndexOfCreatedAt = CursorUtil.getColumnIndexOrThrow(_cursor, "created_at");
          final List<BookEntity> _result = new ArrayList<BookEntity>(_cursor.getCount());
          while (_cursor.moveToNext()) {
            final BookEntity _item;
            final long _tmpId;
            _tmpId = _cursor.getLong(_cursorIndexOfId);
            final long _tmpCloudId;
            _tmpCloudId = _cursor.getLong(_cursorIndexOfCloudId);
            final long _tmpUserId;
            _tmpUserId = _cursor.getLong(_cursorIndexOfUserId);
            final String _tmpBookMd5;
            _tmpBookMd5 = _cursor.getString(_cursorIndexOfBookMd5);
            final String _tmpTitle;
            _tmpTitle = _cursor.getString(_cursorIndexOfTitle);
            final int _tmpTotalChapters;
            _tmpTotalChapters = _cursor.getInt(_cursorIndexOfTotalChapters);
            final int _tmpSyncState;
            _tmpSyncState = _cursor.getInt(_cursorIndexOfSyncState);
            final long _tmpCreatedAt;
            _tmpCreatedAt = _cursor.getLong(_cursorIndexOfCreatedAt);
            _item = new BookEntity(_tmpId,_tmpCloudId,_tmpUserId,_tmpBookMd5,_tmpTitle,_tmpTotalChapters,_tmpSyncState,_tmpCreatedAt);
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
  public Object getBookById(final long id, final Continuation<? super BookEntity> $completion) {
    final String _sql = "SELECT * FROM books WHERE id = ? LIMIT 1";
    final RoomSQLiteQuery _statement = RoomSQLiteQuery.acquire(_sql, 1);
    int _argIndex = 1;
    _statement.bindLong(_argIndex, id);
    final CancellationSignal _cancellationSignal = DBUtil.createCancellationSignal();
    return CoroutinesRoom.execute(__db, false, _cancellationSignal, new Callable<BookEntity>() {
      @Override
      @Nullable
      public BookEntity call() throws Exception {
        final Cursor _cursor = DBUtil.query(__db, _statement, false, null);
        try {
          final int _cursorIndexOfId = CursorUtil.getColumnIndexOrThrow(_cursor, "id");
          final int _cursorIndexOfCloudId = CursorUtil.getColumnIndexOrThrow(_cursor, "cloud_id");
          final int _cursorIndexOfUserId = CursorUtil.getColumnIndexOrThrow(_cursor, "user_id");
          final int _cursorIndexOfBookMd5 = CursorUtil.getColumnIndexOrThrow(_cursor, "book_md5");
          final int _cursorIndexOfTitle = CursorUtil.getColumnIndexOrThrow(_cursor, "title");
          final int _cursorIndexOfTotalChapters = CursorUtil.getColumnIndexOrThrow(_cursor, "total_chapters");
          final int _cursorIndexOfSyncState = CursorUtil.getColumnIndexOrThrow(_cursor, "sync_state");
          final int _cursorIndexOfCreatedAt = CursorUtil.getColumnIndexOrThrow(_cursor, "created_at");
          final BookEntity _result;
          if (_cursor.moveToFirst()) {
            final long _tmpId;
            _tmpId = _cursor.getLong(_cursorIndexOfId);
            final long _tmpCloudId;
            _tmpCloudId = _cursor.getLong(_cursorIndexOfCloudId);
            final long _tmpUserId;
            _tmpUserId = _cursor.getLong(_cursorIndexOfUserId);
            final String _tmpBookMd5;
            _tmpBookMd5 = _cursor.getString(_cursorIndexOfBookMd5);
            final String _tmpTitle;
            _tmpTitle = _cursor.getString(_cursorIndexOfTitle);
            final int _tmpTotalChapters;
            _tmpTotalChapters = _cursor.getInt(_cursorIndexOfTotalChapters);
            final int _tmpSyncState;
            _tmpSyncState = _cursor.getInt(_cursorIndexOfSyncState);
            final long _tmpCreatedAt;
            _tmpCreatedAt = _cursor.getLong(_cursorIndexOfCreatedAt);
            _result = new BookEntity(_tmpId,_tmpCloudId,_tmpUserId,_tmpBookMd5,_tmpTitle,_tmpTotalChapters,_tmpSyncState,_tmpCreatedAt);
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
  public Flow<List<BookEntity>> getBooksByUserId(final long userId) {
    final String _sql = "SELECT * FROM books WHERE user_id = ? ORDER BY created_at DESC";
    final RoomSQLiteQuery _statement = RoomSQLiteQuery.acquire(_sql, 1);
    int _argIndex = 1;
    _statement.bindLong(_argIndex, userId);
    return CoroutinesRoom.createFlow(__db, false, new String[] {"books"}, new Callable<List<BookEntity>>() {
      @Override
      @NonNull
      public List<BookEntity> call() throws Exception {
        final Cursor _cursor = DBUtil.query(__db, _statement, false, null);
        try {
          final int _cursorIndexOfId = CursorUtil.getColumnIndexOrThrow(_cursor, "id");
          final int _cursorIndexOfCloudId = CursorUtil.getColumnIndexOrThrow(_cursor, "cloud_id");
          final int _cursorIndexOfUserId = CursorUtil.getColumnIndexOrThrow(_cursor, "user_id");
          final int _cursorIndexOfBookMd5 = CursorUtil.getColumnIndexOrThrow(_cursor, "book_md5");
          final int _cursorIndexOfTitle = CursorUtil.getColumnIndexOrThrow(_cursor, "title");
          final int _cursorIndexOfTotalChapters = CursorUtil.getColumnIndexOrThrow(_cursor, "total_chapters");
          final int _cursorIndexOfSyncState = CursorUtil.getColumnIndexOrThrow(_cursor, "sync_state");
          final int _cursorIndexOfCreatedAt = CursorUtil.getColumnIndexOrThrow(_cursor, "created_at");
          final List<BookEntity> _result = new ArrayList<BookEntity>(_cursor.getCount());
          while (_cursor.moveToNext()) {
            final BookEntity _item;
            final long _tmpId;
            _tmpId = _cursor.getLong(_cursorIndexOfId);
            final long _tmpCloudId;
            _tmpCloudId = _cursor.getLong(_cursorIndexOfCloudId);
            final long _tmpUserId;
            _tmpUserId = _cursor.getLong(_cursorIndexOfUserId);
            final String _tmpBookMd5;
            _tmpBookMd5 = _cursor.getString(_cursorIndexOfBookMd5);
            final String _tmpTitle;
            _tmpTitle = _cursor.getString(_cursorIndexOfTitle);
            final int _tmpTotalChapters;
            _tmpTotalChapters = _cursor.getInt(_cursorIndexOfTotalChapters);
            final int _tmpSyncState;
            _tmpSyncState = _cursor.getInt(_cursorIndexOfSyncState);
            final long _tmpCreatedAt;
            _tmpCreatedAt = _cursor.getLong(_cursorIndexOfCreatedAt);
            _item = new BookEntity(_tmpId,_tmpCloudId,_tmpUserId,_tmpBookMd5,_tmpTitle,_tmpTotalChapters,_tmpSyncState,_tmpCreatedAt);
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
  public Object getBookByCloudId(final long cloudId,
      final Continuation<? super BookEntity> $completion) {
    final String _sql = "SELECT * FROM books WHERE cloud_id = ? LIMIT 1";
    final RoomSQLiteQuery _statement = RoomSQLiteQuery.acquire(_sql, 1);
    int _argIndex = 1;
    _statement.bindLong(_argIndex, cloudId);
    final CancellationSignal _cancellationSignal = DBUtil.createCancellationSignal();
    return CoroutinesRoom.execute(__db, false, _cancellationSignal, new Callable<BookEntity>() {
      @Override
      @Nullable
      public BookEntity call() throws Exception {
        final Cursor _cursor = DBUtil.query(__db, _statement, false, null);
        try {
          final int _cursorIndexOfId = CursorUtil.getColumnIndexOrThrow(_cursor, "id");
          final int _cursorIndexOfCloudId = CursorUtil.getColumnIndexOrThrow(_cursor, "cloud_id");
          final int _cursorIndexOfUserId = CursorUtil.getColumnIndexOrThrow(_cursor, "user_id");
          final int _cursorIndexOfBookMd5 = CursorUtil.getColumnIndexOrThrow(_cursor, "book_md5");
          final int _cursorIndexOfTitle = CursorUtil.getColumnIndexOrThrow(_cursor, "title");
          final int _cursorIndexOfTotalChapters = CursorUtil.getColumnIndexOrThrow(_cursor, "total_chapters");
          final int _cursorIndexOfSyncState = CursorUtil.getColumnIndexOrThrow(_cursor, "sync_state");
          final int _cursorIndexOfCreatedAt = CursorUtil.getColumnIndexOrThrow(_cursor, "created_at");
          final BookEntity _result;
          if (_cursor.moveToFirst()) {
            final long _tmpId;
            _tmpId = _cursor.getLong(_cursorIndexOfId);
            final long _tmpCloudId;
            _tmpCloudId = _cursor.getLong(_cursorIndexOfCloudId);
            final long _tmpUserId;
            _tmpUserId = _cursor.getLong(_cursorIndexOfUserId);
            final String _tmpBookMd5;
            _tmpBookMd5 = _cursor.getString(_cursorIndexOfBookMd5);
            final String _tmpTitle;
            _tmpTitle = _cursor.getString(_cursorIndexOfTitle);
            final int _tmpTotalChapters;
            _tmpTotalChapters = _cursor.getInt(_cursorIndexOfTotalChapters);
            final int _tmpSyncState;
            _tmpSyncState = _cursor.getInt(_cursorIndexOfSyncState);
            final long _tmpCreatedAt;
            _tmpCreatedAt = _cursor.getLong(_cursorIndexOfCreatedAt);
            _result = new BookEntity(_tmpId,_tmpCloudId,_tmpUserId,_tmpBookMd5,_tmpTitle,_tmpTotalChapters,_tmpSyncState,_tmpCreatedAt);
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
