package desktop

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// ExecValue is two-dimensional representation of the [Exec key].
// This is required to meet the specification's criteria that field codes must not
// be used in quotes while they can be used outside of them.
//
// This means that in `Exec=test %f "hello"%cthere "%i"`, %f and %c should be expanded but %i
// shouldn't.
// This is represented using:
//
//	[
//		[{arg: "test", isFieldCode: false}],
//		[
//			{arg: "hello", isFieldCode: false},
//			{arg: "c", isFieldCode: true},
//			{arg: "there", isFieldCode: false},
//		],
//		[{arg: "%i", isFieldCode: false}],
//	]
//
// %F, %U, and %i will always be in separate arguments as the spec dictates.
//
// [Exec key]: https://specifications.freedesktop.org/desktop-entry-spec/latest/exec-variables.html
type ExecValue [][]execArgPart

// CanOpenFiles returns true if opening local file(s) is explicitly supported.
func (e ExecValue) CanOpenFiles() bool {
	return e.HasAnyFieldCode([]byte{'f', 'F', 'u', 'U'})
}

// CanOpenUrls returns true if opening URL(s) is explicitly supported.
func (e ExecValue) CanOpenUrls() bool {
	return e.HasAnyFieldCode([]byte{'u', 'U'})
}

// HasAnyFieldCode returns true if any of the given field codes are present.
// If no field codes are given, false is returned.
func (e ExecValue) HasAnyFieldCode(fieldCodes []byte) bool {
	if len(fieldCodes) == 0 {
		return false
	}

	for _, parts := range e {
		for _, part := range parts {
			if !part.isFieldCode {
				continue
			}

			// Field codes cannot be more than a single letter according to the spec.
			// This is enforced elsewhere, these are just sanity checks.
			if len(part.arg) == 0 {
				panic("ExecValue with fieldCode: true, empty arg")
			}

			if len(part.arg) > 1 {
				panic(fmt.Sprintf(
					"ExecValue with fieldCode: true, arg longer than 1 character: %s",
					part.arg,
				))
			}

			if slices.Contains(fieldCodes, part.arg[0]) {
				return true
			}
		}
	}

	return false
}

// execArgPart represents the parts of the Exec key.
// If isFieldCode is true, arg represents a field code without the "%".
// Otherwise, arg is a plain argument that will not be expanded.
type execArgPart struct {
	// if isFieldCode is true, a single letter field code.
	// if isFieldCode is false, a part of a program argument or the complete argument.
	arg         string
	isFieldCode bool
}

// FieldCodeProvider provides the functions that allow expansion of the Exec field codes.
type FieldCodeProvider struct {
	// GetDesktopFileLocation related to the %k field code.
	// If an empty string is returned, the field code is not expanded.
	GetDesktopFileLocation func() string

	// GetFile relates to the %f field code.
	// If an empty string is returned, the field code is not expanded.
	GetFile func() string

	// GetFiles relates to the %F field code.
	// If the slice is empty, the field code is not expanded.
	GetFiles func() []string

	// GetIcon relates to the %i field code.
	// If an empty string is returned, the field code is not expanded.
	GetIcon func() string

	// GetName relates to the %c field code.
	// If an empty string is returned, the field code is not expanded.
	GetName func() string

	// GetUrl relates to the %u field code.
	// If an empty string is returned, the field code is not expanded.
	GetUrl func() string

	// GetUrls relates to the %U field code.
	// If the slice is empty, the field code is not expanded.
	GetUrls func() []string
}

var (
	ErrAsciiRequired           = errors.New("must be ascii")
	ErrCharacterMustBeQuoted   = errors.New("character must be quoted")
	ErrEscapeOutsideQuotes     = errors.New("invalid character escaped")
	ErrFieldCodeIncomplete     = errors.New("unexpected end of string, field code not completed")
	ErrFieldCodeMustBeOwnArg   = errors.New("%F and %U must be separate arguments")
	ErrQuoteNotCompleted       = errors.New("double quote does not have matching closing quote")
	ErrTooManyFileFieldCodes   = errors.New("more than one file field code (fuFU)")
	ErrUnknownEscapedCharacter = errors.New("character must not be escaped")
	ErrUnknownFieldCode        = errors.New("unknown field code")
)

