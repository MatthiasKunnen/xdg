package mimeapps

import (
	"encoding/json"
	"fmt"
	"github.com/MatthiasKunnen/xdg/basedir"
	"github.com/MatthiasKunnen/xdg/desktop"
	"github.com/google/go-cmp/cmp"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func ExampleGetPreferredApplications() {
	mimeAppsLists := GetLists(os.Getenv("XDG_CURRENT_DESKTOP"))
	desktopFilePaths, err := desktop.GetDesktopFiles(desktop.GetDesktopFileLocations())
	if err != nil {
		log.Fatalf("Could not get desktop files: %v", err)
	}

	applications := GetPreferredApplications(mimeAppsLists, desktopFilePaths)

	for mime, paths := range applications {
		fmt.Printf("MIME type %s has the following desktop files: %s\n", mime, strings.Join(paths, ":"))
	}
}

func overrideEnv(t *testing.T, keyValue map[string]string) {
	t.Cleanup(func() {
		basedir.Reinit()
	})
	for key, value := range keyValue {
		originalValue := os.Getenv(key)
		err := os.Setenv(key, value)
		if err != nil {
			t.Errorf("Could not set environment variable %s=%v: %v", key, value, err)
		}
		t.Cleanup(func() {
			_ = os.Setenv(key, originalValue)
		})
	}
	basedir.Reinit()
}

func TestGetListsWithDesktop(t *testing.T) {
	overrideEnv(t, map[string]string{
		"XDG_CONFIG_HOME": "/home/user/.config",
		"XDG_CONFIG_DIRS": "/etc/xdg",
		"XDG_DATA_HOME":   "/home/user/.local/share",
		"XDG_DATA_DIRS":   "/usr/local/share/:/usr/share/",
	})

	actual := GetLists("gnome")
	expected := []ListLocation{
		{Path: "/home/user/.config/gnome-mimeapps.list", HasDesktopFiles: false},
		{Path: "/home/user/.config/mimeapps.list", HasDesktopFiles: false},
		{Path: "/etc/xdg/gnome-mimeapps.list", HasDesktopFiles: false},
		{Path: "/etc/xdg/mimeapps.list", HasDesktopFiles: false},
		{Path: "/home/user/.local/share/applications/gnome-mimeapps.list", HasDesktopFiles: false},
		{Path: "/home/user/.local/share/applications/mimeapps.list", HasDesktopFiles: true},
		{Path: "/usr/local/share/applications/gnome-mimeapps.list", HasDesktopFiles: false},
		{Path: "/usr/local/share/applications/mimeapps.list", HasDesktopFiles: true},
		{Path: "/usr/share/applications/gnome-mimeapps.list", HasDesktopFiles: false},
		{Path: "/usr/share/applications/mimeapps.list", HasDesktopFiles: true},
	}

	if !cmp.Equal(actual, expected) {
		t.Errorf("GetLists output does not match:\n%s", cmp.Diff(expected, actual))
	}
}

func TestGetListsWithoutDesktop(t *testing.T) {
	overrideEnv(t, map[string]string{
		"XDG_CONFIG_HOME": "/home/user/.config",
		"XDG_CONFIG_DIRS": "/etc/xdg",
		"XDG_DATA_HOME":   "/home/user/.local/share",
		"XDG_DATA_DIRS":   "/usr/local/share/:/usr/share/",
	})

	actual := GetLists("")
	expected := []ListLocation{
		{Path: "/home/user/.config/mimeapps.list", HasDesktopFiles: false},
		{Path: "/etc/xdg/mimeapps.list", HasDesktopFiles: false},
		{Path: "/home/user/.local/share/applications/mimeapps.list", HasDesktopFiles: true},
		{Path: "/usr/local/share/applications/mimeapps.list", HasDesktopFiles: true},
		{Path: "/usr/share/applications/mimeapps.list", HasDesktopFiles: true},
	}

	if !cmp.Equal(actual, expected) {
		t.Errorf("GetLists output does not match:\n%s", cmp.Diff(expected, actual))
	}
}

func getScenarioMimeapps(scenarioName string, t *testing.T) ([]ListLocation, desktop.IdPathMap) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(cwd, "testdata", scenarioName)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	locations := make([]ListLocation, len(entries))
	desktopFileDirs := make([]string, len(entries))
	for i, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		desktopFileDirs[i] = filepath.Join(dir, entry.Name())
		locations[i] = ListLocation{
			Path:            filepath.Join(dir, entry.Name(), "mimeapps.list"),
			HasDesktopFiles: true,
		}
	}

	idPathMap, err := desktop.GetDesktopFiles(desktopFileDirs)
	if err != nil {
		t.Fatal(err)
	}

	return locations, idPathMap
}

