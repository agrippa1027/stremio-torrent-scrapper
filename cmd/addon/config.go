package main

import (
	"os"
	"scrapper/internal/addon"

	"gopkg.in/yaml.v3"
)

func loadConfigFile(path string, cfg *addon.Config) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	_ = yaml.Unmarshal(data, cfg)
}
