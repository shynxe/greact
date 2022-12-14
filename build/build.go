package build

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shynxe/greact/config"
)

var (
	configPath string
	devMode    bool
)

// Build is the main function of the build command
func Build(args []string) {
	parseFlags(args)

	err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = build()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func parseFlags(args []string) {
	// flags:
	// -c, --config: path to config file
	// -dev: dev mode
	// new flagset for build command
	flagSet := flag.NewFlagSet("build", flag.ExitOnError)
	flagSet.StringVar(&configPath, "c", "", "path to config file")
	flagSet.StringVar(&configPath, "config", "", "path to config file")
	flagSet.BoolVar(&devMode, "dev", false, "dev mode")

	flagSet.Usage = func() {
		fmt.Println("usage: greact build [options]")
		fmt.Println()
		fmt.Println("options:")
		flagSet.PrintDefaults()
	}

	// parse flags
	flagSet.Parse(args)
}

func loadConfig() error {
	// if no config file is specified, use the default config file name
	if configPath == "" {
		configPath = config.DefaultConfigFileName
	}

	// load config file
	err := config.LoadConfig(configPath)

	if err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}

	// validate config file
	err = config.ValidateConfig(config.GetConfig())

	if err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}

	return nil
}

func build() error {
	// create client if it doesn't exist
	if !clientExists() {
		err := createClient()
		// TODO: create rollback method for client creation
		if err != nil {
			return fmt.Errorf("error creating client: %w", err)
		}

		err = createBuildPath()
		if err != nil {
			return fmt.Errorf("error creating default greact folder: %w", err)
		}

		err = createHydrater()
		if err != nil {
			return fmt.Errorf("error creating hydrater: %w", err)
		}
	} else if err := clientValid(); err != nil {
		return fmt.Errorf("invalid client: %w", err)
	}

	err := createHTMLTemplate()
	if err != nil {
		return fmt.Errorf("error creating html template: %w", err)
	}

	err = createRenderer()
	if err != nil {
		return fmt.Errorf("error creating renderer: %w", err)
	}

	// build client
	fmt.Println("building pages...")
	err = buildClient()
	if err != nil {
		return fmt.Errorf("error building pages: %w", err)
	}

	// print count of .html files in build folder
	fmt.Printf("successfully built %d pages!\n", countHTMLFiles())

	return nil
}

func getSourcePageNames() []string {
	files, err := os.ReadDir(config.SourcePath)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var pageNames []string
	for _, file := range files {
		filename := file.Name()
		fileExt := filepath.Ext(filename)

		if fileExt == ".js" {
			pageName := filename[:len(filename)-len(fileExt)]
			pageNames = append(pageNames, pageName)
		}
	}

	return pageNames
}

func countHTMLFiles() int {
	files, err := os.ReadDir(config.StaticPath)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	count := 0
	for _, file := range files {
		filename := file.Name()
		fileExt := filepath.Ext(filename)

		if fileExt == ".html" {
			count++
		}
	}

	return count
}

func createBuildPath() error {
	os.MkdirAll(config.BuildPath, os.ModePerm)

	return nil
}

func createHydrater() error {
	rendererPath := config.BuildPath + "/.greact-hydrater.js"
	_, err := os.Create(rendererPath)
	if err != nil {
		return err
	}

	err = os.WriteFile(
		rendererPath,
		[]byte(hydrater),
		os.ModePerm,
	)
	if err != nil {
		return err
	}

	return nil
}

func createRenderer() error {
	rendererPath := config.BuildPath + "/.greact-renderer.js"
	_, err := os.Create(rendererPath)
	if err != nil {
		return err
	}

	renderer := "import React from 'react';\nimport * as ReactDOMServer from 'react-dom/server';\n"

	// get all page names
	pageNames := getSourcePageNames()

	// create import statements, src is the SOURCEFOLDER of config
	for _, pageName := range pageNames {
		renderer += fmt.Sprintf("import %s from '../%s/%s.js';\n", strings.Title(pageName), config.GetConfig().SourceFolder, pageName)
	}

	// create render functions
	for _, pageName := range pageNames {
		renderer += fmt.Sprintf("const render%s = (props) => {\n", strings.Title(pageName))
		renderer += fmt.Sprintf("    return ReactDOMServer.renderToString(React.createElement(%s, props));\n", strings.Title(pageName))
		renderer += "}\n\n"
	}

	// create render function
	renderer += "const render = (page, props) => {\n"
	for _, pageName := range pageNames {
		renderer += fmt.Sprintf("    if (page === '%s') {\n", pageName)
		renderer += fmt.Sprintf("        return render%s(props);\n", strings.Title(pageName))
		renderer += "    }\n"
	}
	renderer += "}\n\n"

	// export render function
	renderer += "export default render;\n"

	err = os.WriteFile(
		rendererPath,
		[]byte(renderer),
		os.ModePerm,
	)
	if err != nil {
		return err
	}

	return nil
}

