package com.storytrim.app.data.repository;

import android.content.Context;
import com.storytrim.app.core.database.AppDatabase;
import com.storytrim.app.core.database.dao.BookDao;
import com.storytrim.app.core.database.dao.ChapterDao;
import com.storytrim.app.core.database.dao.ContentDao;
import com.storytrim.app.core.network.ApiClient;
import com.storytrim.app.core.network.TrimService;
import com.storytrim.app.core.parser.FileParser;
import com.storytrim.app.core.utils.ZipUtils;
import com.storytrim.app.feature.book.BookService;
import dagger.internal.DaggerGenerated;
import dagger.internal.Factory;
import dagger.internal.QualifierMetadata;
import dagger.internal.ScopeMetadata;
import javax.annotation.processing.Generated;
import javax.inject.Provider;

@ScopeMetadata("javax.inject.Singleton")
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
public final class BookRepository_Factory implements Factory<BookRepository> {
  private final Provider<ApiClient> apiClientProvider;

  private final Provider<BookService> bookServiceProvider;

  private final Provider<TrimService> trimServiceProvider;

  private final Provider<BookDao> bookDaoProvider;

  private final Provider<ChapterDao> chapterDaoProvider;

  private final Provider<ContentDao> contentDaoProvider;

  private final Provider<FileParser> fileParserProvider;

  private final Provider<AppDatabase> appDatabaseProvider;

  private final Provider<ZipUtils> zipUtilsProvider;

  private final Provider<Context> contextProvider;

  public BookRepository_Factory(Provider<ApiClient> apiClientProvider,
      Provider<BookService> bookServiceProvider, Provider<TrimService> trimServiceProvider,
      Provider<BookDao> bookDaoProvider, Provider<ChapterDao> chapterDaoProvider,
      Provider<ContentDao> contentDaoProvider, Provider<FileParser> fileParserProvider,
      Provider<AppDatabase> appDatabaseProvider, Provider<ZipUtils> zipUtilsProvider,
      Provider<Context> contextProvider) {
    this.apiClientProvider = apiClientProvider;
    this.bookServiceProvider = bookServiceProvider;
    this.trimServiceProvider = trimServiceProvider;
    this.bookDaoProvider = bookDaoProvider;
    this.chapterDaoProvider = chapterDaoProvider;
    this.contentDaoProvider = contentDaoProvider;
    this.fileParserProvider = fileParserProvider;
    this.appDatabaseProvider = appDatabaseProvider;
    this.zipUtilsProvider = zipUtilsProvider;
    this.contextProvider = contextProvider;
  }

  @Override
  public BookRepository get() {
    return newInstance(apiClientProvider.get(), bookServiceProvider.get(), trimServiceProvider.get(), bookDaoProvider.get(), chapterDaoProvider.get(), contentDaoProvider.get(), fileParserProvider.get(), appDatabaseProvider.get(), zipUtilsProvider.get(), contextProvider.get());
  }

  public static BookRepository_Factory create(Provider<ApiClient> apiClientProvider,
      Provider<BookService> bookServiceProvider, Provider<TrimService> trimServiceProvider,
      Provider<BookDao> bookDaoProvider, Provider<ChapterDao> chapterDaoProvider,
      Provider<ContentDao> contentDaoProvider, Provider<FileParser> fileParserProvider,
      Provider<AppDatabase> appDatabaseProvider, Provider<ZipUtils> zipUtilsProvider,
      Provider<Context> contextProvider) {
    return new BookRepository_Factory(apiClientProvider, bookServiceProvider, trimServiceProvider, bookDaoProvider, chapterDaoProvider, contentDaoProvider, fileParserProvider, appDatabaseProvider, zipUtilsProvider, contextProvider);
  }

  public static BookRepository newInstance(ApiClient apiClient, BookService bookService,
      TrimService trimService, BookDao bookDao, ChapterDao chapterDao, ContentDao contentDao,
      FileParser fileParser, AppDatabase appDatabase, ZipUtils zipUtils, Context context) {
    return new BookRepository(apiClient, bookService, trimService, bookDao, chapterDao, contentDao, fileParser, appDatabase, zipUtils, context);
  }
}
