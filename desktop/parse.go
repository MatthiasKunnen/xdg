package desktop

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

const desktopActionPrefix = "Desktop Action "
const requiredGroupHeader = "[Desktop Entry]"
const requiredGroupName = "Desktop Entry"

const (
	StartupNotifyUnset = iota
	StartupNotifyTrue
	StartupNotifyFalse
)

const (
	TypeApplication = "Application"
	TypeLink        = "Link"
	TypeDirectory   = "Directory"
)

const (
	parseStateLookingForDEGroup = iota
	parseStateLookingForGroupsOrKeys
)

var ErrEscapeIncomplete = errors.New("unexpected end of string, escape sequence not completed")
var ErrActionHasNoGroup = errors.New("action has no matching Desktop Action Group")

func Parse(reader io.Reader) (*Entry, error) {
	var entry Entry
	sc := bufio.NewScanner(reader)

	seenKeys := make(map[string]bool)
	seenGroups := make(map[string]bool)
	actions := make(map[string]bool)
	var currentAction *Action

	parseState := parseStateLookingForDEGroup
	var groupName string

	lineNumber := -1
	for sc.Scan() {
		lineNumber++
		line := strings.TrimRight(sc.Text(), " \t")
		switch {
		case len(line) == 0:
			continue
		case strings.HasPrefix(line, "#"):
			continue
		}

		if parseState == parseStateLookingForDEGroup {
			if line != requiredGroupHeader {
				return &entry, fmt.Errorf(
					"parse failure at line %d, expected %s, found %s",
					lineNumber,
					requiredGroupHeader,
					line,
				)
			} else {
				parseState = parseStateLookingForGroupsOrKeys
				seenGroups[requiredGroupName] = true
				continue
			}
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if currentAction != nil && currentAction.Name.Default != "" {
				entry.Actions = append(entry.Actions, *currentAction)
			}
			currentAction = nil

			groupName = line[1 : len(line)-1]
			if seenGroups[groupName] {
				return &entry, fmt.Errorf(
					"parse failure at line %d, duplicate group %s",
					lineNumber,
					groupName,
				)
			}
			seenGroups[groupName] = true
			clear(seenKeys)

			if strings.HasPrefix(groupName, desktopActionPrefix) {
				actionName := groupName[len(desktopActionPrefix):]

				// Action groups that are not in the Actions key are ignored
				if _, exists := actions[actionName]; exists {
					actions[actionName] = true
					currentAction = &Action{}
				}
			}

			if entry.OtherGroups == nil {
				entry.OtherGroups = make(map[string]map[string]string)
			}

			entry.OtherGroups[groupName] = make(map[string]string)
			continue
		}

		keyValSplit := strings.SplitN(line, "=", 2)
		if len(keyValSplit) < 2 {
			return &entry, fmt.Errorf("parse failure on line %d, tried to read key-value"+
				" line but no value could be determined. Line: %s", lineNumber, line)
		}

		key := keyValSplit[0]
		value := keyValSplit[1]

		if !isValidKey(key) {
			return &entry, fmt.Errorf(
				"parse failure at line %d, invalid key: %s",
				lineNumber,
				key,
			)
		}

		if !isValidValue(value) {
			return &entry, fmt.Errorf(
				"parse failure at line %d, invalid value: %s",
				lineNumber,
				value,
			)
		}

		if seenKeys[key] {
			return &entry, fmt.Errorf(
				"parse failure at line %d, duplicate key %s",
				lineNumber,
				key,
			)
		}
		seenKeys[key] = true

		switch {
		case groupName == "":
			switch key {
			case "Actions":
				list, err := parseList(value)
				if err != nil {
					return &entry, fmt.Errorf(
						"parse failure on line %d, error parsing Actions \"%s\": %w",
						lineNumber,
						value,
						err,
					)
				}

				for _, actionName := range list {
					actions[actionName] = false
				}
			default:
				err := applyMainKeyValue(&entry, key, value)
				if err != nil {
					return &entry, fmt.Errorf(
						"parse failure on line %d, error key='%s', value='%s': %w",
						lineNumber,
						key,
						value,
						err,
					)
				}
			}
		case currentAction != nil:
			keyName, locale, err := parseKey(key)
			if err != nil {
				return &entry, err
			}
			switch keyName {
			case "Name":
				err := assignLocaleString(&currentAction.Name, locale, value)
				if err != nil {
					return &entry, fmt.Errorf(
						"parse failure on line %d, error parsing action.Name %s: %w",
						lineNumber,
						value,
						err,
					)
				}
			case "Icon":
				err := assignIconString(&currentAction.Icon, locale, value)
				if err != nil {
					return &entry, fmt.Errorf(
						"parse failure on line %d, error parsing action.Name %s: %w",
						lineNumber,
						value,
						err,
					)
				}
			case "Exec":
				execValue, err := NewExec(value)
				if err != nil {
					return &entry, fmt.Errorf(
						"parse failure on line %d, error parsing action.Exec %s: %w",
						lineNumber,
						value,
						err,
					)
				}
				currentAction.Exec = execValue
			default:
			}
		default:
			entry.OtherGroups[groupName][key] = value
		}
	}

	if err := sc.Err(); err != nil {
		return &entry, fmt.Errorf("failed reading line on line %d: %w", lineNumber, err)
	}

	for actionName, hasGroup := range actions {
		if hasGroup {
			continue
		}

		return &entry, fmt.Errorf(
			"parse failure, %w: \"%s\"",
			ErrActionHasNoGroup,
			actionName,
		)
	}

	if currentAction != nil && currentAction.Name.Default != "" {
		entry.Actions = append(entry.Actions, *currentAction)
	}

	if entry.Name.Default == "" {
		return &entry, fmt.Errorf("failed to parse: Name field is required")
	}

	if entry.Type == "" {
		return &entry, fmt.Errorf("failed to parse: Type field is required")
	}

	if entry.Type == "Link" && !seenKeys["URL"] {
		return &entry, fmt.Errorf("failed to parse: URL field is required for type Link")
	}

	return &entry, nil
}

