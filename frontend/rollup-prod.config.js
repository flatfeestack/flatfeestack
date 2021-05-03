import svelte from 'rollup-plugin-svelte';
import resolve from '@rollup/plugin-node-resolve';
import sveltePreprocess from 'svelte-preprocess';
import typescript from 'rollup-plugin-typescript2';
import css from 'rollup-plugin-css-only';
import commonjs from '@rollup/plugin-commonjs';
import { terser } from "rollup-plugin-terser";
import license from 'rollup-plugin-license';
import brotli from "rollup-plugin-brotli";
import gzipPlugin from 'rollup-plugin-gzip'

module.exports = {
  input: './src/main.ts',
  output: {
    format: 'iife',
    file: 'public/build/bundle.js',
    sourcemap: false
  },
  plugins: [
    svelte({ emitCss: false, preprocess: sveltePreprocess(), include: ['src/**/*.svelte', 'node_modules/svelte-*/src/**/*.svelte'],}),
    resolve({ browser: true, dedupe: ['svelte'], extensions: ['.ts', '.js'] }),
    typescript({ sourceMap: false,}),
    css({ output: 'bundle.css' }),
    commonjs(),
    terser({format: {comments: false}}),
    license({thirdParty: {output: 'public/dependencies.txt' }}),
    brotli(),
    gzipPlugin()
  ]
}
