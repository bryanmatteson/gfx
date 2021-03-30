package afm_test

import (
	"os"
	"testing"

	"go.matteson.dev/gfx/internal/font/afm"
)

func TestParser(t *testing.T) {
	file, err := os.Open("/Users/bryan/Go/src/go.matteson.dev/fitz/deps/mupdf/resources/fonts/urw/input/NimbusSans-Regular.afm")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	metrics, err := afm.ParseReader(file)
	if err != nil {
		t.Fatal(err)
	}
	_ = metrics
}
