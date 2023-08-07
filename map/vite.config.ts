import {defineConfig} from 'vite'
import topLevelAwait from 'vite-plugin-top-level-await';
import {viteStaticCopy} from "vite-plugin-static-copy";

export default defineConfig({
    plugins: [
        topLevelAwait({
            // The export name of top-level await promise for each chunk module
            promiseExportName: "__tla",
            // The function to generate import names of top-level await promise in each chunk module
            promiseImportName: i => `__tla_${i}`
        }),
        viteStaticCopy({
            targets: [
                {
                    src: '../data/maptiles/*',
                    dest: 'tiles'
                },
                {
                    src: '../data/mapdata/*',
                    dest: 'data'
                }
            ]
        })
    ],
});