package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shynxe/greact/build"
	"github.com/shynxe/greact/config"
)

const (
	Version = "0.0.1"
)

var (
	showVersion bool
)

func main() {
	parseArgs()

	if showVersion {
		fmt.Println("version: " + Version)
		return
	}

	if len(os.Args) < 2 {
		flag.Usage()
		return
	}

	command := os.Args[1]

	if command != "" {
		switch command {
		case "build":
			build.Build(os.Args[2:])
		case "run":
			build.Run()
		case "init":
			config.InitConfig()
		case "help":
			flag.Usage()
		default:
			fmt.Println("unknown command:", command)
		}
	} else {
		flag.Usage()
	}
}

func parseArgs() {
	flag.BoolVar(&showVersion, "version", false, "show version")

	flag.Usage = func() {
		printLogo()
		fmt.Println()
		fmt.Println("usage: greact [options] [command]")
		fmt.Println()
		fmt.Println("options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("commands:")
		fmt.Println("  build\t\tbuild the react pages")
		fmt.Println("  run\t\tstart the server")
		fmt.Println("  init\t\tinitialize the config file")
		fmt.Println("  help\t\tshow this help")
	}

	flag.Parse()
}

func printLogo() {
	fmt.Print(`
          ____                  __ 
   ____ _/ __ \___  ____ ______/ /_
  / __ ` + "`" + `/ /_/ / _ \/ __ ` + "`" + `/ ___/ __/
 / /_/ / _, _/  __/ /_/ / /__/ /_  
 \__, /_/ |_|\___/\__,_/\___/\__/  
/____/                             
`)
}
