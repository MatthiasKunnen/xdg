package desktop

import (
	"strings"
	"testing"
)

func TestMagicIsDesktopFileEmpty(t *testing.T) {
	isDesktopFile, err := MagicIsDesktopFile(strings.NewReader(""))

	if err != nil {
		t.Fatalf("Empty file should not result in an error: %v", err)
	}

	if isDesktopFile {
		t.Fatalf("Empty file should not be recognized as a desktop file")
	}
}

func TestMagicIsDesktopFileSuccessNoBom(t *testing.T) {
	isDesktopFile, err := MagicIsDesktopFile(strings.NewReader(`[Desktop Entry]
Name=Hello
`))

	if err != nil {
		t.Fatalf("Correct file should not result in an error: %v", err)
	}

	if !isDesktopFile {
		t.Fatalf("File should be recognized as a desktop file")
	}
}

func TestMagicIsDesktopFileSuccessWithBom(t *testing.T) {
	isDesktopFile, err := MagicIsDesktopFile(strings.NewReader(
		"\xef\xbb\xbf[Desktop Entry]\nName=Hello",
	))

	if err != nil {
		t.Fatalf("Correct file with UTF-8 BOM should not result in an error: %v", err)
	}

	if !isDesktopFile {
		t.Fatalf("Correct file with UTF-8 BOM should be recognized as a desktop file")
	}
}

func TestMagicIsDesktopFileIncorrectBom(t *testing.T) {
	isDesktopFile, err := MagicIsDesktopFile(strings.NewReader(
		"\xef\xbb\xbe[Desktop Entry]\nName=Hello",
	))

	if err != nil {
		t.Fatalf("Correct file with incorrect BOM should not result in an error: %v", err)
	}

	if isDesktopFile {
		t.Fatalf("Correct file with incorrect BOM should not be recognized as a desktop file")
	}
}

func TestMagicIsDesktopFileSuccessWithComments(t *testing.T) {
	isDesktopFile, err := MagicIsDesktopFile(strings.NewReader(`# Hello there # Maybe
[Desktop Entry]
Name=Hello
`))
	if err != nil {
		t.Fatalf("Correct file with comments should not result in an error: %v", err)
	}

	if !isDesktopFile {
		t.Fatalf("Correct file with comments should be recognized as a desktop file")
	}
}

func TestMagicIsDesktopFileSuccessWithNewlines(t *testing.T) {
	isDesktopFile, err := MagicIsDesktopFile(strings.NewReader(`

[Desktop Entry]
Name=Hello
`))
	if err != nil {
		t.Fatalf("Correct file with newlines not result in an error: %v", err)
	}

	if !isDesktopFile {
		t.Fatalf("Correct file with newlines should be recognized as a desktop file")
	}
}

func TestMagicIsDesktopFileSuccessWithNonUtf8InComment(t *testing.T) {
	isDesktopFile, err := MagicIsDesktopFile(strings.NewReader(
		"# Invalid UTF8 \xD8\x00\n[Desktop Entry]\nName=Hello\n",
	))
	if err != nil {
		t.Fatalf("Invalid UTF-8 in comments should not result in an error: %v", err)
	}

	if !isDesktopFile {
		t.Fatalf("Invalid UTF-8 in comments should not disqualify desktop file")
	}
}
