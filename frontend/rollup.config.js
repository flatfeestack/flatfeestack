import svelte from "rollup-plugin-svelte";
import resolve from "@rollup/plugin-node-resolve";
import sveltePreprocess from "svelte-preprocess";
import typescript from "rollup-plugin-typescript2";
import css from "rollup-plugin-css-only";
import commonjs from "@rollup/plugin-commonjs";
import execute from "rollup-plugin-execute";

import serve from "rollup-plugin-serve";

export default [
    {
        input: "./src/main.ts",
        output: {
            format: "esm",
            dir: "public/build",
            sourcemap: true,
            manualChunks: {
                deps: ["ethers", "ky", "svelte-routing", "@stripe/stripe-js", "canvas-confetti"]
            }
        },
        plugins: [
            svelte({emitCss: true, preprocess: sveltePreprocess(), compilerOptions: {hydratable: true}}),
            resolve({browser: true, dedupe: ["svelte"], extensions: [".ts", ".js"]}),
            typescript({sourceMap: true}),
            css({output: "bundle.css"}),
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
            svelte({emitCss: false, preprocess: sveltePreprocess(), compilerOptions: {generate: "ssr"}}),
            resolve({preferBuiltins: true, dedupe: ["svelte"], extensions: [".ts", ".js"]}),
            typescript({sourceMap: true}),
            commonjs(),
            serve({contentBase: "public", port: 9085, host: "0.0.0.0", historyApiFallback: true})
        ]
    },
    {
        input: "./public/server/ssr.js",
        output: {
            file: "./public/index.html"
        },
        plugins: [
            execute("node generate-index.js"),
        ]
    }
];
