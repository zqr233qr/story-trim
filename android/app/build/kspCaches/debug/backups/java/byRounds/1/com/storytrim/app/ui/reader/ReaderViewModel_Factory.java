package com.storytrim.app.ui.reader;

import android.content.Context;
import com.storytrim.app.data.repository.BookRepository;
import dagger.internal.DaggerGenerated;
import dagger.internal.Factory;
import dagger.internal.QualifierMetadata;
import dagger.internal.ScopeMetadata;
import javax.annotation.processing.Generated;
import javax.inject.Provider;

@ScopeMetadata
@QualifierMetadata("dagger.hilt.android.qualifiers.ApplicationContext")
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
public final class ReaderViewModel_Factory implements Factory<ReaderViewModel> {
  private final Provider<BookRepository> bookRepositoryProvider;

  private final Provider<Context> appContextProvider;

  public ReaderViewModel_Factory(Provider<BookRepository> bookRepositoryProvider,
      Provider<Context> appContextProvider) {
    this.bookRepositoryProvider = bookRepositoryProvider;
    this.appContextProvider = appContextProvider;
  }

  @Override
  public ReaderViewModel get() {
    return newInstance(bookRepositoryProvider.get(), appContextProvider.get());
  }

  public static ReaderViewModel_Factory create(Provider<BookRepository> bookRepositoryProvider,
      Provider<Context> appContextProvider) {
    return new ReaderViewModel_Factory(bookRepositoryProvider, appContextProvider);
  }

  public static ReaderViewModel newInstance(BookRepository bookRepository, Context appContext) {
    return new ReaderViewModel(bookRepository, appContext);
  }
}
