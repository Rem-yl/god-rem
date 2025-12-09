package code

import (
	"encoding/json"
	"fmt"
	"os"
)

func ReadConfig(filename string) (string, error) {
	// 1. 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return "", fmt.Errorf("file: %s not exist: %w", filename, err)
	}

	// 读取文件
	data, err := os.ReadFile(filename)

	if err != nil {
		return "", fmt.Errorf("read file: %s error: %w", filename, err)
	}

	return string(data), nil
}

type Config struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

func validateConfig(config *Config) error {
	if config.Port < 0 || config.Port > 65536 {
		return fmt.Errorf("port: %d not in range [0, 65535]", config.Port)
	}

	return nil
}

func LoadConfig(filename string) (*Config, error) {
	// 1. 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("file: %s not exist: %w", filename, err)
	}

	// 2. 读取文件
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file: %s error: %w", filename, err)
	}

	var config Config
	// 3. 解析文件
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("file: %s load json error: %w", filename, err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("validate config error: %w", err)
	}

	return &config, nil
}
