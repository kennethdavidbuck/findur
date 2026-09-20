import { defineConfig } from 'vitest/config'
import { loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, '.', '')
  if (env.VITE_REQUIRE_BUILD_SHA === 'true' && !/^[0-9a-f]{40}$/.test(env.VITE_BUILD_SHA || '')) {
    throw new Error('VITE_BUILD_SHA must be a full lowercase Git revision for production builds')
  }
  return ({
  plugins: [react()],
  server: {
    proxy: {
      '/api': env.VITE_API_PROXY || 'http://localhost:10000',
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: './src/test-setup.ts',
    restoreMocks: true,
  },
  })
})
