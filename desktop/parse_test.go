package desktop

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	result, err := Parse(strings.NewReader(`
[Desktop Entry]
Type=Application
Name=Firefox
Name[nl]=Vuurvos
Name[nl_BE]=Vúúrvos
`))

	if err != nil {
		t.Fatal(err)
	}

	if result.Type != "Application" {
		t.Errorf("result.Type = %v, want %v", result.Type, "Application")
	}

	nlBeName := result.Name.ToLocale("nl_BE")
	if nlBeName != "Vúúrvos" {
		t.Errorf("Name with locale nl_BE is %s, expected Vúúrvos", nlBeName)
	}
}
func TestParseWithComment(t *testing.T) {
	result, err := Parse(strings.NewReader(`
# Test
[Desktop Entry]
# This

# Thing
Type=Application
Name=Firefox
`))

	if err != nil {
		t.Fatal(err)
	}

	if result.Type != "Application" {
		t.Errorf("result.Type = %v, want %v", result.Type, "Application")
	}
}

func TestParseExtraGroup(t *testing.T) {
	result, err := Parse(strings.NewReader(`
[Desktop Entry]
Type=Application
Name=Firefox
Name[nl]=Vuurvos
Name[nl_BE]=Vúúrvos

[Extra]
X-Crazy=Hello
`))

	if err != nil {
		t.Fatal(err)
	}

	if result.Type != "Application" {
		t.Errorf("result.Type = %v, want %v", result.Type, "Application")
	}

	if _, exists := result.OtherGroups["Extra"]; !exists {
		t.Errorf("OtherGroup['Extra'] is not set")
	}

	keyVal := result.OtherGroups["Extra"]["X-Crazy"]
	if keyVal != "Hello" {
		t.Errorf("result.OtherGroups[\"Extra\"][\"X-Crazy\"] is %s, expected: Hello", keyVal)
	}
}

func TestParseKeywords(t *testing.T) {
	result, err := Parse(strings.NewReader(`
[Desktop Entry]
Type=Application
Name=Firefox
Keywords=browser;Internet;WWW;
`))

	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"browser", "Internet", "WWW"}
	if !slices.Equal(result.Keywords.Default, expected) {
		t.Errorf("Keywords is %v, expected: %v", result.Keywords.Default, expected)
	}
}

func TestParseKeywordsNoEolSemicolon(t *testing.T) {
	result, err := Parse(strings.NewReader(`
[Desktop Entry]
Type=Application
Name=Firefox
Keywords=browser;Internet;WWW
`))

	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"browser", "Internet", "WWW"}
	if !slices.Equal(result.Keywords.Default, expected) {
		t.Errorf("Keywords is %v, expected: %v", result.Keywords.Default, expected)
	}
}

