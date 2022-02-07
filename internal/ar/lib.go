package ar

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/awrenn/apt-s3/internal/debug"
)

// The .ar standard has never been specified...
// So as far as I can tell, the only way to extract .deb files is to exec the OS's ar binary.
// It isn't present in any libc I could find.
func ListFiles(ctx context.Context, path string) (paths []string, err error) {
	err = checkFile(path)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "ar", "t", path)
	cmd.Stderr = debug.LoadLogWriter(ctx)

	out, err := cmd.Output()
	if err != nil {
		return nil, attemptToConvertError(err)
	}
	lines := strings.Split(string(out), "\n")
	return lines, nil
}

// Other packages should not be responsible for converting cmd error to a regular error for printing to the user.
// So attempt to convert to sentinal value here.
func attemptToConvertError(err error) error {
	return err
}

func GetFileReader(ctx context.Context, path, file string) (r io.ReadCloser, err error) {
	err = checkFile(path)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "ar", "p", path, file)

	cmd.Stderr = debug.LoadLogWriter(ctx)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, attemptToConvertError(err)
	}
	return stdout, nil
}

func checkFile(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return ErrIsDirectory
	}
	return nil
}
