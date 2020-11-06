import typescript from '@rollup/plugin-typescript';
import { nodeResolve } from '@rollup/plugin-node-resolve';
import html from '@open-wc/rollup-plugin-html';
import copy from 'rollup-plugin-copy'
import postcss from 'rollup-plugin-postcss'
import url from '@rollup/plugin-url';

export default {
    input: 'index.html',
    external: ['leaflet'],
    output: {
        dir: 'dist',
        format: 'iife',
        sourcemap: true,
        entryFileNames: '[name]-[hash].js',
        globals: {
            leaflet: 'L',
        },
    },
    plugins: [
        copy({
            targets: [{ src: '../thumbnail.png', dest: 'dist' }]
        }),
        postcss({
            minimize: true,
            plugins: []
        }),
        url({
            limit: 2048,
            include: ['**/*.svg', '**/*.png', '**/*.jpg', '**/*.gif'],
            fileName: '[name]-[hash][extname]',
        }),
        html(),
        typescript(),
        nodeResolve(),
    ],
};