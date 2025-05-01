const esbuild = require('esbuild');
const path = require('path');
const postcss = require('postcss');
const cssnano = require('cssnano');
const fs = require('fs');
const { lessLoader } = require('esbuild-plugin-less');
const glob = require("glob");

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

doBuildCSS();