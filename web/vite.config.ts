import { defineConfig } from 'vite'
import solidPlugin from 'vite-plugin-solid'

const apiURL = 'http://localhost:8888'

export default defineConfig({
  plugins: [solidPlugin()],
  build: {
    outDir: '../dist',
    target: 'esnext',
  },
  server: {
    proxy: {
      '/v1': {
        target: apiURL,
        changeOrigin: true,
      }
    }
  }
})
