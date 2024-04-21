const esbuild = require('esbuild');
const path = require('path');
const fs = require('fs');

esbuild.build({
  entryPoints: [
    './client/main.ts',
    './client/static/main.ts',
  ],
  outdir: path.resolve(__dirname, 'www', 'js'),
  bundle: true,
  sourcemap: true,
  splitting: true,
  format: 'esm',
  target: 'es2015',
  entryNames: '[name].[hash]',
  publicPath: '/assets/js/',
  metafile: true,
  plugins: [
    {
      name: 'manifest',
      setup(build) {
        build.onEnd(async (result) => {
          if (result.metafile) {
            const manifest = {};
            const outputs = result.metafile.outputs;

            for (const file in outputs) {
              const entryPoint = outputs[file].entryPoint;
              if (entryPoint) {
                manifest[entryPoint] = file.replace('www/js', '/assets/js');
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
}).catch(() => process.exit(1));