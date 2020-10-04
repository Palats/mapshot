import typescript from '@rollup/plugin-typescript';
import { nodeResolve } from '@rollup/plugin-node-resolve';
import html from '@open-wc/rollup-plugin-html';

export default {
    input: 'index.html',
    output: {
        dir: 'dist',
        format: 'iife',
        sourcemap: true,
        entryFileNames: '[name]-[hash].js',
    },
    plugins: [
        html(),
        nodeResolve(),
        typescript(),
    ],
};