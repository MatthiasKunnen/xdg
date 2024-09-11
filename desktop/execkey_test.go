package desktop

import (
	"errors"
	"slices"
	"testing"
)

func TestNewExecKeyBackslashEscape(t *testing.T) {
	exec, err := NewExec(`test "\\\\"`)
	if err != nil {
		t.Fatal(err)
	}

	if len(exec) != 2 {
		t.Errorf("len(exec)=%d; want 2", len(exec))
	}

	if exec[0][0].isFieldCode == true {
		t.Errorf("exec[0].isFieldCode=true; want false")
	}

	if exec[0][0].arg != "test" {
		t.Errorf("exec[0].arg=%q; want \"test\"", exec[0][0].arg)
	}

	if exec[1][0].isFieldCode == true {
		t.Errorf("exec[1].isFieldCode=true; want false")
	}

	if exec[1][0].arg != `\` {
		t.Errorf("exec[1].arg=%q; want \"\\\"", exec[1][0].arg)
	}
}

func TestNewExec_FieldCodes(t *testing.T) {
	result, err := NewExec(`test %f %i %ch "hello"%kthere`)

	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 5 {
		t.Errorf("len(result)=%d; want 5", len(result))
	}

	if result[1][0].isFieldCode == false {
		t.Errorf("result[1][0].isFieldCode=false; want true")
	}

	if result[1][0].arg != "f" {
		t.Errorf("result[1][0].arg=%q; want \"f\"", result[1][0].arg)
	}

	if result[2][0].isFieldCode == false {
		t.Errorf("result[2][0].isFieldCode=false; want true")
	}

	if result[2][0].arg != `i` {
		t.Errorf("result[2][0].arg=%q; want \"i\"", result[2][0].arg)
	}

	if len(result[3]) != 2 {
		t.Errorf("len(result[3])=%d; want 2", len(result[3]))
	}

	expected3 := []execArgPart{
		{arg: "c", isFieldCode: true},
		{arg: "h", isFieldCode: false},
	}
	if !slices.Equal(result[3], expected3) {
		t.Errorf("result[3]=%v; want %v", result[3], expected3)
	}

	expected4 := []execArgPart{
		{arg: "hello", isFieldCode: false},
		{arg: "k", isFieldCode: true},
		{arg: "there", isFieldCode: false},
	}
	if !slices.Equal(result[4], expected4) {
		t.Errorf("result[4]=%v; want %v", result[4], expected3)
	}
}

func TestNewExec_KeyTooManyFieldCodes(t *testing.T) {
	_, err := NewExec(`test %f %F`)
	if !errors.Is(err, ErrTooManyFileFieldCodes) {
		t.Errorf("err = %v; want ErrTooManyFileFieldCodes", err)
	}
}

func TestNewExec_KeyTooManyFieldCodesInQuotes(t *testing.T) {
	_, err := NewExec(`test %f "%F"`)

	switch {
	case errors.Is(err, ErrTooManyFileFieldCodes):
		t.Errorf("got unexpected ErrTooManyFileFieldCodes")
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewExec_RequiredQuotationOfReserved(t *testing.T) {
	tests := []string{
		`\t`, `\n`, `'`, ">", "<", "~", "|", "&", ";", "$", "*", "?", "#",
		"(", ")", "`",
	}

	for _, test := range tests {
		_, err := NewExec(test)

		if !errors.Is(err, ErrCharacterMustBeQuoted) {
			t.Errorf("err = %v; want ErrCharacterMustBeQuoted for %s", err, test)
		}
	}
}

func TestNewExec_DeprecatedRemoved(t *testing.T) {
	result, err := NewExec(`%d %D %n %N %v %m`)

	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 0 {
		t.Errorf("len(result)=%d; want 0. result: %v", len(result), result)
	}
}

func TestNewExec_EscapePercent(t *testing.T) {
	result, err := NewExec(`%%`)

	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 1 {
		t.Errorf("len(result)=%d; want 1", len(result))
	}

	if result[0][0].arg != "%" {
		t.Errorf("result[0].arg=%q; want \"%%\"", result[0][0].arg)
	}
}

func TestNewExec_UnknownFieldCode(t *testing.T) {
	_, err := NewExec(`%X`)

	switch {
	case errors.Is(err, ErrUnknownFieldCode):
	default:
		t.Errorf("err = %v; want ErrUnknownFieldCode", err)
	}
}

func TestExecValue_ToArguments_FCf(t *testing.T) {
	exec, err := NewExec(`test Well%cHello %f "--location="%k`)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"test", "Well_Name_Hello", "/usr/bin/true", "--location=/tmp/d.desktop"}
	actual := exec.ToArguments(FieldCodeProvider{
		GetDesktopFileLocation: func() string {
			return "/tmp/d.desktop"
		},
		GetFile: func() string {
			return "/usr/bin/true"
		},
		GetName: func() string {
			return "_Name_"
		},
	})
	if !slices.Equal(expected, actual) {
		t.Errorf("Expected: %v; actual: %v", expected, actual)
	}
}

func TestExecValue_ToArguments_Icons(t *testing.T) {
	exec, err := NewExec(`test %i`)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"test", "--icon", "banana.jpeg"}
	actual := exec.ToArguments(FieldCodeProvider{
		GetIcon: func() string {
			return "banana.jpeg"
		},
	})
	if !slices.Equal(expected, actual) {
		t.Errorf("Expected: %v; actual: %v", expected, actual)
	}
}

func TestExecValue_ToArguments_IconsTogether(t *testing.T) {
	exec, err := NewExec(`test%i`)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"test", "--icon", "banana.jpeg"}
	actual := exec.ToArguments(FieldCodeProvider{
		GetIcon: func() string {
			return "banana.jpeg"
		},
	})
	if !slices.Equal(expected, actual) {
		t.Errorf("Expected: %v; actual: %v", expected, actual)
	}
}

func TestExecValue_ToArguments_FCF(t *testing.T) {
	exec, err := NewExec(`test%F`)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"test", "/usr/bin/true", "/usr/bin/false"}
	actual := exec.ToArguments(FieldCodeProvider{
		GetFiles: func() []string {
			return []string{"/usr/bin/true", "/usr/bin/false"}
		},
	})
	if !slices.Equal(expected, actual) {
		t.Errorf("Expected: %v; actual: %v", expected, actual)
	}
}

func TestExecValue_CanOpenFiles(t *testing.T) {
	test := func(value string, expected bool) {
		exec, err := NewExec(value)
		if err != nil {
			t.Fatalf("Unexpected error creating exec value: %v", err)
		}

		if exec.CanOpenFiles() != expected {
			t.Errorf("CanOpenFiles \"%s\" = %v; want %v", value, !expected, expected)
		}
	}

	test(`test %f`, true)
	test(`test %k %u`, true)
	test(`test "%f"`, false)
	test(`test %k`, false)
}
