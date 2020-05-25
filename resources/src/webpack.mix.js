const mix = require('laravel-mix');
require('laravel-mix-purgecss');

// ...
mix.setPublicPath('../assets/')
    .setResourceRoot('./src')
    /*.webpackConfig(() => ({
        resolve: {
            modules: ['src/js', 'node_modules'],
        },
    }))*/
    .js('src/assets/js/app.js', 'js')
    .sass('src/assets/sass/app.scss', 'css')
    .copyDirectory('src/icon', 'public/icon')
    .copyDirectory('src/images', 'public/images')
    .copyDirectory('src/video', 'public/video')
    .copyDirectory('src/fonts', 'public/fonts')
    .sourceMaps()
    .purgeCss()
    .browserSync('localhost:8080');