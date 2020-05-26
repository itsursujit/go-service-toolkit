const mix = require('laravel-mix');
require('laravel-mix-purgecss');

// ...
mix.js('resources/assets/js/app.js', 'assets/js')
    .sass('resources/assets/sass/app.scss', 'assets/css')
    .copy('resources/assets/images', 'assets/images')
    .extract()
    .purgeCss()
    .browserSync('http://localhost:8080');