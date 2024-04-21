const esbuild = require('esbuild');
const path = require('path');
const postcss = require('postcss');
const cssnano = require('cssnano');
const fs = require('fs');
const { lessLoader } = require('esbuild-plugin-less');
const glob = require("glob");

const args = process.argv.slice(2);
const buildCSS = args.includes('--css');
const buildJS = args.includes('--js');
(async () => {
    if (buildCSS) {
        doBuildCSS();
    }
    if (buildJS) {
        doBuildJS();
    }
})().catch((error) => {
    console.error('Build error:', error);
    process.exit(1);
});

async function doBuildCSS() {

    const lessEntryPoints = glob.sync('less/*.less', { ignore: 'less/*.mix.less' });

    await esbuild.build({
        entryPoints: lessEntryPoints,
        outdir: 'www/css',
        bundle: false,
        plugins: [
            lessLoader({}, {
                async transform(source) {
                    try {
                        const {css} = await postcss([cssnano]).process(source, {from: undefined});
                        return css;
                    } catch (error) {
                        handleError(error);
                        return '';
                    }
                },
            }),
        ],
    })
}

function handleError(error) {
    console.error('Error during LESS compilation:', error);
    // Additional error handling logic...
}

async function doBuildJS() {
    await esbuild.build({
        entryPoints: ['./clientScripts/*.js'],
        outdir: 'www/js/scripts',
        bundle: false,
        minify: true,
        sourcemap: true,
        platform: 'browser',
        target: 'es2015',
    })
    await esbuild.build({
        entryPoints: [
            './client/main.ts',
            './client/static/main.ts',
        ],
        outdir: path.resolve(__dirname, 'www', 'js'),
        bundle: true,
        sourcemap: true,
        splitting: true,
        minify: true,
        format: 'esm',
        target: 'es2015',
        entryNames: '[name].[hash]',
        publicPath: '/assets/js/',
        metafile: true,
        plugins: [
            {
                name: 'manifest',
                setup(build) {
                    const {outdir, publicPath} = build.initialOptions;
                    build.onEnd(async (result) => {
                        if (result.metafile) {
                            const manifest = {};
                            const outputs = result.metafile.outputs;

                            for (const file in outputs) {
                                const entryPoint = outputs[file].entryPoint;
                                if (entryPoint) {
                                    manifest[entryPoint] = path.join(publicPath, path.relative(outdir, file));
                                }
                            }

                            await fs.promises.writeFile(
                                path.resolve(__dirname, 'manifest.json'),
                                JSON.stringify(manifest, null, 2)
                            );
                        }
                    });
                },
            },
        ],
    })
}