package config

import (
	"fmt"
)

func ValidateConfig(config Config) error {
	if config.ClientPath == "" {
		return fmt.Errorf("clientPath is empty")
	}

	if config.SourceFolder == "" {
		return fmt.Errorf("sourceFolder is empty")
	}

	if config.BuildFolder == "" {
		return fmt.Errorf("buildFolder is empty")
	}

	if config.StaticFolder == "" {
		return fmt.Errorf("staticFolder is empty")
	}

	if config.PublicPath == "" {
		return fmt.Errorf("publicPath is empty")
	}

	return nil
}
