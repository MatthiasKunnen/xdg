// Package basedir contains all environment variables as specified by the
// [XDG Base Directory Specification]
//
// [XDG Base Directory Specification]: https://specifications.freedesktop.org/basedir-spec/0.8/
package basedir

import (
	"os"
	"path/filepath"
	"strings"
)

var (
	// CacheHome is the single base directory relative to which user-specific non-essential (cached)
	// data should be written. This directory is defined by the environment variable
	// $XDG_CACHE_HOME.
	CacheHome string

	// ConfigHome is the single base directory relative to which user-specific configuration files
	// should be written. This directory is defined by the environment variable $XDG_CONFIG_HOME.
	ConfigHome string

	// ConfigDirs is a set of preference ordered base directories relative to which configuration
	// files should be searched. This set of directories is defined by the environment
	// variable $XDG_CONFIG_DIRS.
	ConfigDirs []string

	// DataHome is a single base directory relative to which user-specific data files should be
	// written. This directory is defined by the environment variable $XDG_DATA_HOME.
	DataHome string

	// DataDirs is a set of preference ordered base directories relative to which data files should
	// be searched. This set of directories is defined by the environment variable $XDG_DATA_DIRS.
	DataDirs []string

	// Home is the equivalent of $HOME. It will always be non-empty.
	Home string

	// LocalBin is a single base directory relative to which user-specific executable files may be
	// written.
	LocalBin string

	// RuntimeDir is a single base directory relative to which user-specific runtime files and other
	// file objects should be placed. This directory is defined by the environment variable
	// $XDG_RUNTIME_DIR.
	RuntimeDir string

	// StateHome is a single base directory relative to which user-specific state data should be
	// written. This directory is defined by the environment variable $XDG_STATE_HOME.
	StateHome string
)

func init() {
	Reinit()
}

// Reinit reinitializes the basedir values. Use this if you change XDG environment variables.
func Reinit() {
	home := os.Getenv("HOME")
	if home == "" {
		// $HOME must always be set in a POSIX environment.
		panic("$HOME environment variable not set")
	}

	CacheHome = singleVar("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	ConfigHome = singleVar("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	ConfigDirs = listVar("XDG_CONFIG_DIRS", []string{"/etc/xdg"})
	DataDirs = listVar("XDG_DATA_DIRS", []string{"/usr/local/share/", "/usr/share/"})
	DataHome = singleVar("XDG_DATA_HOME", filepath.Join(home, ".local/share"))
	Home = home
	LocalBin = filepath.Join(home, ".local/bin")
	RuntimeDir = singleVar("XDG_RUNTIME_DIR", "")
	StateHome = singleVar("XDG_STATE_HOME", filepath.Join(home, ".local/state"))
}

func singleVar(envName string, defaultValue string) string {
	envValue := os.Getenv(envName)
	if envValue == "" || !filepath.IsAbs(envValue) {
		return defaultValue
	}

	return envValue
}

func listVar(envName string, defaultValue []string) []string {
	envValue := os.Getenv(envName)
	if envValue == "" {
		return defaultValue
	}

	result := make([]string, 0)
	for _, path := range strings.Split(envValue, ":") {
		if path == "" || !filepath.IsAbs(path) {
			continue
		}

		result = append(result, path)
	}

	if len(result) == 0 {
		return defaultValue
	}

	return result
}
