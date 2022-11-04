import {defineConfig, LibraryOptions, type PluginOption, splitVendorChunkPlugin} from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { visualizer } from "rollup-plugin-visualizer";
import compress from 'vite-plugin-compression'
import ssr from 'vite-plugin-ssr/plugin'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    svelte({compilerOptions: {hydratable: true}}),
    ssr({prerender: true}),
    visualizer({
      emitFile: true,
      filename: "stats.html"
    }) as PluginOption,
    compress({
      algorithm:'brotliCompress'
    }),
    compress({
      algorithm:'gzip'
    })],
  //mode:'development',
  server: {
    host: '0.0.0.0',
    port: 9085
  },
  build: {
    //minify: false,
  }
})
