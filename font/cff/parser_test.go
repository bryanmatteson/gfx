package cff_test

import (
	"io/ioutil"
	"testing"

	"github.com/bryanmatteson/gfx/font/cff"
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
	glyph, err := coll.FirstFont.GenerateGlyph("dollar")
	if err != nil {
		t.Fatal(err)
	}

	_ = glyph
}
