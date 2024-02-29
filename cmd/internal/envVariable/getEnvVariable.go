package envVariable

import (
	"os"
)

func envVariable(key string) (string, error) {

	err := os.Setenv(key, "gopher")

	if err != nil {
		return "", err
	}

	return os.Getenv(key), nil
}
