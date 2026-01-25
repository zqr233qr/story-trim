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
import com.storytrim.app.core.database.entity.ReadingHistoryEntity;
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
public final class ReadingHistoryDao_Impl implements ReadingHistoryDao {
  private final RoomDatabase __db;

  private final EntityInsertionAdapter<ReadingHistoryEntity> __insertionAdapterOfReadingHistoryEntity;

  public ReadingHistoryDao_Impl(@NonNull final RoomDatabase __db) {
    this.__db = __db;
    this.__insertionAdapterOfReadingHistoryEntity = new EntityInsertionAdapter<ReadingHistoryEntity>(__db) {
      @Override
      @NonNull
      protected String createQuery() {
        return "INSERT OR REPLACE INTO `reading_history` (`book_id`,`last_chapter_id`,`last_prompt_id`,`updated_at`) VALUES (?,?,?,?)";
      }

      @Override
      protected void bind(@NonNull final SupportSQLiteStatement statement,
          @NonNull final ReadingHistoryEntity entity) {
        statement.bindLong(1, entity.getBookId());
        statement.bindLong(2, entity.getLastChapterId());
        statement.bindLong(3, entity.getLastPromptId());
        statement.bindLong(4, entity.getUpdatedAt());
      }
    };
  }

  @Override
  public Object insertReadingHistory(final ReadingHistoryEntity history,
      final Continuation<? super Unit> $completion) {
    return CoroutinesRoom.execute(__db, true, new Callable<Unit>() {
      @Override
      @NonNull
      public Unit call() throws Exception {
        __db.beginTransaction();
        try {
          __insertionAdapterOfReadingHistoryEntity.insert(history);
          __db.setTransactionSuccessful();
          return Unit.INSTANCE;
        } finally {
          __db.endTransaction();
        }
      }
    }, $completion);
  }

  @Override
  public Object getReadingHistory(final long bookId,
      final Continuation<? super ReadingHistoryEntity> $completion) {
    final String _sql = "SELECT * FROM reading_history WHERE book_id = ?";
    final RoomSQLiteQuery _statement = RoomSQLiteQuery.acquire(_sql, 1);
    int _argIndex = 1;
    _statement.bindLong(_argIndex, bookId);
    final CancellationSignal _cancellationSignal = DBUtil.createCancellationSignal();
    return CoroutinesRoom.execute(__db, false, _cancellationSignal, new Callable<ReadingHistoryEntity>() {
      @Override
      @Nullable
      public ReadingHistoryEntity call() throws Exception {
        final Cursor _cursor = DBUtil.query(__db, _statement, false, null);
        try {
          final int _cursorIndexOfBookId = CursorUtil.getColumnIndexOrThrow(_cursor, "book_id");
          final int _cursorIndexOfLastChapterId = CursorUtil.getColumnIndexOrThrow(_cursor, "last_chapter_id");
          final int _cursorIndexOfLastPromptId = CursorUtil.getColumnIndexOrThrow(_cursor, "last_prompt_id");
          final int _cursorIndexOfUpdatedAt = CursorUtil.getColumnIndexOrThrow(_cursor, "updated_at");
          final ReadingHistoryEntity _result;
          if (_cursor.moveToFirst()) {
            final long _tmpBookId;
            _tmpBookId = _cursor.getLong(_cursorIndexOfBookId);
            final long _tmpLastChapterId;
            _tmpLastChapterId = _cursor.getLong(_cursorIndexOfLastChapterId);
            final int _tmpLastPromptId;
            _tmpLastPromptId = _cursor.getInt(_cursorIndexOfLastPromptId);
            final long _tmpUpdatedAt;
            _tmpUpdatedAt = _cursor.getLong(_cursorIndexOfUpdatedAt);
            _result = new ReadingHistoryEntity(_tmpBookId,_tmpLastChapterId,_tmpLastPromptId,_tmpUpdatedAt);
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
