const path = require('path');

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

module.exports = [
    {
        ...commonConfig,
        entry: './client/main.ts',
        output: {
            path: path.resolve(__dirname, 'www', "js"),
            filename: 'main.js'
        }
    },
    {
        ...commonConfig,
        entry: './client/static/main.ts',
        output: {
            path: path.resolve(__dirname, 'www', "js","static"),
            filename: 'main.js'
        }
    }
]