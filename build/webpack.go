package build

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/shynxe/greact/config"
)

func createWebpackConfig() error {
	// get userConfig
	userConfig := config.GetConfig()

	files, err := os.ReadDir(config.SourcePath)
	if err != nil {
		return err
	}

	entryPoints := ""
	htmlWebpackPlugins := ""
	for _, file := range files {
		// only add javascript files
		fileExt := filepath.Ext(file.Name())
		if fileExt == ".js" {
			fileNameNoExt := filepath.Base(file.Name())
			fileNameNoExt = fileNameNoExt[:len(fileNameNoExt)-len(fileExt)]
			entryPoints += fileNameNoExt + ": path.join(__dirname, '" + userConfig.SourceFolder + "', '" + file.Name() + "'),\n\t\t"
			htmlWebpackPlugins += "new HtmlWebpackPlugin({\n\t\t\ttemplate: path.join(__dirname, '" + userConfig.BuildFolder + "', '.greact-template.html'),\n\t\t\tfilename: '" + fileNameNoExt + ".html',\n\t\t\tchunks: ['render', '" + fileNameNoExt + "'],\n\t\t\tpublicPath: '" + userConfig.PublicPath + "',\n\t\t}),\n\t\t"
		}
	}

	// create template
	tmpl, err := template.New("webpack").Parse(webpackConfigTemplate)
	if err != nil {
		return err
	}

	// create file
	// check if clientPath exists
	if _, err := os.Stat(userConfig.ClientPath); os.IsNotExist(err) {
		// create clientPath
		err = os.Mkdir(userConfig.ClientPath, 0755)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(userConfig.ClientPath + "/webpack.config.js")
	if err != nil {
		return err
	}
	defer f.Close()

	type WebpackConfig struct {
		EntryPoints        string
		HtmlWebpackPlugins string
		BuildPath          string
		StaticPath         string
		PublicPath         string
	}

	webpackConfig := WebpackConfig{
		EntryPoints:        entryPoints,
		HtmlWebpackPlugins: htmlWebpackPlugins,
		BuildPath:          userConfig.BuildFolder,
		StaticPath:         userConfig.StaticFolder,
		PublicPath:         userConfig.PublicPath,
	}

	err = tmpl.Execute(f, webpackConfig)
	if err != nil {
		return err
	}

	return nil
}

const webpackConfigTemplate = `const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');

module.exports = {
	entry: {
		{{.EntryPoints}}
		render: path.join(__dirname, "{{.BuildPath}}", ".greact-renderer.js"),
	},
	output: {
		path: path.join(__dirname, "{{.StaticPath}}"),
		filename: "[name].[contenthash:8].js",
		libraryTarget: "umd",
		library: "[name]",
		clean: true,
	},
	module: {
		rules: [
			{
				test: /\.?js$/,
				exclude: /node_modules/,
				use: {
					loader: "babel-loader",
					options: {
						plugins: ['@babel/plugin-transform-react-jsx']
					}
				}
			},
		]
	},
	optimization: {
		runtimeChunk: 'single',
		chunkIds: 'deterministic',
		splitChunks: {
			chunks: 'all',
			maxInitialRequests: Infinity,
			minSize: 0,
			cacheGroups: {
				vendor: {
					test: /[\\/]node_modules[\\/]/,
					name(module) {
						const packageName = module.context.match(/[\\/]node_modules[\\/](.*?)([\\/]|$)/)[1];
						return ` + "`" + `npm.${packageName.replace('@', '')}` + "`" + `;
					},
				},
			},
		},
	},
	plugins: [
		{{.HtmlWebpackPlugins}}
	],
};`
