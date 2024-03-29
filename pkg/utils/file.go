package utils

import (
	"os"
	"path/filepath"
)

func ReadFile(filename string) ([]byte, error) {
	if !filepath.IsAbs(filename) {
		filename = filepath.Join("images", filename)
	}

	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return fileContent, nil
}
