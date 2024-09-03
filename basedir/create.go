package basedir

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CreateConfigFile will create or truncate a config file in the first directory it can.
// It will attempt to create any directory necessary.
// E.g. suffix "sway/config" will first try to create $XDG_CONFIG_HOME/sway/config and its
// subdirectories before moving on to $XDG_CONFIG_DIRS.
// If all options are exhausted and the file could not be created, an error is returned.
// Directories are created with 0o700 permissions as per the basedir spec.
func CreateConfigFile(suffix string) (*os.File, string, error) {
	return createFileAt(suffix, ConfigHome, ConfigDirs)
}

// CreateSystemConfigFile creates or truncates a config file in the first config dir it can that is
// not under $HOME.
// It will attempt to create any directory necessary.
// E.g. suffix "sway/config" will to create sway/config and its subdirectories for every
// $XDG_CONFIG_DIRS until it succeeds.
// If all options are exhausted and the file could not be created, an error is returned.
// Directories are created with 0o700 permissions as per the basedir spec.
func CreateSystemConfigFile(suffix string) (*os.File, string, error) {
	return createSystemFileAt(suffix, ConfigDirs)
}

// CreateDataFile will create or truncate a data file in the first directory it can.
// It will attempt to create any directory necessary.
// E.g. suffix "xorg/Xorg.0.log" will first try to create $XDG_DATA_HOME/xorg/Xorg.0.log and its
// subdirectories before moving on to $XDG_DATA_DIRS.
// If all options are exhausted and the file could not be created, an error is returned.
// Directories are created with 0o700 permissions as per the basedir spec.
func CreateDataFile(suffix string) (*os.File, string, error) {
	return createFileAt(suffix, DataHome, DataDirs)
}

// CreateSystemDataFile creates or truncates a data file in the first data dir it can that is
// not under $HOME.
// It will attempt to create any directory necessary.
// E.g. suffix "xorg/Xorg.0.log" will to create xorg/Xorg.0.log and its subdirectories for every
// $XDG_DATA_DIRS until it succeeds.
// If all options are exhausted and the file could not be created, an error is returned.
// Directories are created with 0o700 permissions as per the basedir spec.
func CreateSystemDataFile(suffix string) (*os.File, string, error) {
	return createSystemFileAt(suffix, DataDirs)
}

// createFileAt attempts to create file $dir/$suffix and its subdirectories if needed.
// First, it tries to create the file in the primary dir, falling back on the secondary directories.
// The first successfully created file is returned.
// Directories are created with 0o700 permissions as per the basedir spec.
// primary can be left empty.
func createFileAt(suffix string, primary string, secondaries []string) (*os.File, string, error) {
	var err error
	if primary != "" {
		file, path, err := createFile(primary, suffix)
		if err == nil {
			return file, path, err
		}
	}

	for _, dir := range secondaries {
		file, path, err2 := createFile(dir, suffix)
		if err2 == nil {
			return file, path, err2
		}

		err = errors.Join(err, err2)
	}

	return nil, "", err
}

// createSystemFileAt attempts to create file $dir/$suffix and its subdirectories if needed.
// Dirs that are under $HOME are ignored.
// The first successfully created file is returned.
// Directories are created with 0o700 permissions as per the basedir spec.
func createSystemFileAt(suffix string, dirs []string) (*os.File, string, error) {
	var err error

	for _, dir := range dirs {
		if strings.HasPrefix(dir, Home) {
			continue
		}

		file, path, err2 := createFile(dir, suffix)
		if err2 == nil {
			return file, path, err2
		}

		err = errors.Join(err, err2)
	}

	return nil, "", err
}

// createFile attempts to create or truncate the file $dir/$suffix.
// Subdirectories are created if needed using mode 0o700.
func createFile(prefix string, suffix string) (*os.File, string, error) {
	path := filepath.Join(prefix, suffix)
	dir := filepath.Dir(path)

	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return nil, "", err
	}

	file, err := os.Create(path)

	if err != nil {
		return file, path, fmt.Errorf("createFile: failed to create file %s: %w", path, err)
	}

	return file, path, nil
}
