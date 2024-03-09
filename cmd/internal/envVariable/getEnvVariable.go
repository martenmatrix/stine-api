package envVariable

import (
	"os"
)

func EnvVariable(key string) (string, error) {

	err := os.Setenv(key, "gopher")

	if err != nil {
		return "", err
	}

	return os.Getenv(key), nil
}
