package com.storytrim.app.core.database;

import com.storytrim.app.core.database.dao.ContentDao;
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
public final class DatabaseModule_ProvideContentDaoFactory implements Factory<ContentDao> {
  private final Provider<AppDatabase> databaseProvider;

  public DatabaseModule_ProvideContentDaoFactory(Provider<AppDatabase> databaseProvider) {
    this.databaseProvider = databaseProvider;
  }

  @Override
  public ContentDao get() {
    return provideContentDao(databaseProvider.get());
  }

  public static DatabaseModule_ProvideContentDaoFactory create(
      Provider<AppDatabase> databaseProvider) {
    return new DatabaseModule_ProvideContentDaoFactory(databaseProvider);
  }

  public static ContentDao provideContentDao(AppDatabase database) {
    return Preconditions.checkNotNullFromProvides(DatabaseModule.INSTANCE.provideContentDao(database));
  }
}
