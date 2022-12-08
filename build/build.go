package build

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shynxe/greact-cli/config"
)

var (
	configPath string
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
	// new flagset for build command
	flagSet := flag.NewFlagSet("build", flag.ExitOnError)
	flagSet.StringVar(&configPath, "c", "", "path to config file")
	flagSet.StringVar(&configPath, "config", "", "path to config file")

	flagSet.Usage = func() {
		fmt.Println("usage: greact-cli build [options]")
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

		err = createWebpackConfig()
		if err != nil {
			return fmt.Errorf("error creating webpack config: %w", err)
		}

		err = createBuildPath()
		if err != nil {
			return fmt.Errorf("error creating default greact folder: %w", err)
		}

		err = createHTMLTemplate()
		if err != nil {
			return fmt.Errorf("error creating html template: %w", err)
		}

		err = createRenderer()
		if err != nil {
			return fmt.Errorf("error creating renderer: %w", err)
		}
	} else if err := clientValid(); err != nil {
		return fmt.Errorf("invalid client: %w", err)
	}

	// build client
	fmt.Println("building pages...")
	err := buildClient()
	if err != nil {
		return fmt.Errorf("error building pages: %w", err)
	}

	// print count of .html files in build folder
	fmt.Printf("successfully built %d pages!", countHTMLFiles())

	return nil
}

func countHTMLFiles() int {
	files, err := os.ReadDir(config.BuildPath)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	count := 0
	for _, file := range files {
		fileExt := filepath.Ext(file.Name())

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

func createRenderer() error {
	rendererPath := config.BuildPath + "/.greact-renderer.js"
	_, err := os.Create(rendererPath)
	if err != nil {
		return err
	}

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

	err = os.WriteFile(
		templatePath,
		[]byte(HTMLTemplate),
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
	os.MkdirAll(config.GetConfig().ClientPath, os.ModePerm)

	// create clientPath/src directory and add a simple index.js file react page
	os.MkdirAll(config.SourcePath, os.ModePerm)
	os.Create(config.SourcePath + "/index.js")
	// write sample react page to index.js
	os.WriteFile(config.SourcePath+"/index.js", []byte(reactSamplePage), os.ModePerm)

	// create package.json file
	os.Create(config.GetConfig().ClientPath + "/package.json")
	os.WriteFile(config.GetConfig().ClientPath+"/package.json", []byte(packageJSON), os.ModePerm)

	// run npm install and wait for it to finish
	fmt.Println("installing dependencies...")

	currentDir, _ := os.Getwd()
	os.Chdir(config.GetConfig().ClientPath)

	output := exec.Command("npm", "install")
	if err := output.Run(); err != nil {
		return fmt.Errorf("error installing dependencies (%w)", err)
	}
	os.Chdir(currentDir)
	return nil
}

func buildClient() error {
	// call webpack in clientPath directory
	currentDir, _ := os.Getwd()
	os.Chdir(config.GetConfig().ClientPath)
	output := exec.Command("npx", "webpack", "--mode", "production")
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
	"description": "",
	"main": "index.js",
	"scripts": {
	  "build": "webpack --mode production",
	  "dev": "webpack-dev-server --mode development --open --hot",
	  "test": "echo \"Error: no test specified\" && exit 1"
	},
	"author": "",
	"license": "ISC",
	"devDependencies": {
	  "@babel/cli": "^7.19.3",
	  "@babel/core": "^7.20.5",
	  "@babel/plugin-transform-react-jsx-source": "^7.19.6",
	  "@babel/preset-react": "^7.18.6",
	  "babel-loader": "^9.1.0",
	  "html-webpack-plugin": "^5.5.0",
	  "webpack": "^5.75.0",
	  "webpack-cli": "^5.0.0",
	  "webpack-dev-server": "^4.11.1"
	},
	"dependencies": {
	  "react": "^18.2.0",
	  "react-dom": "^18.2.0"
	}
  }
`

const HTMLTemplate = `<html>

<head>
  <title>SSR Demo</title>
  <meta charset="utf-8" />
</head>

<body>
  <div id="root">{{SSR}}</div>
</body>
<script>
  const hydrate = (fn) => {
    if (document.readyState != 'loading') {
      fn();
    } else {
      document.addEventListener('DOMContentLoaded', fn);
    }
  }

  hydrate(function () {
    // {{__HYDRATION__}}
  })

</script>

</html>`

const renderer = `import React from 'react';
import ReactDOM from 'react-dom';

const _page = ({component, props}) => {
    return React.createElement(component, props);    
}

const render = (component, props) => {
    ReactDOM.hydrate(_page({component, props}), document.getElementById('root'));
}

export default render;`
