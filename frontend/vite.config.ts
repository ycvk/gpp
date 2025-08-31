import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { NaiveUiResolver } from 'unplugin-vue-components/resolvers'
import Icons from 'unplugin-icons/vite'
import IconsResolver from 'unplugin-icons/resolver'

export default defineConfig({
  base: "./",  // 使用相对路径
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
        // 简化输出路径
        entryFileNames: 'assets/[name].js',
        chunkFileNames: 'assets/[name].js',
        assetFileNames: 'assets/[name].[ext]'
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