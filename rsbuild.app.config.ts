import {defineConfig} from '@rsbuild/core';
import {pluginLess} from '@rsbuild/plugin-less';

export default defineConfig({
    plugins: [
        pluginLess(),
    ],
    source: {
        entry: {
            // These correspond to your second esbuild call (bundle: true)
            main: './client/main.ts',
            staticMain: './client/static/main.ts',
        },
    },
    output: {
        distPath: {
            root: 'www', // Base output directory
            js: 'js',    // Subdirectory for JS files (from entries)
        },
        target: "web",
        minify: true,
        sourceMap: {
            js: 'cheap-module-source-map',
            css: true,
        },
        filename: {
            js: '[name]-[contenthash].js',
            css: '[name]-[contenthash].css',
        },
        manifest: true, // Generates an asset-manifest.json (or similar name)
    },
    tools: {
        htmlPlugin: false,
    }
});