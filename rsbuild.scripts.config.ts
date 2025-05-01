// rsbuild.scripts.config.ts
import { defineConfig } from '@rsbuild/core';
import * as path from 'path';
import { globSync } from 'glob';

// Helper function (same as before)
function getClientScriptEntries(): Record<string, string> {
  const entries: Record<string, string> = {};
  const files = globSync('./clientScripts/*.js');
  files.forEach((filePath) => {
    const baseName = path.basename(filePath, path.extname(filePath));
    // Key determines output path relative to output.distPath.js
    const entryKey = `scripts/${baseName}`; // -> outputs to js/scripts/
    entries[entryKey] = "./" + filePath;
  });
  return entries;
}

export default defineConfig({
  // No plugins needed here unless scripts import CSS/Less etc.
  source: {
    // Only include the client scripts
    entry: getClientScriptEntries(),
  },
  output: {
    cleanDistPath: false,
    distPath: {
      root: 'www', // IMPORTANT: Same root output directory
      js: 'js',    // Base JS output directory (www/js/)
                    // The entry key 'scripts/...' places files in www/js/scripts/
      // No css path needed if scripts don't output CSS
    },
    target: "web",
    // IMPORTANT: Filename *without* hash for these scripts
    filename: {
      js: '[name].js', // -> e.g., scripts/scriptA.js
      // css: '[name].css', // Only if scripts could produce CSS
    },
    manifest: false,
    sourceMap: {
      js: "cheap-source-map",
      // css: true,
    },
    // Ensure minification is enabled for production builds
    minify: true,
  },
  tools:{
    htmlPlugin: false,
  }
});