func ParseFile(path string) (*Entry, error) {
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return nil, fmt.Errorf("ParseFile, failed to open file %s: %w", path, err)
	}

	return Parse(file)
}

func isValidKey(key string) bool {
	if len(key) == 0 {
		return false
	}

	if strings.HasSuffix(key, "[]") {
		return false
	}

	if !isAsciiNoControl(key) {
		return false
	}

	return true
}

// parseKey parses a key and separates the key and locale.
func parseKey(key string) (string, string, error) {
	var locale string
	var startOfLocale = len(key)
	if strings.HasSuffix(key, "]") {
		startOfLocale = strings.Index(key, "[")
		if startOfLocale == -1 {
			return "", "", fmt.Errorf("key does not have matching opening bracket: %s", key)
		}
		locale = key[startOfLocale+1 : len(key)-1]
	}

	return key[:startOfLocale], locale, nil
}

func isAsciiNoControl(value string) bool {
	for _, r := range value {
		if r > unicode.MaxASCII || unicode.IsControl(r) {
			return false
		}
	}

	return true
}

func isValidValue(value string) bool {
	if len(value) == 0 {
		return false
	}

	return utf8.ValidString(value)
}

func applyMainKeyValue(entry *Entry, key string, value string) error {
	key, locale, err := parseKey(key)
	if err != nil {
		return err
	}

	switch key {
	case "Type":
		s, err := parseString(value)
		if err != nil {
			return err
		}
		entry.Type = s
	case "Version":
		s, err := parseString(value)
		if err != nil {
			return err
		}
		entry.Version = s
	case "Name":
		err := assignLocaleString(&entry.Name, locale, value)
		if err != nil {
			return err
		}
	case "GenericName":
		err := assignLocaleString(&entry.GenericName, locale, value)
		if err != nil {
			return err
		}
	case "NoDisplay":
		boolean, err := parseBoolean(value)
		if err != nil {
			return err
		}
		entry.NoDisplay = boolean
	case "Comment":
		err := assignLocaleString(&entry.Comment, locale, value)
		if err != nil {
			return err
		}
	case "Icon":
		err := assignIconString(&entry.Icon, locale, value)
		if err != nil {
			return err
		}
	case "Hidden":
		boolean, err := parseBoolean(value)
		if err != nil {
			return err
		}
		entry.Hidden = boolean
	case "OnlyShowIn":
		list, err := parseList(value)
		if err != nil {
			return err
		}
		entry.OnlyShowIn = list
	case "NotShowIn":
		list, err := parseList(value)
		if err != nil {
			return err
		}
		entry.NotShowIn = list
	case "DBusActivatable":
		boolean, err := parseBoolean(value)
		if err != nil {
			return err
		}
		entry.DBusActivatable = boolean
	case "TryExec":
		s, err := parseString(value)
		if err != nil {
			return err
		}
		entry.TryExec = s
	case "Exec":
		execVal, err := NewExec(value)
		if err != nil {
			return err
		}
		entry.Exec = execVal
	case "Path":
		s, err := parseString(value)
		if err != nil {
			return err
		}
		entry.Path = s
	case "Terminal":
		boolean, err := parseBoolean(value)
		if err != nil {
			return err
		}
		entry.Terminal = boolean
	case "Actions":
		return errors.New("applyMainKeyValue: Cannot handle key Actions")
	case "MimeType":
		list, err := parseList(value)
		if err != nil {
			return err
		}
		entry.MimeType = list
	case "Categories":
		list, err := parseList(value)
		if err != nil {
			return err
		}
		entry.Categories = list
	case "Implements":
		list, err := parseList(value)
		if err != nil {
			return err
		}
		entry.Implements = list
	case "Keywords":
		err := assignLocaleStrings(&entry.Keywords, locale, value)
		if err != nil {
			return err
		}
	case "StartupNotify":
		hasStartupNotifySupport, err := parseBoolean(value)
		switch {
		case err != nil:
			return err
		case hasStartupNotifySupport:
			entry.StartupNotify = StartupNotifyTrue
		default:
			entry.StartupNotify = StartupNotifyFalse
		}
	case "StartupWMClass":
		s, err := parseString(value)
		if err != nil {
			return err
		}
		entry.StartupWMClass = s
	case "URL":
		s, err := parseString(value)
		if err != nil {
			return err
		}
		entry.URL = s
	case "PrefersNonDefaultGPU":
		boolean, err := parseBoolean(value)
		if err != nil {
			return err
		}
		entry.PrefersNonDefaultGPU = boolean
	case "SingleMainWindow":
		boolean, err := parseBoolean(value)
		if err != nil {
			return err
		}
		entry.SingleMainWindow = boolean
	default:
		if entry.OtherKeys == nil {
			entry.OtherKeys = make(map[string]string)
		}

		entry.OtherKeys[key] = value
	}

	return nil
}

