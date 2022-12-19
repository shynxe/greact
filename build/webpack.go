package build

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/shynxe/greact/config"
)

type WebpackConfig struct {
	EntryPoints        string
	HtmlWebpackPlugins string
	BuildFolder        string
	StaticFolder       string
	PublicPath         string
}

var getSourceFiles = func() []string {
	fileNames := []string{}
	files, err := os.ReadDir(config.SourcePath)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}

	return fileNames
}

func getJSFiles(filenames []string) []string {
	var sourceFiles []string
	for _, filename := range filenames {
		// only add javascript files
		fileExt := filepath.Ext(filename)
		if fileExt == ".js" {
			sourceFiles = append(sourceFiles, filename)
		}
	}

	return sourceFiles
}

var getJSSourceFiles = func() []string {
	files := getSourceFiles()
	return getJSFiles(files)
}

func getWebpackConfig(userConfig config.Config) WebpackConfig {
	jsFiles := getJSSourceFiles()

	entryPoints := ""
	htmlWebpackPlugins := ""
	for _, file := range jsFiles {
		fileNameNoExt := filepath.Base(file)
		fileNameNoExt = fileNameNoExt[:len(fileNameNoExt)-len(".js")]
		entryPoints += fileNameNoExt + ": path.join(__dirname, '" + userConfig.SourceFolder + "', '" + file + "'),\n\t\t"
		htmlWebpackPlugins += "new HtmlWebpackPlugin({\n\t\t\ttemplate: path.join(__dirname, '" + userConfig.BuildFolder + "', '.greact-template.html'),\n\t\t\tfilename: '" + fileNameNoExt + ".html',\n\t\t\tchunks: ['hydrate', '" + fileNameNoExt + "'],\n\t\t\tpublicPath: '" + userConfig.PublicPath + "',\n\t\t}),\n\t\t"
	}

	return WebpackConfig{
		EntryPoints:        entryPoints,
		HtmlWebpackPlugins: htmlWebpackPlugins,
		BuildFolder:        userConfig.BuildFolder,
		StaticFolder:       userConfig.StaticFolder,
		PublicPath:         userConfig.PublicPath,
	}
}

func createWebpackConfig() error {
	userConfig := config.GetConfig()

	// create template
	tmpl, err := template.New("webpack").Parse(webpackConfigTemplate)
	if err != nil {
		return err
	}

	if _, err := os.Stat(userConfig.ClientPath); os.IsNotExist(err) {
		// create clientPath
		err = os.Mkdir(userConfig.ClientPath, 0755)
		if err != nil {
			return err
		}
	}

	// create webpack config for client pages
	f, err := os.Create(userConfig.ClientPath + "/webpack.config.js")
	if err != nil {
		return err
	}
	defer f.Close()

	webpackConfig := getWebpackConfig(userConfig)

	err = tmpl.Execute(f, webpackConfig)
	if err != nil {
		return err
	}

	// now create server webpack config
	tmpl, err = template.New("webpack-server").Parse(serverWebpackConfig)
	if err != nil {
		return err
	}

	f, err = os.Create(userConfig.ClientPath + "/server-webpack.config.js")
	if err != nil {
		return err
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
		hydrate: path.join(__dirname, "{{.BuildFolder}}", ".greact-hydrater.js"),
	},
	output: {
		path: path.join(__dirname, "{{.StaticFolder}}"),
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

const serverWebpackConfig = `const path = require('path');

module.exports = {
	entry: {        
		render: path.join(__dirname, "{{.BuildFolder}}", ".greact-renderer.js"),
	},
	output: {
		path: path.join(__dirname, "{{.BuildFolder}}"),
		filename: "[name].js",
		libraryTarget: "umd",
		library: "[name]",
        globalObject: 'this',
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
};`
