import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { NaiveUiResolver } from 'unplugin-vue-components/resolvers'
import Icons from 'unplugin-icons/vite'
import IconsResolver from 'unplugin-icons/resolver'

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      imports: [
        'vue',
        'vue-router',
        {
          'naive-ui': [
            'useDialog',
            'useMessage',
            'useNotification',
            'useLoadingBar'
          ]
        }
      ],
      dts: './auto-imports.d.ts',
    }),
    Components({
      resolvers: [
        NaiveUiResolver(),
        IconsResolver({
          prefix: 'i',
          enabledCollections: ['ionicons']
        })
      ],
      dts: './components.d.ts',
    }),
    Icons({
      compiler: 'vue3',
      autoInstall: true
    })
  ],
  
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@/components': resolve(__dirname, 'src/components'),
      '@/views': resolve(__dirname, 'src/views'),
      '@/composables': resolve(__dirname, 'src/composables'),
      '@/styles': resolve(__dirname, 'src/styles'),
      '@/types': resolve(__dirname, 'src/types'),
      '@/wailsjs': resolve(__dirname, 'wailsjs')
    }
  },
  
  // 构建优化配置
  build: {
    target: 'esnext',
    minify: 'terser',
    cssCodeSplit: true,
    sourcemap: false,
    
    // Rollup配置
    rollupOptions: {
      output: {
        // 手动代码分割策略
        manualChunks(id) {
          // 第三方库分组
          if (id.includes('node_modules')) {
            // Vue生态系统
            if (id.includes('vue')) {
              return 'vue-vendor'
            }
            // Naive UI
            if (id.includes('naive-ui')) {
              return 'naive-ui'
            }
            // 其他工具库
            return 'vendor'
          }
          
          // 业务代码分组
          if (id.includes('/src/composables/')) {
            return 'composables'
          }
          
          if (id.includes('/src/components/business/')) {
            return 'business-components'
          }
          
          if (id.includes('/src/components/layout/')) {
            return 'layout-components'
          }
          
          if (id.includes('/src/views/Settings')) {
            return 'settings'
          }
        },
        
        // 优化文件命名
        chunkFileNames: 'js/[name]-[hash].js',
        entryFileNames: 'js/[name]-[hash].js',
        assetFileNames: (assetInfo) => {
          const info = assetInfo.name!.split('.')
          const ext = info[info.length - 1]
          if (/png|jpe?g|svg|gif|tiff|bmp|ico/i.test(ext)) {
            return `img/[name]-[hash].[ext]`
          }
          if (ext === 'css') {
            return `css/[name]-[hash].[ext]`
          }
          if (/woff|woff2|eot|ttf|otf/i.test(ext)) {
            return `fonts/[name]-[hash].[ext]`
          }
          return `assets/[name]-[hash].[ext]`
        }
      }
    },
    
    // Terser压缩配置
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
        pure_funcs: ['console.log', 'console.info']
      },
      mangle: {
        safari10: true
      },
      format: {
        comments: false
      }
    },
    
    // 性能优化
    reportCompressedSize: false,
    chunkSizeWarningLimit: 1000,
    
    // 资源内联阈值（4KB以下的资源会被内联）
    assetsInlineLimit: 4096
  },
  
  // CSS配置
  css: {
    preprocessorOptions: {
      css: {
        charset: false
      }
    },
    postcss: {
      plugins: []
    }
  },
  
  // 开发服务器配置
  server: {
    host: true,
    port: 5173,
    strictPort: true,
    hmr: {
      overlay: true
    }
  },
  
  // 依赖优化
  optimizeDeps: {
    include: [
      'vue',
      'vue-router',
      'naive-ui'
    ],
    exclude: []
  }
})