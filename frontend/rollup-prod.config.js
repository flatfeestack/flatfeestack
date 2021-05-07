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
    format: 'esm',
    dir: 'public/build',
    sourcemap: false,
    manualChunks: {
      deps: ['ethers', 'ky', 'svelte-routing', '@stripe/stripe-js'],
    }
  },
  plugins: [
    svelte({ emitCss: false, preprocess: sveltePreprocess()}),
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
