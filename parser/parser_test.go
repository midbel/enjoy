package parser

import (
	"os"
	"testing"
)

func TestParser(t *testing.T) {
	files := []string{
		"testdata/variables.js",
		"testdata/control.js",
		"testdata/func.js",
	}
	for _, f := range files {
		parseFile(t, f)
	}
}

func parseFile(t *testing.T, file string) {
	t.Helper()
	r, err := os.Open(file)
	if err != nil {
		t.Errorf("fail to open file: %s", file)
		return
	}
	defer r.Close()

	_, err = Parse(r)
	if err != nil {
		t.Errorf("fail to parse file %s: %s", file, err)
	}
}