// NewExec parses the given strings as an Exec key from the Desktop Entry specification.
// See https://specifications.freedesktop.org/desktop-entry-spec/1.5/exec-variables.html.
func NewExec(value string) (ExecValue, error) {
	if value == "" {
		return nil, fmt.Errorf("error: Exec value is empty")
	}

	if !isAsciiNoControl(value) {
		return nil, fmt.Errorf("exec value %w. Got: \"%s\"", ErrAsciiRequired, value)
	}

	value, err := unescapeString(value)
	if err != nil {
		return nil, err
	}

	result := make(ExecValue, 0)
	quoted := false
	var nextArg strings.Builder

	argParts := make([]execArgPart, 0)
	containsFileFieldCode := false
	escaped := false

	appendArgPart := func() {
		part := nextArg.String()
		if part == "" {
			return
		}

		argParts = append(argParts, execArgPart{
			arg:         part,
			isFieldCode: false,
		})
		nextArg.Reset()
	}

	for i := 0; i < len(value); i++ {
		char := value[i]

		if escaped {
			switch char {
			case '"', '`', '$', '\\':
				nextArg.WriteByte(char)
				escaped = false
				continue
			default:
				return nil, fmt.Errorf("parseExec: %w: %c", ErrUnknownEscapedCharacter, char)
			}
		}

		switch char {
		case '\\':
			if !quoted {
				return nil, fmt.Errorf("parseExec: %w", ErrEscapeOutsideQuotes)
			}
			escaped = true
			continue
		case '"':
			appendArgPart()
			quoted = !quoted
		case ' ':
			switch {
			case quoted:
				nextArg.WriteByte(' ')
			case nextArg.Len() == 0 && len(argParts) == 0:
				continue
			default:
				appendArgPart()
				result = append(result, argParts)
				argParts = nil
			}
		case '%':
			switch {
			case quoted:
				nextArg.WriteByte(char)
				continue
			case i+1 > len(value):
				return nil, fmt.Errorf("parseExec: %w", ErrFieldCodeIncomplete)
			default:
				fieldCode := value[i+1]
				addFieldCode := false

				switch fieldCode {
				case '%':
					nextArg.WriteByte('%')
				case 'd', 'D', 'n', 'N', 'v', 'm':
					// Deprecated
				case 'F', 'U':
					if containsFileFieldCode {
						return nil, fmt.Errorf("parseExec: %w", ErrTooManyFileFieldCodes)
					}

					if i+2 < len(value) && value[i+2] != ' ' {
						return nil, fmt.Errorf("parseExec: %w", ErrFieldCodeMustBeOwnArg)
					}

					containsFileFieldCode = true
					addFieldCode = true
				case 'f', 'u':
					if containsFileFieldCode {
						return nil, fmt.Errorf("parseExec: %w", ErrTooManyFileFieldCodes)
					}
					containsFileFieldCode = true
					addFieldCode = true
				case 'i', 'c', 'k':
					addFieldCode = true
				default:
					return nil, fmt.Errorf("%w: %c", ErrUnknownFieldCode, fieldCode)
				}
				i++

				if addFieldCode {
					appendArgPart()
					argParts = append(argParts, execArgPart{
						arg:         string(fieldCode),
						isFieldCode: true,
					})
				}
			}
		case '\t', '\n', '\'', '>', '<', '~', '|', '&', ';', '$', '*', '?', '#',
			'(', ')', '`':
			if !quoted {
				return nil, fmt.Errorf("parseExec: %w: %c", ErrCharacterMustBeQuoted, char)
			}
			nextArg.WriteByte(char)
		default:
			nextArg.WriteByte(char)
		}
	}

	if escaped {
		return nil, ErrEscapeIncomplete
	}

	if quoted {
		return nil, fmt.Errorf("parseExec: %w", ErrQuoteNotCompleted)
	}

	appendArgPart()
	if len(argParts) > 0 {
		result = append(result, argParts)
	}

	return result, nil
}

// ToArguments converts the Exec value to a list of arguments ready to be passed for execution.
func (e ExecValue) ToArguments(handler FieldCodeProvider) []string {
	result := make([]string, 0, len(e))
	var argument strings.Builder

	addArguments := func(arg ...string) {
		if argument.Len() > 0 {
			result = append(result, argument.String())
			argument.Reset()
		}
		result = append(result, arg...)
	}

	for _, parts := range e {
		for _, part := range parts {
			if part.isFieldCode {
				switch part.arg {
				case "f":
					if handler.GetFile == nil {
						continue
					}
					file := handler.GetFile()
					if file != "" {
						argument.WriteString(file)
					}
				case "F":
					if handler.GetFiles == nil {
						continue
					}
					files := handler.GetFiles()
					if len(files) > 0 {
						addArguments(files...)
					}
				case "u":
					if handler.GetUrl == nil {
						continue
					}
					url := handler.GetUrl()
					if url != "" {
						argument.WriteString(url)
					}
				case "U":
					if handler.GetUrls == nil {
						continue
					}
					urls := handler.GetUrls()
					if len(urls) > 0 {
						addArguments(urls...)
					}
				case "i":
					if handler.GetIcon == nil {
						continue
					}
					icon := handler.GetIcon()
					if icon != "" {
						addArguments("--icon", icon)
					}
				case "c":
					if handler.GetName == nil {
						continue
					}
					translatedName := handler.GetName()
					if translatedName != "" {
						argument.WriteString(translatedName)
					}
				case "k":
					if handler.GetDesktopFileLocation == nil {
						continue
					}
					location := handler.GetDesktopFileLocation()
					if location != "" {
						argument.WriteString(location)
					}
				}
			} else {
				argument.WriteString(part.arg)
			}
		}

		if argument.Len() > 0 {
			result = append(result, argument.String())
			argument.Reset()
		}
	}

	return result
}
