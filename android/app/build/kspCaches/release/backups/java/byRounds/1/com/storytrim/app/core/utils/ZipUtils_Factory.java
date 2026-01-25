package com.storytrim.app.core.utils;

import dagger.internal.DaggerGenerated;
import dagger.internal.Factory;
import dagger.internal.QualifierMetadata;
import dagger.internal.ScopeMetadata;
import javax.annotation.processing.Generated;

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
public final class ZipUtils_Factory implements Factory<ZipUtils> {
  @Override
  public ZipUtils get() {
    return newInstance();
  }

  public static ZipUtils_Factory create() {
    return InstanceHolder.INSTANCE;
  }

  public static ZipUtils newInstance() {
    return new ZipUtils();
  }

  private static final class InstanceHolder {
    private static final ZipUtils_Factory INSTANCE = new ZipUtils_Factory();
  }
}