func TestGetAssociationsS01(t *testing.T) {
	mimeappsLists, idPathMap := getScenarioMimeapps("scenario01", t)
	associations := GetAssociations(mimeappsLists, idPathMap)

	expectedTextPlain := []string{"foo1.desktop"}
	actualTextPlain := associations["text/plain"]
	if !slices.Equal(expectedTextPlain, actualTextPlain) {
		t.Errorf("text/plain, expected: %v, actual: %v", expectedTextPlain, actualTextPlain)
	}

	expectedTextCsv := []string{"foo1.desktop"}
	actualTextCsv := associations["text/csv"]
	if !slices.Equal(expectedTextCsv, actualTextCsv) {
		t.Errorf("text/csv, expected: %v, actual: %v", expectedTextCsv, actualTextCsv)
	}

	if len(associations["text/html"]) > 0 {
		t.Errorf("text/html, expected no associations, actual: %v", associations["text/html"])
	}
}

func TestGetAssociationsS02(t *testing.T) {
	mimeappsLists, idPathMap := getScenarioMimeapps("scenario02", t)
	associations := GetAssociations(mimeappsLists, idPathMap)

	expectedMap := map[string][]string{
		"text/plain": {"foo1.desktop"},
		"text/html":  {"foo1.desktop"},
	}

	if len(associations) != len(expectedMap) {
		t.Errorf("len(associations) = %d, expected: %d", len(associations), len(expectedMap))
	}

	for mime, desktopIds := range associations {
		if !slices.Equal(desktopIds, expectedMap[mime]) {
			t.Errorf(
				"%s has incorrect associations. Expected: %v, actual: %v",
				mime,
				expectedMap[mime],
				desktopIds,
			)
		}
	}
}

func TestGetAssociationsS03(t *testing.T) {
	mimeappsLists, idPathMap := getScenarioMimeapps("scenario03", t)

	associations := GetAssociations(mimeappsLists, idPathMap)

	if len(associations) > 0 {
		t.Errorf("expected empty associations, got: %v", associations)
	}
}

func TestGetAssociationsS04Precedence(t *testing.T) {
	mimeappsLists, idPathMap := getScenarioMimeapps("scenario04", t)

	associations := GetAssociations(mimeappsLists, idPathMap)

	expectedAmountOfAssociations := 3
	actualAmountOfAssociations := len(associations)
	if actualAmountOfAssociations != expectedAmountOfAssociations {
		t.Errorf(
			"expected %d associations, got: %d",
			expectedAmountOfAssociations,
			actualAmountOfAssociations,
		)
	}

	expectedTextRtf := []string{"libreoffice-writer.desktop"}
	actualTextRtf := associations["text/rtf"]
	if !slices.Equal(expectedTextRtf, actualTextRtf) {
		t.Errorf("text/rtf, expected: %v, actual: %v", expectedTextRtf, actualTextRtf)
	}

	expectedTextPlain := []string{"libreoffice-writer.desktop", "firefox.desktop", "vim.desktop"}
	actualTextPlain := associations["text/plain"]
	if !slices.Equal(expectedTextPlain, actualTextPlain) {
		t.Errorf("text/plain, expected: %v, actual: %v", expectedTextPlain, actualTextPlain)
	}

	var expectedTextCsv []string
	actualTextCsv := associations["text/csv"]
	if !slices.Equal(expectedTextCsv, actualTextCsv) {
		t.Errorf("text/csv, expected: %v, actual: %v", expectedTextCsv, actualTextCsv)
	}

	expectedTextC := []string{"vim.desktop"}
	actualTextC := associations["text/x-c"]
	if !slices.Equal(expectedTextC, actualTextC) {
		t.Errorf("text/x-c, expected: %v, actual: %v", expectedTextC, actualTextC)
	}
}

func TestGetPreferredApplicationsS05Regression(t *testing.T) {
	// This test is meant to catch future regressions. Its accuracy at time of writing is unchecked.
	mimeappsLists, idPathMap := getScenarioMimeapps("scenario05", t)
	associations := GetPreferredApplications(mimeappsLists, idPathMap)

	expectedFilePath := filepath.Join("testdata/scenario05/preferred_applications.json")
	expectedData, err := os.ReadFile(expectedFilePath)
	if err != nil {
		t.Fatalf("error reading '%s': %v", expectedFilePath, err)
	}

	var expected Associations
	err = json.Unmarshal(expectedData, &expected)

	if !cmp.Equal(associations, expected) {
		t.Errorf("Scenario 5 wrong output:\n%s", cmp.Diff(expected, associations))
	}
}
