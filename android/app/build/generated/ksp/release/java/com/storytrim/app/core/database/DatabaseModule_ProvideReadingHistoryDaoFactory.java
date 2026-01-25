package com.storytrim.app.core.database;

import com.storytrim.app.core.database.dao.ReadingHistoryDao;
import dagger.internal.DaggerGenerated;
import dagger.internal.Factory;
import dagger.internal.Preconditions;
import dagger.internal.QualifierMetadata;
import dagger.internal.ScopeMetadata;
import javax.annotation.processing.Generated;
import javax.inject.Provider;

@ScopeMetadata
@QualifierMetadata
@DaggerGenerated
@Generated(
    value = "dagger.internal.codegen.ComponentProcessor",
    comments = "https://dagger.dev"
)
@SuppressWarnings({
    "unchecked",
    "rawtypes",
    "KotlinInternal",
    "KotlinInternalInJava",
    "cast"
})
public final class DatabaseModule_ProvideReadingHistoryDaoFactory implements Factory<ReadingHistoryDao> {
  private final Provider<AppDatabase> databaseProvider;

  public DatabaseModule_ProvideReadingHistoryDaoFactory(Provider<AppDatabase> databaseProvider) {
    this.databaseProvider = databaseProvider;
  }

  @Override
  public ReadingHistoryDao get() {
    return provideReadingHistoryDao(databaseProvider.get());
  }

  public static DatabaseModule_ProvideReadingHistoryDaoFactory create(
      Provider<AppDatabase> databaseProvider) {
    return new DatabaseModule_ProvideReadingHistoryDaoFactory(databaseProvider);
  }

  public static ReadingHistoryDao provideReadingHistoryDao(AppDatabase database) {
    return Preconditions.checkNotNullFromProvides(DatabaseModule.INSTANCE.provideReadingHistoryDao(database));
  }
}
