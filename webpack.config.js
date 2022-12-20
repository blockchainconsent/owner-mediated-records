const HtmlWebPackPlugin = require("html-webpack-plugin");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");

var path = require('path');

var APP_DIR = path.resolve(__dirname, './views');

module.exports = {
  entry: APP_DIR + '/index.jsx',
  output: {
    filename: 'bundledPayments.bundle.js',
    path: __dirname + "/dist/js"
  },
  // Emit source maps so we can debug our code in the browser
  devtool: 'source-map',
  module: {
    rules: [
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: {
          loader: "babel-loader"
        }
      },
      {
        test : /\.jsx?/,
        include : APP_DIR,
        loader : "babel-loader"
      },
      {
        test: /\.html$/,
        use: [
          {
            loader: "html-loader",
            options: { minimize: true }
          }
        ]
      },
      {
        test: /\.css$/,
        use: [MiniCssExtractPlugin.loader, "css-loader"]
      }
    ]
  },
  // Since Webpack only understands JavaScript, we need to
  // add a plugin to tell it how to handle html files.
  plugins: [
    // Configure HtmlWebPackPlugin to use our own index.html file
    // as a template.
    new HtmlWebPackPlugin({
      template: './views/index.html'
    }),
    new MiniCssExtractPlugin({
      filename: "[name].css",
      chunkFilename: "[id].css"
    })
  ]
};
