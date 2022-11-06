import {defineConfig} from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import compress from 'vite-plugin-compression'
import ssr from 'vite-plugin-ssr/plugin'
const mode = process.env.MODE || 'dev';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    //svelte(),
    svelte({compilerOptions: {hydratable: true}}),
    ssr({prerender: true}),

    compress({
      algorithm:'brotliCompress',
      disable: mode === 'dev'
    }),
    compress({
      algorithm:'gzip',
      disable: mode === 'dev'
    })],
  mode: mode === 'dev' ? 'development': 'production',
  build: {
    minify: mode !== 'dev',
    emptyOutDir: true,
  }
})
