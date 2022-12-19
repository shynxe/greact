package config

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	DefaultConfigFileName = "greact.env"
)

var (
	configFileName string
)

func InitConfig() {
	setConfigFileName()
	setConfigDefaults()
	getUserInput()
	createConfigFile()
}

func createConfigFile() {
	err := viper.WriteConfig()

	if err != nil {
		fmt.Println("Error creating config file: " + err.Error())
		return
	}

	fmt.Println("Config file created: " + configFileName)
}

func getUserInput() {
	var config = Config{}

	fmt.Print("Enter the path to the client directory [default: " + viper.GetString("clientPath") + "]: ")
	fmt.Scanln(&config.ClientPath)
	if config.ClientPath != "" {
		viper.Set("clientPath", config.ClientPath)
	}

	fmt.Print("Enter the source folder name [default: " + viper.GetString("sourceFolder") + "]: ")
	fmt.Scanln(&config.SourceFolder)
	if config.SourceFolder != "" {
		viper.Set("sourceFolder", config.SourceFolder)
	}

	fmt.Print("Enter the build folder name [default: " + viper.GetString("buildFolder") + "]: ")
	fmt.Scanln(&config.BuildFolder)
	if config.BuildFolder != "" {
		viper.Set("buildFolder", config.BuildFolder)
	}

	fmt.Print("Enter the static files folder name [default: " + viper.GetString("staticFolder") + "]: ")
	fmt.Scanln(&config.StaticFolder)
	if config.StaticFolder != "" {
		viper.Set("staticFolder", config.StaticFolder)
	}

	fmt.Print("Enter the public path [default: " + viper.GetString("publicPath") + "]: ")
	fmt.Scanln(&config.PublicPath)
	if config.PublicPath != "" {
		viper.Set("publicPath", config.PublicPath)
	}

}

func setConfigDefaults() {
	viper.SetDefault("clientPath", "./client")
	viper.SetDefault("sourceFolder", "pages")
	viper.SetDefault("buildFolder", "build")
	viper.SetDefault("staticFolder", "static")
	viper.SetDefault("publicPath", "/public/")
}

func setConfigFileName() {
	fmt.Print("Enter the name of the config file [default: " + DefaultConfigFileName + "]: ")
	fmt.Scanln(&configFileName)

	if configFileName == "" {
		configFileName = DefaultConfigFileName
	}

	viper.SetConfigFile(configFileName)
	viper.SetConfigType("dotenv")
}
