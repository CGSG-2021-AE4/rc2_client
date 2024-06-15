package api

import (
	"encoding/json"
	"os"
)

// Config
func LoadConfig(filename string) (*Config, error) {
	var config Config

	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Error implementation - found in Effective Go
type rcError string

func (err rcError) Error() string {
	return string(err)
}
