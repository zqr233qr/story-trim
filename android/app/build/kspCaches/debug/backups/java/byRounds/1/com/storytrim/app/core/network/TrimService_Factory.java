package com.storytrim.app.core.network;

import dagger.internal.DaggerGenerated;
import dagger.internal.Factory;
import dagger.internal.QualifierMetadata;
import dagger.internal.ScopeMetadata;
import javax.annotation.processing.Generated;
import javax.inject.Provider;
import okhttp3.OkHttpClient;

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
public final class TrimService_Factory implements Factory<TrimService> {
  private final Provider<OkHttpClient> okHttpClientProvider;

  private final Provider<AuthInterceptor> authInterceptorProvider;

  public TrimService_Factory(Provider<OkHttpClient> okHttpClientProvider,
      Provider<AuthInterceptor> authInterceptorProvider) {
    this.okHttpClientProvider = okHttpClientProvider;
    this.authInterceptorProvider = authInterceptorProvider;
  }

  @Override
  public TrimService get() {
    return newInstance(okHttpClientProvider.get(), authInterceptorProvider.get());
  }

  public static TrimService_Factory create(Provider<OkHttpClient> okHttpClientProvider,
      Provider<AuthInterceptor> authInterceptorProvider) {
    return new TrimService_Factory(okHttpClientProvider, authInterceptorProvider);
  }

  public static TrimService newInstance(OkHttpClient okHttpClient,
      AuthInterceptor authInterceptor) {
    return new TrimService(okHttpClient, authInterceptor);
  }
}
