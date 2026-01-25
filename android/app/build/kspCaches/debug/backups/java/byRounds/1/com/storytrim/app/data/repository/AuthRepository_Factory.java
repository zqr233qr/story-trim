package com.storytrim.app.data.repository;

import com.storytrim.app.core.network.ApiClient;
import com.storytrim.app.core.network.AuthInterceptor;
import com.storytrim.app.data.remote.AuthService;
import dagger.internal.DaggerGenerated;
import dagger.internal.Factory;
import dagger.internal.QualifierMetadata;
import dagger.internal.ScopeMetadata;
import javax.annotation.processing.Generated;
import javax.inject.Provider;

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
public final class AuthRepository_Factory implements Factory<AuthRepository> {
  private final Provider<AuthService> authServiceProvider;

  private final Provider<AuthInterceptor> authInterceptorProvider;

  private final Provider<ApiClient> apiClientProvider;

  public AuthRepository_Factory(Provider<AuthService> authServiceProvider,
      Provider<AuthInterceptor> authInterceptorProvider, Provider<ApiClient> apiClientProvider) {
    this.authServiceProvider = authServiceProvider;
    this.authInterceptorProvider = authInterceptorProvider;
    this.apiClientProvider = apiClientProvider;
  }

  @Override
  public AuthRepository get() {
    return newInstance(authServiceProvider.get(), authInterceptorProvider.get(), apiClientProvider.get());
  }

  public static AuthRepository_Factory create(Provider<AuthService> authServiceProvider,
      Provider<AuthInterceptor> authInterceptorProvider, Provider<ApiClient> apiClientProvider) {
    return new AuthRepository_Factory(authServiceProvider, authInterceptorProvider, apiClientProvider);
  }

  public static AuthRepository newInstance(AuthService authService, AuthInterceptor authInterceptor,
      ApiClient apiClient) {
    return new AuthRepository(authService, authInterceptor, apiClient);
  }
}
