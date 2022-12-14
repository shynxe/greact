package renderer

import (
	"encoding/json"
	"io/ioutil"
	"log"
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

	// replace the script tag
	html = strings.Replace(
		html,
		"// {{__HYDRATION__}}",
		`
		render.default(`+page+`.default, `+string(jsonData)+`);
		`,
		1,
	)

	// replace the {{SSR}} tag with pre-rendered html
	html = strings.Replace(
		html,
		"{{SSR}}",
		`<div id="root">Hello World</div>`,
		1,
	)

	return html
}
