package com.storytrim.app.core.network;

import com.google.gson.Gson;
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
public final class ApiErrorHandler_Factory implements Factory<ApiErrorHandler> {
  private final Provider<Gson> gsonProvider;

  public ApiErrorHandler_Factory(Provider<Gson> gsonProvider) {
    this.gsonProvider = gsonProvider;
  }

  @Override
  public ApiErrorHandler get() {
    return newInstance(gsonProvider.get());
  }

  public static ApiErrorHandler_Factory create(Provider<Gson> gsonProvider) {
    return new ApiErrorHandler_Factory(gsonProvider);
  }

  public static ApiErrorHandler newInstance(Gson gson) {
    return new ApiErrorHandler(gson);
  }
}
