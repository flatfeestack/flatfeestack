import svelte from "rollup-plugin-svelte";
import resolve from "@rollup/plugin-node-resolve";
import sveltePreprocess from "svelte-preprocess";
import typescript from "rollup-plugin-typescript2";
import css from "rollup-plugin-css-only";
import commonjs from "@rollup/plugin-commonjs";
import serve from "rollup-plugin-serve";

export default [{
  input: "./src/main.ts",
  output: {
    format: "esm",
    dir: "public/build",
    sourcemap: true,
    manualChunks: {
      deps: ["ethers", "ky", "svelte-routing", "@stripe/stripe-js"]
    }
  },
  plugins: [
    svelte({ emitCss: true, preprocess: sveltePreprocess(), compilerOptions: { immutable: true, hydratable: true } }),
    resolve({ browser: true, dedupe: ["svelte"], extensions: [".ts", ".js"] }),
    typescript({ sourceMap: true }),
    css({ output: "bundle.css" }),
    commonjs(),
  ]
},
  {
    input: "./src/App.svelte",
    output: {
      name: "app",
      format: "umd",
      file: "public/server/ssr.js",
      sourcemap: true
    },
    plugins: [
      svelte({ emitCss: true, preprocess: sveltePreprocess(), compilerOptions: { immutable: true, generate: "ssr" } }),
      resolve({ preferBuiltins: true, dedupe: ["svelte"], extensions: [".ts", ".js"] }),
      typescript({ sourceMap: true }),
      css({ output: "ssr.css" }),
      commonjs()
    ]
  }
];
