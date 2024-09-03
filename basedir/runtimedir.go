package basedir

import "os"

func CreateRuntimeDir() (string, error) {
	dir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		return "", err
	}

	RuntimeDir = dir

	return dir, nil
}
