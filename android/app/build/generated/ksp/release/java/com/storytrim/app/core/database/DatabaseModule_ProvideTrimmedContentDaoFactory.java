package com.storytrim.app.core.database;

import com.storytrim.app.core.database.dao.TrimmedContentDao;
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
public final class DatabaseModule_ProvideTrimmedContentDaoFactory implements Factory<TrimmedContentDao> {
  private final Provider<AppDatabase> databaseProvider;

  public DatabaseModule_ProvideTrimmedContentDaoFactory(Provider<AppDatabase> databaseProvider) {
    this.databaseProvider = databaseProvider;
  }

  @Override
  public TrimmedContentDao get() {
    return provideTrimmedContentDao(databaseProvider.get());
  }

  public static DatabaseModule_ProvideTrimmedContentDaoFactory create(
      Provider<AppDatabase> databaseProvider) {
    return new DatabaseModule_ProvideTrimmedContentDaoFactory(databaseProvider);
  }

  public static TrimmedContentDao provideTrimmedContentDao(AppDatabase database) {
    return Preconditions.checkNotNullFromProvides(DatabaseModule.INSTANCE.provideTrimmedContentDao(database));
  }
}
