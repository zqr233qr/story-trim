package com.storytrim.app.ui.debug;

import com.storytrim.app.core.database.AppDatabase;
import dagger.internal.DaggerGenerated;
import dagger.internal.Factory;
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
public final class DebugSqlViewModel_Factory implements Factory<DebugSqlViewModel> {
  private final Provider<AppDatabase> appDatabaseProvider;

  public DebugSqlViewModel_Factory(Provider<AppDatabase> appDatabaseProvider) {
    this.appDatabaseProvider = appDatabaseProvider;
  }

  @Override
  public DebugSqlViewModel get() {
    return newInstance(appDatabaseProvider.get());
  }

  public static DebugSqlViewModel_Factory create(Provider<AppDatabase> appDatabaseProvider) {
    return new DebugSqlViewModel_Factory(appDatabaseProvider);
  }

  public static DebugSqlViewModel newInstance(AppDatabase appDatabase) {
    return new DebugSqlViewModel(appDatabase);
  }
}
