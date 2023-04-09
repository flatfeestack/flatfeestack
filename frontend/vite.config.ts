import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";
import { visualizer } from "rollup-plugin-visualizer";
import compress from "vite-plugin-compression";
const mode = process.env.MODE || "dev";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    sveltekit(),

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
    }),
  ],

  mode: mode === "dev" ? "development" : "production",
  server: {
    port: 9085,
  },

  build: {
    minify: mode !== "dev",
    emptyOutDir: true,
  },
});

/*
import { svelte } from "@sveltejs/vite-plugin-svelte";
import ssr from "vite-plugin-ssr/plugin";

export default defineConfig({
  plugins: [
    svelte({ compilerOptions: { hydratable: true } }),
    mode !== "hmr" ? ssr({ prerender: true }) : {},
  ],
});
*/
