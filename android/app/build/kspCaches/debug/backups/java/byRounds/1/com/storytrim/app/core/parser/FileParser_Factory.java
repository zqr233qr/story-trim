package com.storytrim.app.core.parser;

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
public final class FileParser_Factory implements Factory<FileParser> {
  @Override
  public FileParser get() {
    return newInstance();
  }

  public static FileParser_Factory create() {
    return InstanceHolder.INSTANCE;
  }

  public static FileParser newInstance() {
    return new FileParser();
  }

  private static final class InstanceHolder {
    private static final FileParser_Factory INSTANCE = new FileParser_Factory();
  }
}
