const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const CopyWebpackPlugin = require('copy-webpack-plugin');

module.exports = {
    mode: 'development',
    entry: './client.js',
    output: {
        path: path.resolve(__dirname, 'dist'),
        filename: 'bundle.js',
    },
    target: 'web',
    resolve: {
        extensions: ['.js'],
        fallback: {
            "path": false,
            "fs": false,
        },
    },
    plugins: [
        new HtmlWebpackPlugin({
            template: './index.html',
            filename: 'index.html',
        }),
        new HtmlWebpackPlugin({
            template: './tasks.html',
            filename: 'tasks.html',
        }),
        new CopyWebpackPlugin({
            patterns: [
                { from: 'style.css', to: 'style.css' },
                { from: 'favicon.ico', to: 'favicon.ico' }, // Проверяем эту строку
            ],
        }),
    ],
};