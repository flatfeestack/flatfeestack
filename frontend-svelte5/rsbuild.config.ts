import {defineConfig} from '@rsbuild/core';
import {pluginSvelte} from '@rsbuild/plugin-svelte';
import path from 'path';

export default defineConfig({
    environments: {
        // Configure the web environment for browsers
        web: {
            plugins: [
                pluginSvelte()
            ],
            source: {
                entry: {
                    index: './src/index-client.ts', //creates a html
                },
                alias: {
                    '/images': path.resolve(__dirname, 'public/images')
                }
            },
            output: {
                //assetPrefix: './',
                target: 'web'
            }
        },
        // Configure the node environment for SSR
        ssr: {
            plugins: [
                pluginSvelte({
                  svelteLoaderOptions: {
                    compilerOptions: {
                      //@ts-ignore -> this is the right option
                      generate: 'server'
                    }
                  }
                })
            ],
            source: {
                entry: {
                    server: './src/index-server.ts', //creates a js
                }
            },
            output: {
                // Use 'node' target for the Node.js outputs
                target: 'node'
            }
        }
    },
});