func createHTMLTemplate() error {
	templatePath := config.BuildPath + "/.greact-template.html"
	_, err := os.Create(templatePath)
	if err != nil {
		return err
	}

	// if devMode, add refreshScript to HTMLTemplate
	var htmlTemplate string
	if devMode {
		htmlTemplate = strings.Replace(HTMLTemplate, "</head>", refreshScript+"</head>", 1)
	} else {
		htmlTemplate = HTMLTemplate
	}

	err = os.WriteFile(
		templatePath,
		[]byte(htmlTemplate),
		os.ModePerm,
	)
	if err != nil {
		return err
	}

	return nil
}

func clientValid() error {
	// check if sourcePath directory exists
	if _, err := os.Stat(config.SourcePath); os.IsNotExist(err) {
		return fmt.Errorf("sourcePath directory does not exist")
	}

	// check if node_modules directory exists
	if _, err := os.Stat(config.GetConfig().ClientPath + "/node_modules"); os.IsNotExist(err) {
		return fmt.Errorf("node_modules directory does not exist")
	}

	return nil
}

func createClient() error {
	// create clientPath directory
	fmt.Println("creating client...")
	err := os.MkdirAll(config.GetConfig().ClientPath, os.ModePerm)
	if err != nil {
		return err
	}

	// create clientPath/src directory and add a simple index.js file react page
	err = os.MkdirAll(config.SourcePath, os.ModePerm)
	if err != nil {
		return err
	}

	indexFile, err := os.Create(config.SourcePath + "/index.js")
	if err != nil {
		return err
	}

	// write sample react page to index.js
	_, err = indexFile.WriteString(reactSamplePage)
	if err != nil {
		return err
	}

	// create package.json file
	packageFile, err := os.Create(config.GetConfig().ClientPath + "/package.json")
	packageFile.WriteString(packageJSON)
	if err != nil {
		return err
	}

	// install dependencies
	err = installDependencies()
	return err
}

func installDependencies() error {
	fmt.Println("installing dependencies...")

	currentDir, _ := os.Getwd()
	err := os.Chdir(config.GetConfig().ClientPath)
	if err != nil {
		return err
	}

	var dependencies = []string{
		"react",
		"react-dom",
	}

	var devDependencies = []string{
		"@babel/cli",
		"@babel/core",
		"@babel/plugin-transform-react-jsx-source",
		"@babel/preset-react",
		"babel-loader",
		"html-webpack-plugin",
		"webpack",
		"webpack-cli",
		"webpack-dev-server",
	}

	// install dependencies
	for _, dependency := range dependencies {
		output := exec.Command("npm", "install", dependency)
		if err := output.Run(); err != nil {
			return err
		}
	}

	// install dev dependencies
	for _, dependency := range devDependencies {
		output := exec.Command("npm", "install", dependency, "--save-dev")
		if err := output.Run(); err != nil {
			return err
		}
	}

	os.Chdir(currentDir)
	return nil
}

func buildClient() error {
	// call webpack in clientPath directory
	err := createWebpackConfig()
	if err != nil {
		return fmt.Errorf("error creating webpack config: %w", err)
	}

	currentDir, _ := os.Getwd()
	os.Chdir(config.GetConfig().ClientPath)
	output := exec.Command("npx", "webpack", "--mode", "production")
	if err := output.Run(); err != nil {
		return err
	}

	// now call it with config server-webpack.config.js
	output = exec.Command("npx", "webpack", "--mode", "production", "--config", "server-webpack.config.js")
	if err := output.Run(); err != nil {
		return err
	}

	os.Chdir(currentDir)
	return nil
}

func clientExists() bool {
	if _, err := os.Stat(config.GetConfig().ClientPath); os.IsNotExist(err) {
		return false
	}
	return true
}

const reactSamplePage = `import React from 'react';

const App = ({name}) => {
    const [count, setCount] = React.useState(0);

    React.useEffect(() => {
        if (count === 10) {
            setCount(0);
        }
    }, [count]);

    return (
        <div>
			{ name ? <h1>Welcome to gReact, {name}!</h1> : <h1>Welcome to gReact!</h1> }
            <h2>Count: {count}</h2>
            <button onClick={() => setCount(count + 1)}>Increment</button>
        </div>
    );
}

export default App;`

const packageJSON = `{
	"name": "greact",
	"version": "1.0.0",
	"author": ""
}`

const HTMLTemplate = `<html>

<head>
  <title>SSR Demo</title>
  <meta charset="utf-8" />
</head>

<body>
  <div id="root">{{SSR}}</div>
</body>
<script>
  const hydrateDOM = (fn) => {
    if (document.readyState != 'loading') {
      fn();
    } else {
      document.addEventListener('DOMContentLoaded', fn);
    }
  }

  hydrateDOM(function () {
    // {{__HYDRATION__}}
  })
</script>

</html>`

const hydrater = `import React from 'react';
import ReactDOM from 'react-dom';

const _page = ({component, props}) => {
    return React.createElement(component, props);    
}

const hydrate = (component, props) => {
    ReactDOM.hydrate(_page({component, props}), document.getElementById('root'));
}

export default hydrate;`
