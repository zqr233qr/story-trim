package com.storytrim.app.core.network;

import dagger.internal.DaggerGenerated;
import dagger.internal.Factory;
import dagger.internal.QualifierMetadata;
import dagger.internal.ScopeMetadata;
import javax.annotation.processing.Generated;

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
public final class ApiClient_Factory implements Factory<ApiClient> {
  @Override
  public ApiClient get() {
    return newInstance();
  }

  public static ApiClient_Factory create() {
    return InstanceHolder.INSTANCE;
  }

  public static ApiClient newInstance() {
    return new ApiClient();
  }

  private static final class InstanceHolder {
    private static final ApiClient_Factory INSTANCE = new ApiClient_Factory();
  }
}
