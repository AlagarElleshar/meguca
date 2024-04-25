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
    try {
        if (buildCSS) {
            console.log('Starting CSS Build...');
            await doBuildCSS();
            console.log('CSS Build completed successfully.');
        }
        if (buildJS) {
            console.log('Starting JS Build...');
            await doBuildJS();
            console.log('JS Build completed successfully.');
        }
    } catch (error) {
        handleError('Build failed', error);
    }
})();

async function doBuildCSS() {
    try {
        const lessEntryPoints = glob.sync('less/*.less', { ignore: 'less/*.mix.less' });
        console.log(`Building CSS from ${lessEntryPoints.length} LESS files...`);

        await esbuild.build({
            entryPoints: lessEntryPoints,
            outdir: 'www/css',
            bundle: false,
            plugins: [
                lessLoader(
                    {},
                    {
                        async transform(source) {
                            try {
                                const { css } = await postcss([cssnano]).process(source, { from: undefined });
                                return css;
                            } catch (error) {
                                throw new Error('CSS transformation error: ' + error.message);
                            }
                        },
                    },
                ),
            ],
        });
    } catch (error) {
        handleError('CSS Build error', error);
        throw error; // Rethrow to maintain failure state
    }
}

async function doBuildJS() {
    try {
        console.log('Building JavaScript...');

        await esbuild.build({
            entryPoints: ['./clientScripts/*.js'],
            outdir: 'www/js/scripts',
            bundle: false,
            minify: true,
            sourcemap: true,
            platform: 'browser',
            target: 'es2015',
        });

        console.log('Building additional JavaScript bundles...');

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
            entryNames: '[name]-[hash]',
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
        });
    } catch (error) {
        handleError('JavaScript Build error', error);
        throw error; // Rethrow to maintain failure state
    }
}

function handleError(context, error) {
    console.error(`${context}:`, error);
    process.exit(1);
}