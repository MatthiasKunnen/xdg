package desktop

import (
	"fmt"
	"regexp"
)

type localized[T any] struct {
	Default   T
	Localized map[string]T
}

type LocaleString = localized[string]
type LocaleStrings = localized[[]string]

var localeStringRegex = regexp.MustCompile(
	"([a-z]{2,})(?:_([A-Z]{2}))?(?:\\.[a-zA-Z0-9-]+)?(?:@(.+))?$",
)

// ToLocale returns the value of the string according to the requested locale as specified in
// [Localized values for keys].
// Locale has the following format: lang_COUNTRY.ENCODING@MODIFIER where _COUNTRY, .ENCODING, and
// @MODIFIER may be omitted.
//
// [Localized values for keys]: https://specifications.freedesktop.org/desktop-entry-spec/1.5/localized-keys.html
func (s *localized[T]) ToLocale(locale string) T {
	matches := localeStringRegex.FindStringSubmatch(locale)

	if matches == nil {
		return s.Default
	}

	lang := matches[1]
	country := matches[2]
	modifier := matches[3]

	checks := make([]string, 4)

	if country != "" && modifier != "" {
		checks = append(checks, fmt.Sprintf("%s_%s@%s", lang, country, modifier))
	}

	if country != "" {
		checks = append(checks, fmt.Sprintf("%s_%s", lang, country))
	}

	if modifier != "" {
		checks = append(checks, fmt.Sprintf("%s@%s", lang, modifier))
	}

	checks = append(checks, lang)

	for _, matchedKey := range checks {
		maybe := s.Localized[matchedKey]
		switch v := any(maybe).(type) {
		case string:
			if v != "" {
				return maybe
			}
		case []string:
			if v != nil && len(v) > 0 {
				return maybe
			}
		default:
			panic("unsupported type")
		}
	}

	return s.Default
}
