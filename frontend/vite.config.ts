import {defineConfig, LibraryOptions, type PluginOption, splitVendorChunkPlugin} from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { visualizer } from "rollup-plugin-visualizer";
import compress from 'vite-plugin-compression'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    svelte(),
    splitVendorChunkPlugin(),
    visualizer({
      emitFile: true,
      filename: "stats.html"
    }) as PluginOption,
    compress({
      algorithm:'brotliCompress'
    }),
    compress({
      algorithm:'gzip'
    })]
})
