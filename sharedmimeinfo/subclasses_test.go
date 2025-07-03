package sharedmimeinfo_test

import (
	"fmt"
	"github.com/MatthiasKunnen/xdg/sharedmimeinfo"
	"github.com/google/go-cmp/cmp"
	"io"
	"log"
	"strings"
	"testing"
)

func ExampleLoadFromOs() {
	s, err := sharedmimeinfo.LoadFromOs()
	if err != nil {
		log.Fatalf("Failed to load subclasses: %v\n", err)
	}
	// Outputs: text/x-python, application/x-executable, text/plain, application/octet-stream
	println(s.BroaderDfs("text/x-python3"))
}

func ExampleLoadFromReaders() {
	s, err := sharedmimeinfo.LoadFromReaders([]io.Reader{
		strings.NewReader(`image/svg+xml application/xml`),
		strings.NewReader("image/svg+xml text/plain"),
	})
	if err != nil {
		log.Fatalf("Failed to load subclasses: %v\n", err)
	}
	fmt.Println(strings.Join(s.BroaderDfs("image/svg+xml"), ", "))
	// Output: application/xml, text/plain, application/octet-stream
}

func TestSubclass_BroaderOnce(t *testing.T) {
	s, err := sharedmimeinfo.LoadFromReaders([]io.Reader{
		strings.NewReader(`image/svg+xml application/xml
application/xml application/xml2
application/xml2 text/xml`),
		strings.NewReader("image/svg+xml application/svg"),
		strings.NewReader("x-scheme-handler/rstp1 x-scheme-handler/rstp2"),
	})
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		input    string
		expected []string
	}

	cases := []testCase{
		{"application/octet-stream", nil},
		{"application/binary", []string{"application/octet-stream"}},
		{"text/plain", []string{"application/octet-stream"}},
		{"inode/fifo", nil},
		{"image/svg+xml", []string{"application/xml", "application/svg"}},
		{"x-scheme-handler/rstp1", []string{"x-scheme-handler/rstp2"}},
		{"x-scheme-handler/http", nil},
	}

	for _, c := range cases {
		got := s.BroaderOnce(c.input)
		if diff := cmp.Diff(c.expected, got); diff != "" {
			t.Errorf("BroaderOnce(%s) mismatch (-want +got):\n%s", c.input, diff)
		}
	}
}

func TestSubclass_BroaderDfs(t *testing.T) {
	s, err := sharedmimeinfo.LoadFromReaders([]io.Reader{
		strings.NewReader(`image/svg+xml application/xml
application/xml application/xml2
application/xml2 text/xml`),
		strings.NewReader("image/svg+xml application/svg"),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"application/xml",
		"application/xml2",
		"text/xml",
		"application/svg",
		"text/plain",
		"application/octet-stream",
	}
	got := s.BroaderDfs("image/svg+xml")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("BroaderDfs() mismatch (-want +got):\n%s", diff)
	}
}

func TestSubclass_BroaderDfs_nested2(t *testing.T) {
	s, err := sharedmimeinfo.LoadFromReaders([]io.Reader{
		strings.NewReader(`image/svg+xml application/xml
application/xml application/xml2
application/xml2 text/plain`),
		strings.NewReader("image/svg+xml application/svg"),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"application/xml",
		"application/xml2",
		"text/plain",
		"application/svg",
		"application/octet-stream",
	}
	got := s.BroaderDfs("image/svg+xml")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("BroaderDfs() mismatch (-want +got):\n%s", diff)
	}
}

func TestSubclass_BroaderDfs_noDuplicates(t *testing.T) {
	s, err := sharedmimeinfo.LoadFromReaders([]io.Reader{
		strings.NewReader(`image/svg+xml application/xml
application/xml application/xml2
application/xml text/plain
application/xml2 text/plain`),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"application/xml",
		"application/xml2",
		"text/plain",
		"application/octet-stream",
	}
	got := s.BroaderDfs("image/svg+xml")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("BroaderDfs() mismatch (-want +got):\n%s", diff)
	}
}

func TestSubclass_BroaderDfs_inode(t *testing.T) {
	s, err := sharedmimeinfo.LoadFromReaders([]io.Reader{
		strings.NewReader(`inode/mount-point inode/directory
application/svg+xml application/xml`),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"inode/directory",
	}
	got := s.BroaderDfs("inode/mount-point")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("BroaderDfs() mismatch (-want +got):\n%s", diff)
	}
}

func TestSubclass_BroaderDfs_textPlainResultsInOctetStream(t *testing.T) {
	s, err := sharedmimeinfo.LoadFromReaders([]io.Reader{
		strings.NewReader(``),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"application/octet-stream",
	}
	got := s.BroaderDfs("text/plain")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("BroaderDfs() mismatch (-want +got):\n%s", diff)
	}
}

func TestSubclass_BroaderDfs_textResultsInTextPlain1(t *testing.T) {
	s, err := sharedmimeinfo.LoadFromReaders([]io.Reader{
		strings.NewReader(`text/foo application/bar`),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"application/bar",
		"text/plain",
		"application/octet-stream",
	}
	got := s.BroaderDfs("text/foo")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("BroaderDfs() mismatch (-want +got):\n%s", diff)
	}
}

func TestSubclass_BroaderDfs_textResultsInTextPlain2(t *testing.T) {
	s, err := sharedmimeinfo.LoadFromReaders([]io.Reader{
		strings.NewReader(``),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"text/plain",
		"application/octet-stream",
	}
	got := s.BroaderDfs("text/foo")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("BroaderDfs() mismatch (-want +got):\n%s", diff)
	}
}

func TestSubclass_BroaderDfs_octetStreamOnly(t *testing.T) {
	s, err := sharedmimeinfo.LoadFromReaders([]io.Reader{
		strings.NewReader(``),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{}
	got := s.BroaderDfs("application/octet-stream")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("BroaderDfs() mismatch (-want +got):\n%s", diff)
	}
}

func TestSubclass_BroaderDfs_schemeHandler(t *testing.T) {
	s, err := sharedmimeinfo.LoadFromReaders([]io.Reader{
		strings.NewReader(`text/foo application/bar`),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{}
	got := s.BroaderDfs("x-scheme-handler/http")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("BroaderDfs() mismatch (-want +got):\n%s", diff)
	}
}
