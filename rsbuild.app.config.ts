import {defineConfig} from '@rsbuild/core';

export default defineConfig({
    source: {
        entry: {
            // These correspond to your second esbuild call (bundle: true)
            main: './client/main.ts',
            staticMain: './client/static/main.ts',
        },
        tsconfigPath: './client/tsconfig.json',
    },
    output: {

        cleanDistPath: false,
        distPath: {
            root: 'www', // Base output directory
            js: 'js',    // Subdirectory for JS files (from entries)
        },
        target: "web",
        minify: true,
        sourceMap: {
            js:'source-map'
        },
        filename: {
            js: '[name]-[contenthash].js',
            css: '[name]-[contenthash].css',
        },
        manifest: {
            generate: ({manifestData}) => {
                return {
                    main: '/assets' + manifestData.entries.main.initial.js[0],
                    staticMain: '/assets' + manifestData.entries.staticMain.initial.js[0],
                };
            }, // Generates an asset-manifest.json (or similar name)
        }
    },
    tools: {
        htmlPlugin: false,
    },
    performance: {
        chunkSplit: {
            strategy: 'all-in-one'
        },
    }
});