import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';
import { resolve } from 'path';

export default defineConfig({
  plugins: [
    tailwindcss(),
  ],
  build: {
    outDir: 'public', // Output build assets directly to the public folder
    emptyOutDir: false, // Don't clear the public folder, as it contains other static assets
    rollupOptions: {
      input: {
        'css/tailwind': resolve(__dirname, 'public/css/src.css'),
      },
      output: {
        entryFileNames: 'js/[name].js', // In case we have JS entries
        assetFileNames: (assetInfo) => {
          if (assetInfo.name === 'src.css') {
            return 'css/tailwind.css'; // Output the built CSS as tailwind.css
          }
          return 'assets/[name]-[hash][extname]';
        }
      }
    }
  }
});
