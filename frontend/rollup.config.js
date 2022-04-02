import svelte from "rollup-plugin-svelte";
import resolve from "@rollup/plugin-node-resolve";
import sveltePreprocess from "svelte-preprocess";
import typescript from "rollup-plugin-typescript2";
import css from "rollup-plugin-css-only";
import commonjs from "@rollup/plugin-commonjs";
import execute from "rollup-plugin-execute";
import json from '@rollup/plugin-json';
import builtins from 'rollup-plugin-node-builtins';
import globals from 'rollup-plugin-node-globals';
import copy from 'rollup-plugin-copy'
import serve from "rollup-plugin-serve";

export default [
    {
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
            svelte({emitCss: true, preprocess: sveltePreprocess(), compilerOptions: {hydratable: true}}),
            resolve({preferBuiltins: false, browser: true, dedupe: ["svelte"], extensions: [".ts", ".js"]}),
            typescript({sourceMap: true}),
            css({output: "bundle.css"}),
            commonjs({transformMixedEsModules: true}),
            globals(),
            builtins(),
            json(),
            copy({targets: [{ src: 'landing-page/public/images', dest: 'public' }]}),
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
