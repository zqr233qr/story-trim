package com.storytrim.app.utils

object Constants {
    object Api {
        const val BASE_URL = "http://localhost:8080/api/v1/"
        const val CONNECT_TIMEOUT = 30_000L
        const val READ_TIMEOUT = 60_000L
        const val WRITE_TIMEOUT = 60_000L
    }

    object Database {
        const val DATABASE_NAME = "story_trim.db"
        const val DATABASE_VERSION = 1
    }

    object Cache {
        const val MEMORY_CACHE_SIZE = 50
        const val DISK_CACHE_SIZE = 50 * 1024 * 1024L // 50MB
    }

    object SyncState {
        const val LOCAL_ONLY = 0
        const val SYNCED = 1
        const val CLOUD_ONLY = 2
    }
}