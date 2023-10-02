import typescript from '@rollup/plugin-typescript';
import { nodeResolve } from '@rollup/plugin-node-resolve';
import html from '@open-wc/rollup-plugin-html';
import copy from 'rollup-plugin-copy'
import postcss from 'rollup-plugin-postcss'
import url from '@rollup/plugin-url';
import { terser } from "rollup-plugin-terser";

let plugins = [
    postcss({
        minimize: true,
        plugins: []
    }),
    url({
        limit: 2048,
        include: ['**/*.svg', '**/*.png', '**/*.jpg', '**/*.gif'],
        fileName: '[name]-[hash][extname]',
    }),
    html({ name: "index.html" }),
    typescript(),
    nodeResolve(),
    terser(),
];

export default [{
    input: 'viewer.html',
    external: ['leaflet'],
    output: {
        dir: 'dist/viewer',
        format: 'iife',
        sourcemap: true,
        entryFileNames: '[name]-[hash].js',
        globals: {
            leaflet: 'L',
        },
    },
    plugins: [
        copy({
            targets: [{ src: '../thumbnail.png', dest: 'dist/viewer' }]
        }),
        copy({
            targets: [{ src: "manifest.json", dest: "dist/viewer" }],
        }),
    ].concat(plugins),
}, {
    input: 'listing.html',
    external: ['leaflet'],
    output: {
        dir: 'dist/listing',
        format: 'iife',
        sourcemap: true,
        entryFileNames: '[name]-[hash].js',
    },
    plugins: [
        copy({
            targets: [{ src: '../thumbnail.png', dest: 'dist/listing' }]
        }),
    ].concat(plugins),
}];