func TestParseUnescape(t *testing.T) {
	result, err := Parse(strings.NewReader(`
[Desktop Entry]
Type=Application
Name=Firefox\nAnd\ssons\\
Keywords=The\nKeyword\;\sfactory;Hey
Keywords[nl]=\s\n;\;\r\t\\a
Keywords[cr]=Two\\\\;items
`))

	if err != nil {
		t.Fatal(err)
	}

	expected := `Firefox
And sons\`
	if result.Name.Default != expected {
		t.Errorf("Name is %v, expected: %v", result.Name.Default, expected)
	}

	expectedKeywordsDefault := []string{`The
Keyword; factory`, "Hey"}
	if !slices.Equal(result.Keywords.Default, expectedKeywordsDefault) {
		t.Errorf("Keywords is %v, expected: %v", result.Keywords.Default, expected)
	}

	expectedKeywordsNl := []string{" \n", ";\r\t\\a"}
	if !slices.Equal(result.Keywords.ToLocale("nl"), expectedKeywordsNl) {
		t.Errorf(
			"Keywords is %v, expected: %v",
			result.Keywords.ToLocale("nl"),
			expectedKeywordsNl,
		)
	}

	expectedKeywordsCr := []string{"Two\\\\", "items"}
	if !slices.Equal(result.Keywords.ToLocale("cr"), expectedKeywordsCr) {
		t.Errorf(
			"Keywords is %v, expected: %v",
			result.Keywords.ToLocale("cr"),
			expectedKeywordsCr,
		)
	}
}

func TestParseErrorOnUnterminatedEscape(t *testing.T) {
	_, err := Parse(strings.NewReader(`
[Desktop Entry]
Type=Application
Name=Firefox\
`))

	if !errors.Is(err, ErrEscapeIncomplete) {
		t.Errorf("Expected error, got none")
	}
}

func TestParseErrorOnUnterminatedEscape2(t *testing.T) {
	_, err := Parse(strings.NewReader(`
[Desktop Entry]
Type=Application
Name=Firefox
Keywords=Test\;h\
`))

	if !errors.Is(err, ErrEscapeIncomplete) {
		t.Errorf("Expected error, got none")
	}
}

func TestParseUnescapeStringValue(t *testing.T) {
	result, err := Parse(strings.NewReader(`
[Desktop Entry]
Type=Application
Name=Firefox\tTabbed
`))

	if err != nil {
		t.Fatal(err)
	}

	expected := "Firefox\tTabbed"
	if result.Name.Default != expected {
		t.Errorf("Name is %s, expected: %s", result.Name.Default, expected)
	}
}

func TestParse_ActionsWithoutGroup(t *testing.T) {
	_, err := Parse(strings.NewReader(`
[Desktop Entry]
Type=Application
Name=Firefox\tTabbed
Actions=Gallery
`))

	if !errors.Is(err, ErrActionHasNoGroup) {
		t.Errorf("Expected ErrActionHasNoGroup, got %v", err)
	}
}

func TestParse_Actions(t *testing.T) {
	result, err := Parse(strings.NewReader(`
[Desktop Entry]
Type=Application
Name=Firefox\tTabbed
Actions=Gallery;

[Desktop Action Gallery]
Name=Browse gallery
Name[nl]=Bekijk gallerij
Exec=firefox --gallery
`))
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Actions) != 1 {
		t.Errorf("There are %d actions, expected: %d", len(result.Actions), 1)
	}

	expectedDefault := "Browse gallery"
	if result.Actions[0].Name.Default != expectedDefault {
		t.Errorf(
			"Action name is %s, expected: %s",
			result.Actions[0].Name.Default,
			expectedDefault,
		)
	}

	actualNameNl := result.Actions[0].Name.ToLocale("nl")
	expectedNameNl := "Bekijk gallerij"
	if actualNameNl != expectedNameNl {
		t.Errorf("Action name is %s, expected: %s", actualNameNl, expectedNameNl)
	}

	actualExec1 := result.Actions[0].Exec[1]
	expectedExec1 := []execArgPart{
		{arg: "--gallery", isFieldCode: false},
	}
	if !slices.Equal(actualExec1, expectedExec1) {
		t.Errorf("Action Exec arg 1 is: %v, expected: %v", actualExec1, expectedExec1)
	}
}

func TestParse_MultipleActions(t *testing.T) {
	result, err := Parse(strings.NewReader(`
[Desktop Entry]
Type=Application
Name=Firefox\tTabbed
Actions=Gallery;Number2

[Desktop Action Gallery]
Name=Browse gallery
Name[nl]=Bekijk gallerij
Exec=firefox --gallery

[Desktop Action Number2]
Name=Number2

[Desktop Action Not defined]
Name=Browse gallery
`))
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Actions) != 2 {
		t.Errorf("There are %d actions, expected: 2", len(result.Actions))
	}

	expectedDefault1 := "Browse gallery"
	actualDefault1 := result.Actions[0].Name.Default
	if actualDefault1 != expectedDefault1 {
		t.Errorf("Action name is %s, expected: %s", actualDefault1, expectedDefault1)
	}

	expectedDefault2 := "Number2"
	actualDefault2 := result.Actions[1].Name.Default
	if actualDefault2 != expectedDefault2 {
		t.Errorf("Action name is %s, expected: %s", actualDefault2, expectedDefault2)
	}
}
