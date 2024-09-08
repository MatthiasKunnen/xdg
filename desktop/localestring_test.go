package desktop

import "testing"

func sliceToMap[T comparable](src []T) map[T]T {
	result := make(map[T]T)

	for _, t := range src {
		result[t] = t
	}

	return result
}

func TestLocaleString_ToLocale_Default(t *testing.T) {
	expected := "Default"
	lstring := LocaleString{
		Default:   expected,
		Localized: sliceToMap([]string{"fr"}),
	}

	result := lstring.ToLocale("nl")
	if result != expected {
		t.Fatalf("Expected: %s, got: %s", expected, result)
	}
}

func TestLocaleString_ToLocale_langCountryModifier(t *testing.T) {
	expected := "nl_BE@custom"
	lstring := LocaleString{
		Default:   "Default",
		Localized: sliceToMap([]string{"fr", "nl", "nl@custom", "nl_BE", expected}),
	}

	result := lstring.ToLocale("nl_BE@custom")
	if result != expected {
		t.Fatalf("Expected: %s, got: %s", expected, result)
	}
}
func TestLocaleString_ToLocale_langCountryEncModifier(t *testing.T) {
	expected := "nl_BE@custom"
	lstring := LocaleString{
		Default:   "Default",
		Localized: sliceToMap([]string{"fr", "nl", "nl@custom", "nl_BE", expected}),
	}

	result := lstring.ToLocale("nl_BE.UTF-8@custom")
	if result != expected {
		t.Fatalf("Expected: %s, got: %s", expected, result)
	}
}

func TestLocaleString_ToLocale_langCountry(t *testing.T) {
	expected := "nl_BE"
	lstring := LocaleString{
		Default:   "Default",
		Localized: sliceToMap([]string{"fr", "nl", "nl@custom", expected, "nl_BE@custom"}),
	}

	result := lstring.ToLocale("nl_BE.UTF-8")
	if result != expected {
		t.Fatalf("Expected: %s, got: %s", expected, result)
	}
}

func TestLocaleString_ToLocale_langCountryEnc(t *testing.T) {
	expected := "nl_BE"
	lstring := LocaleString{
		Default:   "Default",
		Localized: sliceToMap([]string{"fr", "nl", "nl@custom", expected, "nl_BE@custom"}),
	}

	result := lstring.ToLocale("nl_BE")
	if result != expected {
		t.Fatalf("Expected: %s, got: %s", expected, result)
	}
}

func TestLocaleString_ToLocale_langModifierEnc(t *testing.T) {
	expected := "nl@custom"
	lstring := LocaleString{
		Default:   "Default",
		Localized: sliceToMap([]string{"fr", "nl", expected, "nl_BE", "nl_BE@custom"}),
	}

	result := lstring.ToLocale("nl.UTF-8@custom")
	if result != expected {
		t.Fatalf("Expected: %s, got: %s", expected, result)
	}
}

func TestLocaleString_ToLocale_langModifier(t *testing.T) {
	expected := "nl@custom"
	lstring := LocaleString{
		Default:   "Default",
		Localized: sliceToMap([]string{"fr", "nl", expected, "nl_BE", "nl_BE@custom"}),
	}

	result := lstring.ToLocale("nl@custom")
	if result != expected {
		t.Fatalf("Expected: %s, got: %s", expected, result)
	}
}

func TestLocaleString_ToLocaleSpecExample(t *testing.T) {
	expected := "sr_YU"
	lstring := LocaleString{
		Default:   "Default",
		Localized: sliceToMap([]string{expected, "sr@Latn", "sr"}),
	}

	result := lstring.ToLocale("sr_YU@Latn")
	if result != expected {
		t.Fatalf("Expected: %s, got: %s", expected, result)
	}
}
