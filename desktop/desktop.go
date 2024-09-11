package desktop

import (
	"errors"
	"fmt"
	"github.com/MatthiasKunnen/xdg/basedir"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// GetDirs returns all directories containing .desktop files in accordance with
// [Desktop Menu Specification].
// The order is according to the priority.
// Personal directories such as XDG_CONFIG_HOME are first.
//
// [Desktop Menu Specification]: https://specifications.freedesktop.org/menu-spec/latest/paths.html
func GetDirs() []string {
	result := make([]string, 0)

	result = append(result, filepath.Join(basedir.DataHome, "applications"))

	for _, s := range basedir.DataDirs {
		result = append(result, filepath.Join(s, "applications"))
	}

	return result
}

// IdPathMap maps a [Desktop ID], such as libreoffice-writer.desktop, to its desktop file paths,
// such as:
//   - /home/user/.local/share/applications/libreoffice-writer.desktop
//   - /usr/share/applications/libreoffice/writer.desktop
//
// [Desktop ID]: https://specifications.freedesktop.org/desktop-entry-spec/1.5/file-naming.html#desktop-file-id
type IdPathMap map[string][]string

// LoadById loads the first valid desktop file in the list of paths for the given desktop ID and
// returns the parsed result and the path to the file.
// If no valid desktop file could be found, error will be nil and path will be an empty string.
// Example of desktopId: vim.desktop
func (m IdPathMap) LoadById(desktopId string) (*Entry, string, error) {
	if m[desktopId] == nil {
		return nil, "", nil
	}

	for _, path := range m[desktopId] {
		parsed, err := LoadFile(path)
		if err != nil {
			log.Printf("Failed to load desktop ID %s: %v. Skipping\n", desktopId, err)
			continue
		}

		return parsed, path, nil
	}

	return nil, "", nil
}

// GetDesktopFiles returns a map of all desktop IDs and their respective desktop file path that
// could be found in the given locations.
// To get the standard locations, use GetDesktopFileLocations.
// The slice of desktop file paths is in order of highest to lowest precedence.
func GetDesktopFiles(locations []string) (IdPathMap, error) {
	result := make(IdPathMap)

	for _, dir := range locations {
		err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			if entry.IsDir() {
				return nil
			}

			add := false

			switch filepath.Ext(path) {
			case ".desktop":
				add = true
			case ".directory":
			default:
				isDesktopFile, magicError := MagicIsDesktopFilePath(path)
				if isDesktopFile && magicError == nil {
					add = true
				}
			}

			if add {
				desktopId := strings.ReplaceAll(
					strings.TrimPrefix(path, dir)[1:],
					string(filepath.Separator),
					"-",
				)
				if result[desktopId] == nil {
					result[desktopId] = []string{path}
				} else {
					result[desktopId] = append(result[desktopId], path)
				}
			}

			return nil
		})

		switch {
		case errors.Is(err, os.ErrNotExist):
		case err != nil:
			return result, fmt.Errorf(
				"getDesktopFiles, failed to walk dir %s for desktop files: %w",
				dir,
				err,
			)
		}
	}

	return result, nil
}

// GetDesktopFileLocations returns the directories where desktop files can be found.
// The locations are defined in the [Mime app spec].
//
// [Mime app spec]: https://specifications.freedesktop.org/mime-apps-spec/1.0.1/file.html
func GetDesktopFileLocations() []string {
	locations := make([]string, 0)
	locations = append(locations, filepath.Join(basedir.DataHome, "applications"))

	for _, baseDir := range basedir.DataDirs {
		locations = append(locations, filepath.Join(baseDir, "applications"))
	}

	return locations
}

// LoadById finds the first valid desktop file with the given ID, parses it and returns the result
// and the path of the file.
// If locations is nil, GetDesktopFileLocations will be used.
// If no valid desktop file could be found, error will be nil and path will be an empty string.
// Example of desktopId: vim.desktop
func LoadById(desktopId string, locations []string) (*Entry, string, error) {
	if locations == nil {
		locations = GetDesktopFileLocations()
	}

	for _, dir := range locations {
		attempts := map[string]bool{
			filepath.Join(dir, desktopId): true,
			// Desktop IDs with hyphens such as foo-bar.desktop can mean foo/bar.desktop
			filepath.Join(dir, strings.Replace(desktopId, "-", "/", 1)): true,
		}

		for path, _ := range attempts {
			_, err := os.Stat(path)
			switch {
			case errors.Is(err, os.ErrNotExist):
				continue
			case err != nil:
				log.Printf("Failed to stat desktop file '%s': %v\n", path, err)
				continue
			}

			parsed, err := LoadFile(path)
			if err != nil {
				log.Printf("Failed to load desktop file '%s': %v. Skipping\n", desktopId, err)
				continue
			}

			return parsed, path, nil
		}
	}

	return nil, "", nil
}

func LoadFile(path string) (*Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf(
			"LoadFile: failed to open desktop file '%s'. %w",
			path,
			err,
		)
	}

	parsed, err := Parse(file)
	file.Close()

	if err != nil {
		return nil, fmt.Errorf(
			"LoadFile: failed to parse desktop file '%s'. %w",
			path,
			err,
		)
	}

	return parsed, nil
}
