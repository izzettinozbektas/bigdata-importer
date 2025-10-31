package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
}

type Config struct {
	Database DatabaseConfig `yaml:"database"`
}

func LoadConfig(filename string) (*Config, error) {
	paths := []string{
		filename,                      // current dir
		filepath.Join("..", filename), // one level up
		filepath.Join("..", "..", filename),
		filepath.Join("config", filename),
		"/app/config.yaml", // Docker default path
	}

	var data []byte
	var err error
	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil {
			fmt.Printf("Config loaded from: %s\n", p)
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("config.yaml not found in paths: %v", paths)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
