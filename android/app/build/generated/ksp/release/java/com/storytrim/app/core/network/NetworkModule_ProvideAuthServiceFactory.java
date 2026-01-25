package com.storytrim.app.core.network;

import com.storytrim.app.data.remote.AuthService;
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
public final class NetworkModule_ProvideAuthServiceFactory implements Factory<AuthService> {
  private final Provider<Retrofit> retrofitProvider;

  public NetworkModule_ProvideAuthServiceFactory(Provider<Retrofit> retrofitProvider) {
    this.retrofitProvider = retrofitProvider;
  }

  @Override
  public AuthService get() {
    return provideAuthService(retrofitProvider.get());
  }

  public static NetworkModule_ProvideAuthServiceFactory create(
      Provider<Retrofit> retrofitProvider) {
    return new NetworkModule_ProvideAuthServiceFactory(retrofitProvider);
  }

  public static AuthService provideAuthService(Retrofit retrofit) {
    return Preconditions.checkNotNullFromProvides(NetworkModule.INSTANCE.provideAuthService(retrofit));
  }
}
