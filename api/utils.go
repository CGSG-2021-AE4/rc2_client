package api

import (
	"encoding/json"
	"os"
)

// Errors

func (e rcError) Error() string {
	return e.err
}

func NewError(msg string) rcError {
	return rcError{
		err: msg,
	}
}

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
