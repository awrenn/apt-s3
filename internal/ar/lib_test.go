package ar

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/awrenn/apt-s3/internal/debug"
)

func TestFullFlowWithDummy(t *testing.T) {
	// First, we need to find dummy.deb.
	wd, err := os.Getwd()
	if err != nil {
		t.Skipf("We need to an FS for this test: %s", err)
	}
	path, err := debug.FindDummy(wd)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBuffer(make([]byte, 0))
	ctx := debug.RegisterLogWriter(context.Background(), b)
	files, err := ListFiles(ctx, path)
	if err != nil {
		t.Error("Stderr: ", b.String())
		t.Fatal(err)
	}
	t.Logf("All files: %+v", files)

	var sample string
	for _, f := range files {
		if strings.Contains(f, "control") {
			sample = f
			break
		}
	}
	r, err := GetFileReader(ctx, path, sample)
	if err != nil {
		t.Fatal(err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(out))
}
