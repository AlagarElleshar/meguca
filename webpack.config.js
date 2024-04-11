const path = require('path');
const {WebpackManifestPlugin} = require("webpack-manifest-plugin");

const commonConfig = {
    devtool: 'source-map',
    module: {
        rules: [
            {
                test: /\.ts$/,
                use: 'ts-loader',
            }
        ]
    },
    resolve: {
        extensions: ['.ts', '.js']
    },
}


module.exports = {
    ...commonConfig,
    entry: {
        main: './client/main.ts',
        static: './client/static/main.ts',
    },
    output: {
        path: path.resolve(__dirname, 'www', "js"),
        filename: '[name].[contenthash].js',
        publicPath: '/assets/js',
    },
    plugins: [
        new WebpackManifestPlugin({
            fileName: '../../manifest.json',
            basePath: '',
        }),
    ],
};