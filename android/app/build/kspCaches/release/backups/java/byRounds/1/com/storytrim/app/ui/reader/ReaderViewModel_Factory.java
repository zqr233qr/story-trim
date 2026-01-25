package com.storytrim.app.ui.reader;

import com.storytrim.app.data.repository.BookRepository;
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
public final class ReaderViewModel_Factory implements Factory<ReaderViewModel> {
  private final Provider<BookRepository> bookRepositoryProvider;

  public ReaderViewModel_Factory(Provider<BookRepository> bookRepositoryProvider) {
    this.bookRepositoryProvider = bookRepositoryProvider;
  }

  @Override
  public ReaderViewModel get() {
    return newInstance(bookRepositoryProvider.get());
  }

  public static ReaderViewModel_Factory create(Provider<BookRepository> bookRepositoryProvider) {
    return new ReaderViewModel_Factory(bookRepositoryProvider);
  }

  public static ReaderViewModel newInstance(BookRepository bookRepository) {
    return new ReaderViewModel(bookRepository);
  }
}
