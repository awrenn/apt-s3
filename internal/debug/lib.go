package debug

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	DebugFileKey = "debug-file"
)

func LoadLogWriter(ctx context.Context) io.Writer {
	f, ok := ctx.Value(DebugFileKey).(io.Writer)
	if !ok {
		return nil
	}
	return f
}

func RegisterLogFile(ctx context.Context, path string) (context.Context, error) {
	f, err := os.OpenFile(path, 0200, fs.FileMode(os.O_WRONLY))
	if err != nil {
		return ctx, err
	}
	return context.WithValue(ctx, DebugFileKey, f), nil
}

func RegisterLogWriter(ctx context.Context, w io.Writer) context.Context {
	return context.WithValue(ctx, DebugFileKey, w)
}

func FindDummy(curDir string) (fin string, err error) {
	return findFile(curDir, "dummy.deb")
}

func FindRelease(curDir string) (fin string, err error) {
	return findFile(curDir, "Release")
}

func findFile(curDir, file string) (fin string, err error) {
	if curDir == "/" {
		return "", errors.New("Dummy file not found")
	}
	files, err := os.ReadDir(curDir)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		if f.Name() == file && !f.IsDir() {
			return filepath.Join(curDir, file), nil
		}
	}
	return findFile(filepath.Clean(filepath.Join(curDir, "..")), file)
}
