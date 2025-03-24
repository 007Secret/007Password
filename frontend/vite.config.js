import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import path from 'path';
import process from 'process';

// 准备环境变量，移除可能的引号
const apiUrl = process.env.VUE_APP_API_URL || '/api';
const apiUrlClean = apiUrl.replace(/["']/g, '');

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },
  define: {
    'process.env': {
      NODE_ENV: JSON.stringify(process.env.NODE_ENV || 'development'),
      VUE_APP_API_URL: JSON.stringify(apiUrlClean)
    }
  },
  server: {
    port: 8081,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '/api')
      }
    }
  }
}); 