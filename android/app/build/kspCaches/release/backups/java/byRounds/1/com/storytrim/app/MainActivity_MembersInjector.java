package com.storytrim.app;

import com.storytrim.app.core.network.AuthInterceptor;
import dagger.MembersInjector;
import dagger.internal.DaggerGenerated;
import dagger.internal.InjectedFieldSignature;
import dagger.internal.QualifierMetadata;
import javax.annotation.processing.Generated;
import javax.inject.Provider;

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
public final class MainActivity_MembersInjector implements MembersInjector<MainActivity> {
  private final Provider<AuthInterceptor> authInterceptorProvider;

  public MainActivity_MembersInjector(Provider<AuthInterceptor> authInterceptorProvider) {
    this.authInterceptorProvider = authInterceptorProvider;
  }

  public static MembersInjector<MainActivity> create(
      Provider<AuthInterceptor> authInterceptorProvider) {
    return new MainActivity_MembersInjector(authInterceptorProvider);
  }

  @Override
  public void injectMembers(MainActivity instance) {
    injectAuthInterceptor(instance, authInterceptorProvider.get());
  }

  @InjectedFieldSignature("com.storytrim.app.MainActivity.authInterceptor")
  public static void injectAuthInterceptor(MainActivity instance, AuthInterceptor authInterceptor) {
    instance.authInterceptor = authInterceptor;
  }
}
