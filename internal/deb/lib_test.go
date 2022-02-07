package deb

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/awrenn/apt-s3/internal/debug"
)

func TestListFiles(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dummy, err := debug.FindDummy(wd)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	deb, err := ExtractDeb(ctx, dummy)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Files: %+v", deb.Files)
	required := map[string]bool{
		"debian-binary":  false,
		"control.tar.xz": false,
		"data.tar.xz":    false,
	}
	for _, f := range deb.Files {
		required[f] = true
	}
	for r, f := range required {
		if !f {
			t.Errorf("Missed a required debian file: %s", r)
		}
	}

	control, err := deb.GetFileReader(ctx, "control.tar.xz")
	if err != nil {
		t.Fatal(err)
	}
	defer control.Close()

	raw, err := io.ReadAll(control)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Raw control tar: %s", raw)
}

func TestGeneratePackageFile(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dummy, err := debug.FindDummy(wd)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	deb, err := ExtractDeb(ctx, dummy)
	if err != nil {
		t.Fatal(err)
	}

	packageFile := deb.GeneragePackageFile(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Package file:\n%+v", string(packageFile))
}

func TestAddPackageToRelease(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dummy, err := debug.CreateDummy(wd, "v"+debug.RandomString(16))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	deb, err := ExtractDeb(ctx, dummy)
	if err != nil {
		t.Fatal(err)
	}

	packageFile := deb.GeneragePackageFile(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Package file:\n%+v", string(packageFile))
}
