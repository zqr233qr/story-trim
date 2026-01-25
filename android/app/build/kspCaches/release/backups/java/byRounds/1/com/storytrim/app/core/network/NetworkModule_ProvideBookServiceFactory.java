package com.storytrim.app.core.network;

import com.storytrim.app.feature.book.BookService;
import dagger.internal.DaggerGenerated;
import dagger.internal.Factory;
import dagger.internal.Preconditions;
import dagger.internal.QualifierMetadata;
import dagger.internal.ScopeMetadata;
import javax.annotation.processing.Generated;
import javax.inject.Provider;
import retrofit2.Retrofit;

@ScopeMetadata("javax.inject.Singleton")
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
public final class NetworkModule_ProvideBookServiceFactory implements Factory<BookService> {
  private final Provider<Retrofit> retrofitProvider;

  public NetworkModule_ProvideBookServiceFactory(Provider<Retrofit> retrofitProvider) {
    this.retrofitProvider = retrofitProvider;
  }

  @Override
  public BookService get() {
    return provideBookService(retrofitProvider.get());
  }

  public static NetworkModule_ProvideBookServiceFactory create(
      Provider<Retrofit> retrofitProvider) {
    return new NetworkModule_ProvideBookServiceFactory(retrofitProvider);
  }

  public static BookService provideBookService(Retrofit retrofit) {
    return Preconditions.checkNotNullFromProvides(NetworkModule.INSTANCE.provideBookService(retrofit));
  }
}
