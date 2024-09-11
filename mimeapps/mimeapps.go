package mimeapps

import (
	"errors"
	"github.com/MatthiasKunnen/xdg/basedir"
	"github.com/MatthiasKunnen/xdg/desktop"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// ListLocation holds information of a mimeapps.list file.
type ListLocation struct {
	// The path of the mimeapps.list file.
	Path string

	// HasDesktopFiles states whether there are any .desktop files to be found in the same directory
	// as the path.
	HasDesktopFiles bool
}

// GetLists returns all mimeapps.list files in accordance to freedesktop.org's
// [MIME Application Spec]. Existence of these files is not checked.
// The order is according to the priority, higher priority first.
//
// When desktop is non-empty, files such as $desktop-mimeapps.list are included.
// The value of desktop can be fetched from $XDG_CURRENT_DESKTOP.
//
// [MIME Application Spec]: https://specifications.freedesktop.org/mime-apps-spec/1.0.1/file.html
func GetLists(desktop string) []ListLocation {
	result := make([]ListLocation, 0)

	desktop = strings.ToLower(desktop)

	addMimeappsList(&result, basedir.ConfigHome, desktop, "", false)
	addMimeappsLists(&result, basedir.ConfigDirs, desktop, "", false)
	addMimeappsList(&result, basedir.DataHome, desktop, "applications", true)
	addMimeappsLists(&result, basedir.DataDirs, desktop, "applications", true)

	return result
}

func addMimeappsLists(
	list *[]ListLocation,
	paths []string,
	desktop string,
	subDir string,
	hasDesktopFiles bool,
) {
	for _, s := range paths {
		addMimeappsList(list, s, desktop, subDir, hasDesktopFiles)
	}
}

func addMimeappsList(
	list *[]ListLocation,
	envValue string,
	desktop string,
	subDir string,
	hasDesktopFiles bool,
) {
	if subDir != "" {
		envValue = filepath.Join(envValue, subDir)
	}

	if desktop != "" {
		*list = append(*list, ListLocation{
			Path:            filepath.Join(envValue, subDir, desktop+"-mimeapps.list"),
			HasDesktopFiles: hasDesktopFiles,
		})
	}

	*list = append(*list, ListLocation{
		Path:            filepath.Join(envValue, subDir, "mimeapps.list"),
		HasDesktopFiles: hasDesktopFiles,
	})
}

// GetDefaults returns the desktop IDs of each MIME type in the [Default Applications] section of
// the mimeapps.list.
// See [MIME apps spec].
//
// `associations` is needed to determine a valid association as required by the spec.
// Use [GetAssociations].
//
// desktopIdToPathsMap is used to look up the paths of a desktop file by its ID.
// The value of this parameter can be obtained using [desktop.GetDesktopFiles]
// If it is nil, the filesystem will be scanned for the desktop file.
//
// [MIME apps spec]: https://specifications.freedesktop.org/mime-apps-spec/1.0.1/default.html
func GetDefaults(
	mimeappsFileList []ListLocation,
	associations Associations,
	desktopIdToPathsMap desktop.IdPathMap,
) map[string][]string {
	result := make(map[string][]string)

	for _, location := range mimeappsFileList {
		path := location.Path
		file, err := os.Open(path)
		switch {
		case errors.Is(err, os.ErrNotExist):
			continue
		case err != nil:
			log.Printf("Error opening mimeapps file '%s': %v\n", path, err)
			continue
		}

		parsed, err := Parse(file)
		file.Close()
		if err != nil {
			log.Printf("Failed to parse mimeapps file '%s': %v\n", path, err)
			continue
		}

		for mimeType, desktopIds := range parsed.Default {
			for _, desktopId := range desktopIds {
				var dfPath string
				var dfParseError error
				if desktopIdToPathsMap == nil {
					_, dfPath, dfParseError = desktop.LoadById(desktopId, nil)
				} else {
					_, dfPath, dfParseError = desktopIdToPathsMap.LoadById(desktopId)
				}

				if dfPath == "" {
					continue
				}

				if dfParseError != nil {
					log.Printf("Failed to parse desktop file with ID '%s': %v\n", path, dfParseError)
					continue
				}

				if associations[mimeType] == nil || !slices.Contains(associations[mimeType], desktopId) {
					// If a valid desktop file is found, verify that it is associated with the type
					log.Printf(
						"Mimeapps file %s states %s as default application for mimetype %s "+
							"but the mime type is not in any [Added Associations] section.\n",
						path,
						desktopId,
						mimeType,
					)
					continue
				}

				if result[mimeType] == nil {
					result[mimeType] = []string{desktopId}
				} else {
					result[mimeType] = append(result[mimeType], desktopId)
				}

			}
		}
	}

	return result
}

// Associations is a map of Key=MIME type, Value=List of desktop IDs.
// It can be used to look up all the desktop IDs that support opening a certain MIME type.
type Associations = map[string][]string

// GetAssociations returns all mime-desktop associations created by entries in the
// [Added Associations] and [Remove Associations] sections and the MimeType in the .desktop files.
//
// The following part of the mime apps spec is not implemented:
//
// If the addition or removal refers to a desktop file that doesn't exist at this precedence
// level, or a lower one, then the addition or removal is ignored, even if the desktop
// file exists in a high-precedence directory.
func GetAssociations(
	mimeappsLocations []ListLocation,
	idPathsMap desktop.IdPathMap,
) Associations {
	result := make(Associations)
	blacklistMimeDesktop := make(map[string]map[string]bool)
	blacklistDesktopIds := make(map[string]bool)

	// Maps the desktop ID to the index of the lowest precedence desktop file that can be found in
	// mimeappsLocations. E.g. key=foo.desktop, value=2, means that foo.desktop is next to
	// mimeappsLocations[2] and may also be in any of the higher precedence directories such as
	// mimeappsLocations[1] and mimeappsLocations[0].
	desktopIdLowestIndex := make(map[string]int)

	for desktopId, paths := range idPathsMap {
		lowestPrecedence := -1

		for i, mimeappsPath := range mimeappsLocations {
			dir := filepath.Dir(mimeappsPath.Path)

			for _, path := range paths {
				if isSubPathAbs(path, dir) {
					lowestPrecedence = i
				}
			}
		}

		desktopIdLowestIndex[desktopId] = lowestPrecedence
	}

	for i, location := range mimeappsLocations {
		path := location.Path

		if filepath.Base(path) != "mimeapps.list" {
			// mimeapps files with the format $desktop-mimeapps cannot be used to add/remove
			// associations
			continue
		}

		parsed, err := ParseFile(path)
		switch {
		case errors.Is(err, os.ErrNotExist):
			// A nonexistent mimeapps.list should be treated as an empty file.
		case err != nil:
			log.Printf("Error parsing mimeapps file '%s': %v\n", path, err)
		}

		for mime, desktopIds := range parsed.Added {
			if blacklistMimeDesktop[mime] == nil {
				blacklistMimeDesktop[mime] = make(map[string]bool)
			}

			for _, desktopId := range desktopIds {
				if blacklistDesktopIds[desktopId] {
					continue
				}

				depth, exists := desktopIdLowestIndex[desktopId]
				if !exists || depth < i {
					// If the addition or removal refers to a desktop file that doesn't exist at
					// this precedence level, or a lower one, then the addition or removal is
					// ignored, even if the desktop file exists in a high-precedence directory.
					continue
				}

				if blacklistMimeDesktop[mime][desktopId] {
					continue
				}
				blacklistMimeDesktop[mime][desktopId] = true

				if result[mime] == nil {
					result[mime] = []string{desktopId}
				} else {
					result[mime] = append(result[mime], desktopId)
				}
			}
		}

		for mime, desktopIds := range parsed.Removed {
			if blacklistMimeDesktop[mime] == nil {
				blacklistMimeDesktop[mime] = make(map[string]bool)
			}

			for _, desktopId := range desktopIds {
				if blacklistDesktopIds[desktopId] {
					continue
				}

				depth, exists := desktopIdLowestIndex[desktopId]
				if !exists || depth < i {
					// If the addition or removal refers to a desktop file that doesn't exist at
					// this precedence level, or a lower one, then the addition or removal is
					// ignored, even if the desktop file exists in a high-precedence directory.
					continue
				}

				blacklistMimeDesktop[mime][desktopId] = true
			}
		}

		if !location.HasDesktopFiles {
			continue
		}

		// Add to the results list any .desktop file found in the same directory as the
		// mimeapps.list which lists the given type in its MimeType= line, excluding any
		// desktop files already in the blacklist.
		dirname := filepath.Dir(path)
		// Needed for stable output
		toAdd := make(map[string][]string)
		for desktopId, paths := range idPathsMap {
			if blacklistDesktopIds[desktopId] {
				continue
			}

			for _, desktopFilePath := range paths {
				if !isSubPathAbs(desktopFilePath, dirname) {
					continue
				}

				if blacklistDesktopIds[desktopId] {
					continue
				}
				blacklistDesktopIds[desktopId] = true

				entry, err := desktop.ParseFile(desktopFilePath)
				if err != nil {
					log.Printf("Failed to load desktop file '%s', skipping: %v\n", desktopFilePath, err)
					continue
				}

				for _, mime := range entry.MimeType {
					if blacklistMimeDesktop[mime][desktopId] {
						continue
					}

					toAdd[mime] = append(toAdd[mime], desktopId)

					if blacklistMimeDesktop[mime] == nil {
						blacklistMimeDesktop[mime] = make(map[string]bool)
					}
					blacklistMimeDesktop[mime][desktopId] = true
				}
			}
		}

		for mime, desktopIds := range toAdd {
			slices.Sort(desktopIds)
			if result[mime] == nil {
				result[mime] = desktopIds
			} else {
				result[mime] = append(result[mime], desktopIds...)
			}
		}
	}

	return result
}

// GetPreferredApplications returns the preferred applications for each supported mime type based
// on the mimeapps.list files.
// Applications are ordered with higher priority first. Default applications are listed first.
// This is a combination of [GetAssociations] and [GetDefaults].
func GetPreferredApplications(
	mimeappsFileList []ListLocation,
	desktopIdPathMap desktop.IdPathMap,
) Associations {
	associations := GetAssociations(mimeappsFileList, desktopIdPathMap)
	defaults := GetDefaults(mimeappsFileList, associations, desktopIdPathMap)

	for mime, desktopIds := range defaults {
		if associations[mime] == nil {
			associations[mime] = desktopIds
		} else {
			associations[mime] = removeDuplicates(append(desktopIds, associations[mime]...))
		}
	}

	return associations
}
