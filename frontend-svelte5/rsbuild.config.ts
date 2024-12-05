import {defineConfig} from '@rsbuild/core';
import {pluginSvelte} from '@rsbuild/plugin-svelte';
import * as process from "node:process";
import {pluginCssMinimizer} from "@rsbuild/plugin-css-minimizer";
import { convert } from 'tsconfig-to-swcconfig';
import { resolve } from 'path';

const swcConfig = convert("tsconfig.json"); // This will look for tsconfig.json in the current directory
swcConfig.env = null;
swcConfig.jsc.baseUrl = resolve(__dirname, 'src')

export default defineConfig({
    environments: {
        // Configure the web environment for browsers
        web: {
            plugins: [
                pluginSvelte(),
                pluginCssMinimizer()
            ],
            source: {
                entry: {
                    index: './src/index.ts',
                }
            },
            output: {
                target: 'web',
                minify:  process.env.NODE_ENV === 'production',
            }
        }
    },
    dev: {
        hmr: false,
        liveReload: true,
    },
    tools: {
        swc: swcConfig,
    }
});
