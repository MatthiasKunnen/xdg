package mimeapps

import "testing"

func TestUtilIsSubPathAbs(t *testing.T) {
	test := func(parent string, sub string, expected bool) {
		if isSubPathAbs(sub, parent) != expected {
			t.Errorf("isSubPathAbs(%s, %s) is not %t", sub, parent, expected)
		}
	}

	test("/tmp/dir", "/tmp/dir/test", true)
	test("/tmp/dir1", "/tmp/dir2", false)
	test("/tmp/dir", "/tmp/dirwithlongername/test", false)
}
