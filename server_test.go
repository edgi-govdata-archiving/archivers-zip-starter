package main

import (
	"archive/zip"
	"bytes"

	"testing"
)

func TestFetchUrlIfExists(t *testing.T) {
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	defer zw.Close()

	cases := []struct {
		url         string
		shouldFetch bool
		err         error
	}{
		{"ftp://stuff.ftp.dontfetch.com", false, nil},
		{"http://www.apple.com", true, nil},
		{"https://www.apple.com", true, nil},
	}

	for i, c := range cases {
		fetched, got := FetchUrlIfExists("test", map[string]interface{}{"url": c.url}, zw)

		if got != c.err {
			t.Errorf("case %d error mismatch. expected: '%s', got: '%s'", i, c.err, got)
		}

		if fetched != c.shouldFetch {
			t.Errorf("case %d error fetched mismatch. expected: %t, got: %t", i, c.shouldFetch, fetched)
		}

	}
}
