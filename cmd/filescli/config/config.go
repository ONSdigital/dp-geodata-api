package config

import (
	"os"

	"github.com/go-yaml/yaml"
)

type Config struct {
	Hosts struct {
		Files              string `yaml:"files,omitempty"`
		Zebedee            string `yaml:"zebedee,omitempty"`
		DownloadPublishing string `yaml:"download-publishing,omitempty"`
		DownloadWeb        string `yaml:"download-web,omitempty"`
		Identity           string `yaml:"identity,omitempty"`
		Upload             string `yaml:"upload,omitempty"`
	} `yaml:"hosts,omitempty"`
}

func Load(name string) (*Config, error) {
	content, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(content, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
