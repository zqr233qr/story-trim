package com.storytrim.app.core.database.dao;

import android.database.Cursor;
import android.os.CancellationSignal;
import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.room.CoroutinesRoom;
import androidx.room.EntityInsertionAdapter;
import androidx.room.RoomDatabase;
import androidx.room.RoomSQLiteQuery;
import androidx.room.util.CursorUtil;
import androidx.room.util.DBUtil;
import androidx.sqlite.db.SupportSQLiteStatement;
import com.storytrim.app.core.database.entity.TrimmedContentEntity;
import java.lang.Class;
import java.lang.Exception;
import java.lang.Object;
import java.lang.Override;
import java.lang.String;
import java.lang.SuppressWarnings;
import java.util.Collections;
import java.util.List;
import java.util.concurrent.Callable;
import javax.annotation.processing.Generated;
import kotlin.Unit;
import kotlin.coroutines.Continuation;

@Generated("androidx.room.RoomProcessor")
@SuppressWarnings({"unchecked", "deprecation"})
public final class TrimmedContentDao_Impl implements TrimmedContentDao {
  private final RoomDatabase __db;

  private final EntityInsertionAdapter<TrimmedContentEntity> __insertionAdapterOfTrimmedContentEntity;

  public TrimmedContentDao_Impl(@NonNull final RoomDatabase __db) {
    this.__db = __db;
    this.__insertionAdapterOfTrimmedContentEntity = new EntityInsertionAdapter<TrimmedContentEntity>(__db) {
      @Override
      @NonNull
      protected String createQuery() {
        return "INSERT OR REPLACE INTO `trimmed_contents` (`book_id`,`chapter_id`,`prompt_id`,`content`,`created_at`) VALUES (?,?,?,?,?)";
      }

      @Override
      protected void bind(@NonNull final SupportSQLiteStatement statement,
          @NonNull final TrimmedContentEntity entity) {
        statement.bindLong(1, entity.getBookId());
        statement.bindLong(2, entity.getChapterId());
        statement.bindLong(3, entity.getPromptId());
        statement.bindString(4, entity.getContent());
        statement.bindLong(5, entity.getCreatedAt());
      }
    };
  }

  @Override
  public Object insertTrimmedContent(final TrimmedContentEntity entity,
      final Continuation<? super Unit> $completion) {
    return CoroutinesRoom.execute(__db, true, new Callable<Unit>() {
      @Override
      @NonNull
      public Unit call() throws Exception {
        __db.beginTransaction();
        try {
          __insertionAdapterOfTrimmedContentEntity.insert(entity);
          __db.setTransactionSuccessful();
          return Unit.INSTANCE;
        } finally {
          __db.endTransaction();
        }
      }
    }, $completion);
  }

  @Override
  public Object getTrimmedContent(final long bookId, final long chapterId, final int promptId,
      final Continuation<? super TrimmedContentEntity> $completion) {
    final String _sql = "SELECT * FROM trimmed_contents WHERE book_id = ? AND chapter_id = ? AND prompt_id = ?";
    final RoomSQLiteQuery _statement = RoomSQLiteQuery.acquire(_sql, 3);
    int _argIndex = 1;
    _statement.bindLong(_argIndex, bookId);
    _argIndex = 2;
    _statement.bindLong(_argIndex, chapterId);
    _argIndex = 3;
    _statement.bindLong(_argIndex, promptId);
    final CancellationSignal _cancellationSignal = DBUtil.createCancellationSignal();
    return CoroutinesRoom.execute(__db, false, _cancellationSignal, new Callable<TrimmedContentEntity>() {
      @Override
      @Nullable
      public TrimmedContentEntity call() throws Exception {
        final Cursor _cursor = DBUtil.query(__db, _statement, false, null);
        try {
          final int _cursorIndexOfBookId = CursorUtil.getColumnIndexOrThrow(_cursor, "book_id");
          final int _cursorIndexOfChapterId = CursorUtil.getColumnIndexOrThrow(_cursor, "chapter_id");
          final int _cursorIndexOfPromptId = CursorUtil.getColumnIndexOrThrow(_cursor, "prompt_id");
          final int _cursorIndexOfContent = CursorUtil.getColumnIndexOrThrow(_cursor, "content");
          final int _cursorIndexOfCreatedAt = CursorUtil.getColumnIndexOrThrow(_cursor, "created_at");
          final TrimmedContentEntity _result;
          if (_cursor.moveToFirst()) {
            final long _tmpBookId;
            _tmpBookId = _cursor.getLong(_cursorIndexOfBookId);
            final long _tmpChapterId;
            _tmpChapterId = _cursor.getLong(_cursorIndexOfChapterId);
            final int _tmpPromptId;
            _tmpPromptId = _cursor.getInt(_cursorIndexOfPromptId);
            final String _tmpContent;
            _tmpContent = _cursor.getString(_cursorIndexOfContent);
            final long _tmpCreatedAt;
            _tmpCreatedAt = _cursor.getLong(_cursorIndexOfCreatedAt);
            _result = new TrimmedContentEntity(_tmpBookId,_tmpChapterId,_tmpPromptId,_tmpContent,_tmpCreatedAt);
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
