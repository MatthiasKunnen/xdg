package basedir

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindConfigFile finds the given suffix in order of priority. First, XDG_CONFIG_HOME is checked,
// then, each dir in XDG_CONFIG_DIRS is checked.
// Example for suffix: sway/config.
func FindConfigFile(suffix string) (string, error) {
	return findFile(suffix, ConfigHome, ConfigDirs)
}

// FindDataFile finds the given suffix in order of priority. First, XDG_DATA_HOME is checked,
// then, each dir in XDG_DATA_DIRS is checked.
// Example for suffix: xorg/Xorg.0.log.
func FindDataFile(suffix string) (string, error) {
	return findFile(suffix, DataHome, DataDirs)
}

func findFile(suffix string, primary string, secondary []string) (string, error) {
	primaryPath := filepath.Join(primary, suffix)
	_, err := os.Stat(primaryPath)
	switch {
	case err == nil:
		return primaryPath, nil
	case os.IsNotExist(err):
	case err != nil:
		return "", fmt.Errorf("failed to stat %s: %w", primaryPath, err)
	}

	for _, path := range secondary {
		secondaryPath := filepath.Join(path, suffix)
		_, err := os.Stat(secondaryPath)
		switch {
		case err == nil:
			return secondaryPath, nil
		case os.IsNotExist(err):
		case err != nil:
			return "", fmt.Errorf("failed to stat %s: %w", secondaryPath, err)
		}
	}

	return "", nil
}
