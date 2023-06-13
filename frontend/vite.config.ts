import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import compress from "vite-plugin-compression";
import ssr from "vite-plugin-ssr/plugin";
import { visualizer } from "rollup-plugin-visualizer";
const mode = process.env.MODE || "dev";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    svelte({ compilerOptions: { hydratable: true } }),

    mode !== "hmr" ? ssr({ prerender: true }) : {},

    compress({
      algorithm: "brotliCompress",
      disable: mode === "dev",
    }),
    compress({
      algorithm: "gzip",
      disable: mode === "dev",
    }),
    visualizer({
      emitFile: true,
      filename: "stats.html",
      include: { file: "**/node_modules/**" },
    }),
  ],
  server: {
    host: '0.0.0.0',
    port: 9085,
  },
  mode: mode === "dev" ? "development" : "production",
  build: {
    minify: mode !== "dev",
    emptyOutDir: true,
  },
});
