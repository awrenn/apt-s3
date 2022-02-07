package deb

import (
	"os"
	"testing"
)

func TestParseRelease(t *testing.T) {
	raw, err := os.ReadFile("./Release")
	if err != nil {
		t.Fatal(err)
	}
	man, err := ParseRelease(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v\n", man.Serialize())
}
