import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    rollupOptions: {
      external: [/^(?:#[a-z]+)$/],  // Match all #minpath, #minurl, etc.
    }
  }
})
