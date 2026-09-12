import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// Build output is embedded by the Go binary via //go:embed dist (see ../embed.go).
export default defineConfig({
  plugins: [react()],
  base: '/',
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  server: {
    port: 5173,
    proxy: {
      '/auth': 'http://localhost:8080',
      '/api': 'http://localhost:8080',
      '/health': 'http://localhost:8080',
      '/metrics': 'http://localhost:8080',
    },
  },
});
