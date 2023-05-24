package type1_test

import (
	"io/ioutil"
	"testing"

	"gfx/font/type1"
)

func TestParse(t *testing.T) {
	data, err := ioutil.ReadFile("/Users/bryan/Go/src/go.matteson.dev/fitz/deps/mupdf/resources/fonts/urw/input/NimbusMonoPS-Bold.t1")
	if err != nil {
		t.Fatal(err)
	}
	if err := type1.Parse(data); err != nil {
		t.Fatal(err)
	}
}
