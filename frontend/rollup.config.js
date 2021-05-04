import svelte from 'rollup-plugin-svelte';
import resolve from '@rollup/plugin-node-resolve';
import sveltePreprocess from 'svelte-preprocess';
import typescript from 'rollup-plugin-typescript2';
import css from 'rollup-plugin-css-only';
import commonjs from '@rollup/plugin-commonjs';
import serve from 'rollup-plugin-serve'

module.exports = {
  input: './src/main.ts',
  output: {
    name: 'ffs',
    format: 'iife',
    file: 'public/build/bundle.js',
    sourcemap: true
  },
  plugins: [
    svelte({ emitCss: false, preprocess: sveltePreprocess()}),
    resolve({ browser: true, dedupe: ['svelte'], extensions: ['.ts', '.js'] }),
    typescript({ sourceMap: true,}),
    css({ output: 'bundle.css' }),
    commonjs(),
    serve({contentBase:'public', port:9085, host: '0.0.0.0', historyApiFallback: true})
  ]
}
