//jshint strict: false
module.exports = function (config) {
    config.set({

        basePath: './',

        files: [
            "node_modules/angular/angular.min.js",
            "node_modules/angular-mocks/angular-mocks.js",
            "moment.min.js",
            "todo.js",
            "todo.spec.js",
        ],

        autoWatch: true,

        frameworks: ['jasmine'],

        browsers: ['PhantomJS', 'PhantomJS_custom'],
        customLaunchers: {
            'PhantomJS_custom': {
                base: 'PhantomJS',
                options: {
                    windowName: 'my-window',
                    settings: {
                        webSecurityEnabled: false
                    },
                },
                flags: ['--load-images=true'],
                debug: true
            }
        },
        phantomjsLauncher: {
            exitOnResourceError: true
        },

        plugins: [
            'karma-phantomjs-launcher',
            'karma-jasmine'
        ]

    });
};
