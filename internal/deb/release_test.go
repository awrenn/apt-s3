package deb

import (
	"os"
	"testing"

	"github.com/awrenn/apt-s3/internal/debug"
)

func TestParseRelease(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	release, err := debug.FindRelease(wd)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(release)
	if err != nil {
		t.Fatal(err)
	}
	man, err := ParseRelease(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v\n", man.Serialize())
}
