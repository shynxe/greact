# greact
Simple React Server-Side Rendering library written in Golang

## how to use
1. install greact as a go binary

```
go get github.com/shynxe/greact
go install github.com/shynxe/greact
```

2. initialize a greact project

```greact init```

3. set up an http server to render the react pages (```main.go```) and **install its dependencies (gin)**

```
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shynxe/greact/config"
	"github.com/shynxe/greact/renderer"
)

func main() {
	r := gin.Default()

	config.LoadConfig("greact.env")

	r.GET("/index", func(c *gin.Context) {
		props := map[string]interface{}{
			"name": "World",
		}

		html := renderer.RenderPage("index", props)

		c.Writer.Header().Set("Content-Type", "text/html")
		c.String(http.StatusOK, html)
	})

	// serve static files
	r.Static("/public", config.StaticPath)

	r.Run("localhost:8080")
}
```
4. build & run

```greact dev```

