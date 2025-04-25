package utils

import (
	"fmt"
	"os"
)

func GetLogFile(path string) (*os.File, error) {
	const op = "utils.GetLogFile()"

	var f *os.File
	var err error

	f, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		return nil, fmt.Errorf("%s @ %v", op, err)
	}
	return f, nil
}
