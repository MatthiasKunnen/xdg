package mimeapps

import (
	"os"
	"strings"
)

// removeDuplicates removes duplicates entries from a slice and returns the slice.
// Order is preserved and the first occurrence of every entry is preserved.
// If input is nil, nil is returned.
func removeDuplicates[T comparable](input []T) []T {
	if input == nil {
		return nil
	}

	seen := make(map[T]bool, len(input))
	list := make([]T, 0, len(input))

	for _, item := range input {
		if !seen[item] {
			seen[item] = true
			list = append(list, item)
		}
	}

	return list
}

// isSubPathAbs returns true if sub is a sub path of parent.
// Both parent and sub must be absolute paths.
func isSubPathAbs(sub string, parent string) bool {
	return strings.HasPrefix(sub+string(os.PathSeparator), parent+string(os.PathSeparator))
}
