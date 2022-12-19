package renderer

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

	"github.com/shynxe/greact/config"
)

func RenderPage(page string, props interface{}) string {
	file, err := ioutil.ReadFile(config.StaticPath + "/" + page + ".html")
	if err != nil {
		log.Fatal(err)
	}

	jsonData, err := json.Marshal(props)
	if err != nil {
		log.Fatal(err)
	}

	html := string(file)

	// get the rendered html from the page component

	// node -e "const page=require('./build/render.js');console.log(page.default('index', {name: 'test'}));"
	// run the node command to get the rendered html
	cmd := exec.Command("node", "-e", "const page=require('"+config.BuildPath+"/render.js');console.log(page.default('"+page+"', "+string(jsonData)+"));")
	stdout, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	// replace the script tag
	html = strings.Replace(
		html,
		"// {{__HYDRATION__}}",
		`
		hydrate.default(`+page+`.default, `+string(jsonData)+`);
		`,
		1,
	)

	// replace the {{SSR}} tag with pre-rendered html
	html = strings.Replace(
		html,
		"{{SSR}}",
		string(stdout),
		1,
	)

	return html
}
