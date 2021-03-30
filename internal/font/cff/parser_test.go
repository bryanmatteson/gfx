package cff_test

import (
	"io/ioutil"
	"testing"

	"go.matteson.dev/gfx/internal/font/cff"
)

func TestParser(t *testing.T) {
	data, err := ioutil.ReadFile("/Users/bryan/Go/src/go.matteson.dev/fitz/deps/mupdf/resources/fonts/urw/NimbusSans-Regular.cff")
	if err != nil {
		t.Fatal(err)
	}

	coll, err := cff.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	_ = coll
}
