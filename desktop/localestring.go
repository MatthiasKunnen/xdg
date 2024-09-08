package desktop

import (
	"fmt"
	"regexp"
)

type LocaleString struct {
	Default   string
	Localized map[string]string
}

var localeStringRegex = regexp.MustCompile(
	"([a-z]{2,})(?:_([A-Z]{2}))?(?:\\.[a-zA-Z0-9-]+)?(?:@(.+))?$",
)

// ToLocale returns the value of the string according to the requested locale as specified in
// [Localized values for keys].
// Locale has the following format: lang_COUNTRY.ENCODING@MODIFIER where _COUNTRY, .ENCODING, and
// @MODIFIER may be omitted.
//
// [Localized values for keys]: https://specifications.freedesktop.org/desktop-entry-spec/1.5/localized-keys.html
func (s *LocaleString) ToLocale(locale string) string {
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
		if s.Localized[matchedKey] != "" {
			return s.Localized[matchedKey]
		}
	}

	return s.Default
}
