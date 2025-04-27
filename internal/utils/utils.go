package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func GetLogFile(path string) (*os.File, error) {
	const op = "utils.GetLogFile()"

	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func NormalizeInput(values map[string]map[string]string) (map[string]string, error) {
	m := map[string]string{}

	for k, v := range values {
		for t, c := range v {
			switch c {
			case "lowercase":
				m[k] = strings.ToLower(t)

			case "uppercase":
				m[k] = strings.ToUpper(t)

			case "title":
				runes := []rune(t)
				runes[0] = unicode.ToUpper(runes[0])

				for i := 1; i < len(runes); i++ {
					runes[i] = unicode.ToLower(runes[i])
				}

				m[k] = string(runes)

			default:
				return nil, fmt.Errorf("not supported case")
			}
		}
	}

	return m, nil
}
