const CopyWebpackPlugin = require('copy-webpack-plugin');
const ForkTsCheckerWebpackPlugin = require('fork-ts-checker-webpack-plugin');
const path = require('path');
const ReplaceInFileWebpackPlugin = require('replace-in-file-webpack-plugin');

const DIST_DIR = path.resolve(__dirname, 'dist');
const SRC_DIR = path.resolve(__dirname, 'src');

module.exports = async (env) => {
  const isProduction = Boolean(env.production);

  return {
    context: SRC_DIR,
    mode: isProduction ? 'production' : 'development',
    target: 'web',

    entry: {
      module: path.resolve(SRC_DIR, 'module.ts'),
    },

    output: {
      // Keep backend binaries (gpx_*) when cleaning — only wipe JS/CSS/JSON artifacts
      clean: isProduction ? { keep: /^gpx_/ } : false,
      filename: '[name].js',
      path: DIST_DIR,
      publicPath: '/',
      library: {
        type: 'amd',
      },
    },

    externals: [
      'lodash',
      'jquery',
      'moment',
      'slate',
      'emotion',
      '@emotion/react',
      '@emotion/css',
      'prismjs',
      'slate-plain-serializer',
      '@grafana/slate-react',
      'react',
      'react-dom',
      'react-redux',
      'redux',
      'rxjs',
      'react-router-dom',
      'd3',
      'angular',
      '@grafana/ui',
      '@grafana/runtime',
      '@grafana/data',
      '@grafana/schema',
      /^@grafana\//,
    ],

    plugins: [
      new ForkTsCheckerWebpackPlugin({
        async: Boolean(env.development),
        issue: {
          include: [{ file: '**/*.{ts,tsx}' }],
        },
        typescript: {
          configFile: path.join(__dirname, 'tsconfig.build.json'),
        },
      }),

      new CopyWebpackPlugin({
        patterns: [
          { from: 'img/**/*', to: DIST_DIR, noErrorOnMissing: true },
          { from: 'plugin.json', to: DIST_DIR },
          { from: '../README.md', to: DIST_DIR, noErrorOnMissing: true },
          { from: '../CHANGELOG.md', to: DIST_DIR, noErrorOnMissing: true },
          { from: '../LICENSE', to: DIST_DIR, noErrorOnMissing: true },
        ],
      }),

      new ReplaceInFileWebpackPlugin([
        {
          dir: DIST_DIR,
          files: ['plugin.json'],
          rules: [
            {
              search: '%VERSION%',
              replace: process.env.npm_package_version ?? '0.1.0',
            },
          ],
        },
      ]),
    ],

    resolve: {
      extensions: ['.js', '.jsx', '.ts', '.tsx'],
      modules: [path.resolve(__dirname, 'node_modules'), 'node_modules'],
    },

    module: {
      rules: [
        {
          test: /\.[tj]sx?$/,
          exclude: /node_modules/,
          use: {
            loader: 'swc-loader',
            options: {
              jsc: {
                baseUrl: SRC_DIR,
                target: 'es2018',
                loose: false,
                parser: {
                  syntax: 'typescript',
                  tsx: true,
                  decorators: false,
                  dynamicImport: true,
                },
              },
            },
          },
        },
        {
          test: /\.css$/,
          use: ['style-loader', 'css-loader'],
        },
        {
          test: /\.(scss|sass)$/,
          use: ['style-loader', 'css-loader', 'sass-loader'],
        },
        {
          test: /\.(png|jpe?g|gif|svg)$/,
          type: 'asset/resource',
          generator: {
            filename: 'img/[hash][ext][query]',
          },
        },
        {
          test: /\.(woff|woff2|eot|ttf|otf)$/,
          type: 'asset/resource',
          generator: {
            filename: 'fonts/[hash][ext][query]',
          },
        },
      ],
    },

    devtool: isProduction ? false : 'source-map',

    optimization: {
      minimize: isProduction,
    },
  };
};
