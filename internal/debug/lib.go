package debug

import (
	"context"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	DebugFileKey = "debug-file"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

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
	return CreateDummy(curDir, RandomString(6))
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

func CreateDummy(curDir, version string) (fin string, err error) {
	makefile, err := findFile(curDir, "Makefile")
	if err != nil {
		return "", err
	}
	makedir := filepath.Dir(makefile)
	cmd := exec.Command("make", "dummy")
	cmd.Env = append(cmd.Env, fmt.Sprintf("Version=%s", version))
	cmd.Dir = makedir
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return filepath.Join(makedir, "dummy", fmt.Sprintf("dummy_%s.deb", version)), nil
}

func RandomString(l int) string {
	b := make([]byte, l)
	rand.Read(b)
	b64 := base32.StdEncoding.EncodeToString(b)
	return b64[:l]
}
