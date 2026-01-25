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
import androidx.room.util.StringUtil;
import androidx.sqlite.db.SupportSQLiteStatement;
import com.storytrim.app.core.database.entity.ContentEntity;
import java.lang.Class;
import java.lang.Exception;
import java.lang.Object;
import java.lang.Override;
import java.lang.String;
import java.lang.StringBuilder;
import java.lang.SuppressWarnings;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.concurrent.Callable;
import javax.annotation.processing.Generated;
import kotlin.Unit;
import kotlin.coroutines.Continuation;

@Generated("androidx.room.RoomProcessor")
@SuppressWarnings({"unchecked", "deprecation"})
public final class ContentDao_Impl implements ContentDao {
  private final RoomDatabase __db;

  private final EntityInsertionAdapter<ContentEntity> __insertionAdapterOfContentEntity;

  private final SharedSQLiteStatement __preparedStmtOfDeleteContentByMd5;

  public ContentDao_Impl(@NonNull final RoomDatabase __db) {
    this.__db = __db;
    this.__insertionAdapterOfContentEntity = new EntityInsertionAdapter<ContentEntity>(__db) {
      @Override
      @NonNull
      protected String createQuery() {
        return "INSERT OR REPLACE INTO `contents` (`chapter_md5`,`raw_content`) VALUES (?,?)";
      }

      @Override
      protected void bind(@NonNull final SupportSQLiteStatement statement,
          @NonNull final ContentEntity entity) {
        statement.bindString(1, entity.getChapterMd5());
        statement.bindString(2, entity.getRawContent());
      }
    };
    this.__preparedStmtOfDeleteContentByMd5 = new SharedSQLiteStatement(__db) {
      @Override
      @NonNull
      public String createQuery() {
        final String _query = "DELETE FROM contents WHERE chapter_md5 = ?";
        return _query;
      }
    };
  }

  @Override
  public Object insertContent(final ContentEntity content,
      final Continuation<? super Unit> $completion) {
    return CoroutinesRoom.execute(__db, true, new Callable<Unit>() {
      @Override
      @NonNull
      public Unit call() throws Exception {
        __db.beginTransaction();
        try {
          __insertionAdapterOfContentEntity.insert(content);
          __db.setTransactionSuccessful();
          return Unit.INSTANCE;
        } finally {
          __db.endTransaction();
        }
      }
    }, $completion);
  }

  @Override
  public Object insertContents(final List<ContentEntity> contents,
      final Continuation<? super Unit> $completion) {
    return CoroutinesRoom.execute(__db, true, new Callable<Unit>() {
      @Override
      @NonNull
      public Unit call() throws Exception {
        __db.beginTransaction();
        try {
          __insertionAdapterOfContentEntity.insert(contents);
          __db.setTransactionSuccessful();
          return Unit.INSTANCE;
        } finally {
          __db.endTransaction();
        }
      }
    }, $completion);
  }

  @Override
  public Object deleteContentByMd5(final String md5, final Continuation<? super Unit> $completion) {
    return CoroutinesRoom.execute(__db, true, new Callable<Unit>() {
      @Override
      @NonNull
      public Unit call() throws Exception {
        final SupportSQLiteStatement _stmt = __preparedStmtOfDeleteContentByMd5.acquire();
        int _argIndex = 1;
        _stmt.bindString(_argIndex, md5);
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
          __preparedStmtOfDeleteContentByMd5.release(_stmt);
        }
      }
    }, $completion);
  }

  @Override
  public Object getContentByMd5(final String md5,
      final Continuation<? super ContentEntity> $completion) {
    final String _sql = "SELECT * FROM contents WHERE chapter_md5 = ? LIMIT 1";
    final RoomSQLiteQuery _statement = RoomSQLiteQuery.acquire(_sql, 1);
    int _argIndex = 1;
    _statement.bindString(_argIndex, md5);
    final CancellationSignal _cancellationSignal = DBUtil.createCancellationSignal();
    return CoroutinesRoom.execute(__db, false, _cancellationSignal, new Callable<ContentEntity>() {
      @Override
      @Nullable
      public ContentEntity call() throws Exception {
        final Cursor _cursor = DBUtil.query(__db, _statement, false, null);
        try {
          final int _cursorIndexOfChapterMd5 = CursorUtil.getColumnIndexOrThrow(_cursor, "chapter_md5");
          final int _cursorIndexOfRawContent = CursorUtil.getColumnIndexOrThrow(_cursor, "raw_content");
          final ContentEntity _result;
          if (_cursor.moveToFirst()) {
            final String _tmpChapterMd5;
            _tmpChapterMd5 = _cursor.getString(_cursorIndexOfChapterMd5);
            final String _tmpRawContent;
            _tmpRawContent = _cursor.getString(_cursorIndexOfRawContent);
            _result = new ContentEntity(_tmpChapterMd5,_tmpRawContent);
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
  public Object getContentsByMd5s(final List<String> md5s,
      final Continuation<? super List<ContentEntity>> $completion) {
    final StringBuilder _stringBuilder = StringUtil.newStringBuilder();
    _stringBuilder.append("SELECT * FROM contents WHERE chapter_md5 IN (");
    final int _inputSize = md5s.size();
    StringUtil.appendPlaceholders(_stringBuilder, _inputSize);
    _stringBuilder.append(")");
    final String _sql = _stringBuilder.toString();
    final int _argCount = 0 + _inputSize;
    final RoomSQLiteQuery _statement = RoomSQLiteQuery.acquire(_sql, _argCount);
    int _argIndex = 1;
    for (String _item : md5s) {
      _statement.bindString(_argIndex, _item);
      _argIndex++;
    }
    final CancellationSignal _cancellationSignal = DBUtil.createCancellationSignal();
    return CoroutinesRoom.execute(__db, false, _cancellationSignal, new Callable<List<ContentEntity>>() {
      @Override
      @NonNull
      public List<ContentEntity> call() throws Exception {
        final Cursor _cursor = DBUtil.query(__db, _statement, false, null);
        try {
          final int _cursorIndexOfChapterMd5 = CursorUtil.getColumnIndexOrThrow(_cursor, "chapter_md5");
          final int _cursorIndexOfRawContent = CursorUtil.getColumnIndexOrThrow(_cursor, "raw_content");
          final List<ContentEntity> _result = new ArrayList<ContentEntity>(_cursor.getCount());
          while (_cursor.moveToNext()) {
            final ContentEntity _item_1;
            final String _tmpChapterMd5;
            _tmpChapterMd5 = _cursor.getString(_cursorIndexOfChapterMd5);
            final String _tmpRawContent;
            _tmpRawContent = _cursor.getString(_cursorIndexOfRawContent);
            _item_1 = new ContentEntity(_tmpChapterMd5,_tmpRawContent);
            _result.add(_item_1);
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
