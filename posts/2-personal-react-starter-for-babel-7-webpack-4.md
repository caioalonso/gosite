---
id: 2
title: Personal React App Starter for Babel 7, Webpack 4
date: 2018-11-27
aliases:
    - /posts/personal-react-starter-for-babel-7-webpack-4
    - /2018/11/27/personal-react-starter-for-babel-7-webpack-4.html
---

Here's a personal record of the steps I take to start a React app from zero, understanding what every single dependency does.

If you just want a clean React starting point, clone the repository:
```
git clone https://github.com/caioalonso/react-boilerplate
```

Start here:
```
mkdir projectname
cd projectname
mkdir src
touch src/index.html
touch src/index.js
npm init -y
```

Now the dependencies:
```
npm add --save-dev webpack webpack-cli html-webpack-plugin html-loader webpack-dev-server \
@babel/core babel-loader @babel/preset-env @babel/preset-react \
react react-dom \
prettier
```

- `webpack` bundles js modules, css, compresses images, etc;
- `webpack-cli` cli commands for webpack;
- `html-webpack-plugin` generates the index.html with the correct hashes etc;
- `html-loader` helps the webpack plugin deal with HTML files;
- `webpack-dev-server` serves a local self-refreshing version of the app;
- `@babel/core` core of the Babel compiler, most of the official Babel stuff sits under their @ scope;
- `babel-loader` is a Webpack [loader](https://webpack.js.org/loaders/) to use Babel in it;
- `@babel/preset-env` smart Babel preset that changes depending on which platform you're compiling to;
- `@babel/preset-react` a Babel preset for compiling React stuff (JSX etc);
- `react` just what is needed for defining React components;
- `react-dom` React renderer for the web;
- `prettier` help keep js and css pretty;

Scripts in `package.json`:
```
"scripts": {
	"start": "webpack-dev-server --open --mode development",
    "build": "webpack --mode production",
}
```

Import the presets in `.babelrc`:
```
{
  "presets": [
    "@babel/preset-env",
    "@babel/preset-react"
  ]
}
```

Import the loaders and plugins in `webpack.config.js`:
```
const HtmlWebPackPlugin = require("html-webpack-plugin");

module.exports = {
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
        test: /\.html$/,
        use: [
          {
            loader: "html-loader"
          }
        ]
      }
    ]
  },
  plugins: [
    new HtmlWebPackPlugin({
      template: "./src/index.html",
      filename: "./index.html"
    })
  ]
};
```

Now edit `src/index.html` and `src/index.js` however you'd like, run `npm start` to start the dev server, run `npm run build` to compile it all into `dist`.
