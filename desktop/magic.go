package desktop

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode"
)

const (
	magicDefault = iota
	magicToCommentEnd
)

// MagicIsDesktopFile returns true if the content is likely a desktop file.
// This can be used to do MIME checking of unknown content.
// The content is checked according to the [desktop entry format] spec.
//
// [desktop entry format]: https://specifications.freedesktop.org/desktop-entry-spec/1.5/basic-format.html
func MagicIsDesktopFile(reader io.Reader) (bool, error) {
	expectedEntry := requiredGroupHeader[1:]
	utf8BomHeader := []byte{0xEF, 0xBB, 0xBF}

	r := bufio.NewReader(reader)

	maybeBom, err := r.Peek(len(utf8BomHeader))
	if err != nil {
		return false, nil
	}

	if bytes.Equal(maybeBom, utf8BomHeader) {
		_, err := r.Discard(len(utf8BomHeader))
		if err != nil {
			return false, nil
		}
	}

	var status int

	for {
		readRune, _, err := r.ReadRune()
		switch {
		case readRune == unicode.ReplacementChar:
			// Desktop file must be UTF-8
			if status == magicToCommentEnd {
				// But nonsense in comments doesn't matter
				continue
			}
			return false, nil
		case err != nil:
			return false, nil
		}

		switch status {
		case magicDefault:
			switch readRune {
			case '#':
				status = magicToCommentEnd
				continue
			case '\n':
				continue
			case '[':
				deBuffer := make([]byte, len(expectedEntry))
				_, err := io.ReadFull(r, deBuffer)
				if err != nil {
					return false, nil
				}

				return expectedEntry == string(deBuffer), nil
			default:
				return false, nil
			}
		case magicToCommentEnd:
			if readRune == '\n' {
				status = magicDefault
			}
			continue
		}
	}
}

// MagicIsDesktopFilePath returns true if the file at the given path is likely a desktop file.
// This can be used to do MIME checking of unknown files.
// The file is checked according to the [desktop entry spec].
//
// [desktop entry spec]: https://specifications.freedesktop.org/desktop-entry-spec/1.5/basic-format.html
func MagicIsDesktopFilePath(path string) (bool, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return false, fmt.Errorf(
			"failed to open file '%s' to check if it a desktop file. %w",
			path,
			err,
		)
	}

	return MagicIsDesktopFile(file)
}
