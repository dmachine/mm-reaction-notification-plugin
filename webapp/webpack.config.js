const path = require('path');

module.exports = (env, argv) => {
    const isDevMode = argv.mode === 'development';

    return {
        entry: './src/index.js',
        resolve: {
            modules: [path.resolve(__dirname, 'src'), 'node_modules'],
            extensions: ['*', '.js', '.jsx'],
        },
        module: {
            rules: [
                {
                    test: /\.(js|jsx)$/,
                    exclude: /node_modules/,
                    use: {
                        loader: 'babel-loader',
                        options: {
                            cacheDirectory: true,
                            presets: [
                                '@babel/preset-env',
                                '@babel/preset-react',
                            ],
                        },
                    },
                },
            ],
        },
        // These packages are provided by Mattermost at runtime; do not bundle them.
        externals: {
            react: 'React',
            'react-dom': 'ReactDOM',
            redux: 'Redux',
            'react-redux': 'ReactRedux',
            'prop-types': 'PropTypes',
        },
        output: {
            path: path.join(__dirname, 'dist'),
            filename: 'main.js',
            publicPath: '/',
        },
        devtool: isDevMode ? 'eval-source-map' : false,
        mode: isDevMode ? 'development' : 'production',
    };
};
