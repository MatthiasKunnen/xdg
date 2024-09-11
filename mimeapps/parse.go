package mimeapps

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// MimeApps represents a parsed mimeapps.list file.
// The structure is defined in
// https://specifications.freedesktop.org/mime-apps-spec/1.0.1/index.html
type MimeApps struct {
	Default map[string][]string
	Added   map[string][]string
	Removed map[string][]string
}

const (
	addToNone = iota
	addToDefault
	addToAdded
	addToRemoved
)

func Parse(reader io.Reader) (MimeApps, error) {
	sc := bufio.NewScanner(reader)
	result := MimeApps{}
	result.Default = make(map[string][]string)
	result.Added = make(map[string][]string)
	result.Removed = make(map[string][]string)
	var status int

	for sc.Scan() {
		line := sc.Text()
		switch line {
		case "":
			continue
		case "[Default Applications]":
			status = addToDefault
			continue
		case "[Added Associations]":
			status = addToAdded
			continue
		case "[Removed Associations]":
			status = addToRemoved
			continue
		}

		if status == addToNone {
			continue
		}

		split := strings.SplitN(line, "=", 2)
		switch len(split) {
		case 1:
			continue // Lines without = are ignored. This is the same behavior as xdg-open.
		case 2:
		default:
			return MimeApps{}, fmt.Errorf("parse mimeapps: expected mimetype=.desktop: %s", line)
		}

		mimeType := split[0]
		apps := strings.Split(strings.TrimSuffix(split[1], ";"), ";")

		switch status {
		case addToDefault:
			if result.Default[mimeType] == nil {
				result.Default[mimeType] = apps
			} else {
				result.Default[mimeType] = append(result.Default[mimeType], apps...)
			}
			result.Default[mimeType] = apps
		case addToAdded:
			if result.Added[mimeType] == nil {
				result.Added[mimeType] = apps
			} else {
				result.Added[mimeType] = append(result.Added[mimeType], apps...)
			}
		case addToRemoved:
			if result.Removed[mimeType] == nil {
				result.Removed[mimeType] = apps
			} else {
				result.Removed[mimeType] = append(result.Removed[mimeType], apps...)
			}
		}

	}

	if err := sc.Err(); err != nil {
		return MimeApps{}, fmt.Errorf("failed to parse: %w", err)
	}

	return result, nil
}

func ParseFile(path string) (MimeApps, error) {
	file, err := os.Open(path)
	if err != nil {
		return MimeApps{}, err
	}

	return Parse(file)
}
