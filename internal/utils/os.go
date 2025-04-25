package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetLogFile(path string) (*os.File, error) {
	const op = "utils.GetLogFile()"

	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return nil, fmt.Errorf("%s @ %v", op, err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("%s @ %v", op, err)
	}
	return f, nil
}