func parseBoolean(value string) (bool, error) {
	switch value {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("parseBoolean, invalid boolean value: %s", value)
	}
}

func parseString(value string) (string, error) {
	if !isAsciiNoControl(value) {
		return "", fmt.Errorf("parseString, value of type string must be ASCII. Got: %s", value)
	}

	unescaped, err := unescapeString(value)
	if err != nil {
		return "", fmt.Errorf("parseString, unescape error for %s: %w", value, err)
	}
	value = unescaped

	return value, nil
}

func assignLocaleString(localeString *LocaleString, locale string, value string) error {
	unescaped, err := unescapeString(value)
	if err != nil {
		return err
	}
	value = unescaped

	if locale == "" {
		localeString.Default = value
	} else {
		if localeString.Localized == nil {
			localeString.Localized = make(map[string]string)
		}

		localeString.Localized[locale] = value
	}

	return nil
}

func assignLocaleStrings(localeStrings *LocaleStrings, locale string, value string) error {
	list, err := splitEscapedString(value)
	if err != nil {
		return err
	}

	if locale == "" {
		localeStrings.Default = list
	} else {
		if localeStrings.Localized == nil {
			localeStrings.Localized = make(map[string][]string)
		}

		localeStrings.Localized[locale] = list
	}

	return nil
}

func assignIconString(iconString *IconString, locale, value string) error {
	unescaped, err := unescapeString(value)
	if err != nil {
		return err
	}
	value = unescaped

	if locale == "" {
		iconString.Default = value
	} else {
		if iconString.Localized == nil {
			iconString.Localized = make(map[string]string)
		}
		iconString.Localized[locale] = value
	}

	return nil
}

func parseList(value string) ([]string, error) {
	if !isAsciiNoControl(value) {
		return nil, fmt.Errorf("parseList, value of type string must be ASCII. Got: %s", value)
	}

	return splitEscapedString(value)
}

// unescapeString converts escaped characters such as \n to actual newlines as defined in
// https://specifications.freedesktop.org/desktop-entry-spec/1.5/value-types.html.
func unescapeString(s string) (string, error) {
	var builder strings.Builder
	builder.Grow(len(s))

	i := 0
	for i < len(s) {
		cur := s[i]
		if cur == '\\' {
			if i+1 >= len(s) {
				return "", ErrEscapeIncomplete
			}

			switch s[i+1] {
			case 's':
				builder.WriteByte(' ')
			case 'n':
				builder.WriteByte('\n')
			case 't':
				builder.WriteByte('\t')
			case 'r':
				builder.WriteByte('\r')
			case '\\':
				builder.WriteByte('\\')
			default:
				builder.WriteByte(cur)
				i++
				continue
			}
			i += 2
			continue
		}

		builder.WriteByte(cur)
		i++
	}

	return builder.String(), nil
}

// splitEscapedString splits the input string by semicolons that are not escaped.
func splitEscapedString(s string) ([]string, error) {
	var result []string
	var current strings.Builder
	escaped := false

	for _, char := range s {
		if escaped {
			switch char {
			case ';':
				current.WriteRune(char)
				escaped = false
			default:
				current.WriteRune('\\')
				current.WriteRune(char)
				escaped = false
			}
		} else if char == '\\' {
			escaped = true
		} else if char == ';' {
			// If char is ';' and not escaped, add the current segment to the result slice
			result = append(result, current.String())
			current.Reset()
		} else {
			current.WriteRune(char)
		}
	}

	if escaped {
		return nil, ErrEscapeIncomplete
	}

	if segment := current.String(); segment != "" {
		result = append(result, segment)
	}

	for i := range result {
		unescaped, err := unescapeString(result[i])
		if err != nil {
			return nil, err
		}
		result[i] = unescaped
	}

	return result, nil
}
