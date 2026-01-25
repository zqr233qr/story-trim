package com.storytrim.app.core.utils

import java.io.BufferedInputStream
import java.io.File
import java.io.FileInputStream
import java.io.FileOutputStream
import java.util.zip.ZipEntry
import java.util.zip.ZipInputStream
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ZipUtils @Inject constructor() {

    fun unzip(zipFile: File, targetDir: File) {
        if (!targetDir.exists()) {
            targetDir.mkdirs()
        }

        ZipInputStream(BufferedInputStream(FileInputStream(zipFile))).use { zis ->
            var entry: ZipEntry?
            while (zis.nextEntry.also { entry = it } != null) {
                val file = File(targetDir, entry!!.name)
                
                // 防止 Zip Slip 漏洞
                if (!file.canonicalPath.startsWith(targetDir.canonicalPath + File.separator)) {
                    throw SecurityException("Zip entry is outside of the target dir: ${entry!!.name}")
                }

                if (entry!!.isDirectory) {
                    file.mkdirs()
                } else {
                    file.parentFile?.mkdirs()
                    FileOutputStream(file).use { fos ->
                        val buffer = ByteArray(8192)
                        var len: Int
                        while (zis.read(buffer).also { len = it } > 0) {
                            fos.write(buffer, 0, len)
                        }
                    }
                }
            }
        }
    }
}
