import { defineConfig } from 'vite'
import uni from '@dcloudio/vite-plugin-uni'
import tailwindcss from 'tailwindcss'
import autoprefixer from 'autoprefixer'
import rem2rpx from 'postcss-rem-to-responsive-pixel'

// 尝试多种导入方式以兼容不同版本的 weapp-tailwindcss
import * as tailwind from 'weapp-tailwindcss/vite'
const UnifiedViteWebpackPlugin = (tailwind as any).UnifiedViteWebpackPlugin || (tailwind as any).default?.UnifiedViteWebpackPlugin

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    uni(),
    // 动态加载插件
    ...(typeof UnifiedViteWebpackPlugin === 'function' ? [UnifiedViteWebpackPlugin({
      rem2rpx: true
    })] : [])
  ],
  css: {
    postcss: {
      plugins: [
        tailwindcss(),
        autoprefixer(),
        rem2rpx({
          rootValue: 32,
          propList: ['*'],
          transformUnit: 'rpx',
        }),
      ],
    },
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})