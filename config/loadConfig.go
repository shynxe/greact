package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	ClientPath   string `json:"clientPath"`
	SourceFolder string `json:"sourceFolder"`
	BuildFolder  string `json:"buildFolder"`
	StaticFolder string `json:"staticFolder"`
	PublicPath   string `json:"publicPath"`
}

var config = Config{}

var (
	BuildPath  string
	SourcePath string
	IsLoaded   bool
)

func LoadConfig(path string) error {
	// read the config file
	viper.SetConfigFile(path)
	err := viper.ReadInConfig()

	if err != nil {
		return err
	}

	// get the config values
	err = viper.Unmarshal(&config)

	if err != nil {
		return err
	}

	// set the build and source paths
	BuildPath = config.ClientPath + "/" + config.BuildFolder
	SourcePath = config.ClientPath + "/" + config.SourceFolder

	IsLoaded = true

	return nil
}

func GetConfig() Config {
	return config
}
