package sharedmimeinfo

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/MatthiasKunnen/xdg/basedir"
	"io"
	"os"
	"path"
	"slices"
	"strings"
)

const (
	mimeTextPlain = "text/plain"
	mimeOctet     = "application/octet-stream"
)

type MalformedSubclassError struct {
	FileIndex int
	LineIndex int
}

func (e MalformedSubclassError) Error() string {
	return fmt.Sprintf(
		"malformed subclass line at %d",
		e.LineIndex,
	)
}

type Subclass struct {
	dict map[string][]string
}

// LoadFromOs loads the subclasses files according to both the shared-mime-info spec and
// the basedir spec.
// XDG_DATA_HOME and XDG_DATA_DIRS are retrieved from the environment.
func LoadFromOs() (*Subclass, error) {
	var dirs []string
	dirs = append(dirs, basedir.DataHome)
	dirs = append(dirs, basedir.DataDirs...)
	var files []*os.File
	var readers []io.Reader

	for _, dir := range dirs {
		fPath := path.Join(dir, "mime/subclasses")
		f, err := os.Open(fPath)
		switch {
		case errors.Is(err, os.ErrNotExist):
			continue
		case err != nil:
			return nil, fmt.Errorf("failed to load subclasses file at %s: %w", fPath, err)
		default:
			files = append(files, f)
			readers = append(readers, f)
		}
	}

	defer func() {
		for _, f := range files {
			_ = f.Close()
		}
	}()

	subclasses, err := LoadFromReaders(readers)
	if err == nil {
		return subclasses, nil
	}
	var x MalformedSubclassError
	if errors.As(err, &x) && x.FileIndex >= 0 && x.FileIndex < len(files) {
		return nil, fmt.Errorf(
			"failed to load subclass file %s: %w",
			files[x.FileIndex].Name(),
			err,
		)
	}

	return nil, err
}

// LoadFromReaders loads the subclasses based on the given [io.Reader] slice.
// Order is important as earlier readers have higher precedence.
func LoadFromReaders(readers []io.Reader) (*Subclass, error) {
	mimeSubclass := &Subclass{
		dict: make(map[string][]string),
	}

	for fileIndex, f := range readers {
		scanner := bufio.NewScanner(f)
		lineIndex := 0
		for scanner.Scan() {
			line := scanner.Text()
			specific, broad, found := strings.Cut(line, " ")
			if !found {
				return nil, MalformedSubclassError{
					FileIndex: fileIndex,
					LineIndex: lineIndex,
				}
			}

			if broadList, ok := mimeSubclass.dict[specific]; ok {
				if !slices.Contains(broadList, broad) {
					mimeSubclass.dict[specific] = append(broadList, broad)
				}
			} else {
				mimeSubclass.dict[specific] = []string{broad}
			}
			lineIndex++
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	return mimeSubclass, nil
}

// BroaderOnce returns the direct subclasses of the given MIME type.
// For example, text/javascript returns application/x-executable.
func (s *Subclass) BroaderOnce(mime string) []string {
	broad := s.dict[mime]
	if len(broad) > 0 {
		return broad
	}

	switch {
	case mime == mimeOctet:
		return nil
	case mime == mimeTextPlain:
		return []string{mimeOctet}
	case strings.HasPrefix(mime, "text/"):
		return []string{mimeTextPlain}
	case !strings.HasPrefix(mime, "inode/"):
		return []string{mimeOctet}
	default:
		return nil
	}
}

// BroaderDfs returns all subclasses of the given MIME type.
// The order of the subclasses is priority first and is determined by a depth first,
// pre-order (NLR), search.
// For example, text/javascript returns application/x-executable, text/plain,
// application/octet-stream.
func (s *Subclass) BroaderDfs(mime string) []string {
	visited := make(map[string]struct{})
	toVisit := s.dict[mime]
	result := make([]string, 0, len(toVisit))

	for len(toVisit) > 0 {
		broad := toVisit[0]
		if _, ok := visited[broad]; ok {
			toVisit = toVisit[1:]
			continue
		}

		visited[broad] = struct{}{}
		result = append(result, broad)
		broader := s.dict[broad]
		switch len(broader) {
		case 0:
			toVisit = toVisit[1:]
		case 1:
			toVisit[0] = broader[0]
		default:
			toVisit = append(broader, toVisit[1:]...)
		}
	}

	if _, ok := visited[mimeTextPlain]; !ok {
		for _, item := range result {
			if strings.HasPrefix(item, "text/") {
				result = append(result, mimeTextPlain)
				break
			}
		}
	}
	if _, ok := visited[mimeOctet]; !ok {
		for _, item := range result {
			if !strings.HasPrefix(item, "inode/") {
				result = append(result, mimeOctet)
				break
			}
		}
	}

	return result
}
