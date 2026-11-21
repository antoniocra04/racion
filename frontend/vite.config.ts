import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    outDir: '../internal/web/spa/generated',
    emptyOutDir: true,
  },
  server: {
    proxy: { '/api': 'http://localhost:8080' },
  },
});
