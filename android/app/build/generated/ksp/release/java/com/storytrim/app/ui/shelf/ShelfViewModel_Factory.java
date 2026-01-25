package com.storytrim.app.ui.shelf;

import android.content.Context;
import com.storytrim.app.data.repository.AuthRepository;
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
public final class ShelfViewModel_Factory implements Factory<ShelfViewModel> {
  private final Provider<BookRepository> bookRepositoryProvider;

  private final Provider<AuthRepository> authRepositoryProvider;

  private final Provider<Context> contextProvider;

  public ShelfViewModel_Factory(Provider<BookRepository> bookRepositoryProvider,
      Provider<AuthRepository> authRepositoryProvider, Provider<Context> contextProvider) {
    this.bookRepositoryProvider = bookRepositoryProvider;
    this.authRepositoryProvider = authRepositoryProvider;
    this.contextProvider = contextProvider;
  }

  @Override
  public ShelfViewModel get() {
    return newInstance(bookRepositoryProvider.get(), authRepositoryProvider.get(), contextProvider.get());
  }

  public static ShelfViewModel_Factory create(Provider<BookRepository> bookRepositoryProvider,
      Provider<AuthRepository> authRepositoryProvider, Provider<Context> contextProvider) {
    return new ShelfViewModel_Factory(bookRepositoryProvider, authRepositoryProvider, contextProvider);
  }

  public static ShelfViewModel newInstance(BookRepository bookRepository,
      AuthRepository authRepository, Context context) {
    return new ShelfViewModel(bookRepository, authRepository, context);
  }
}
