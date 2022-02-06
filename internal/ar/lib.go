package ar

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"strings"
)

// The .ar standard has never been specified...
// So as far as I can tell, the only way to extract .deb files is to exec the OS's ar binary.
// It isn't present in any libc I could find.
func ListFiles(ctx context.Context, path string) (paths []string, err error) {
	cmd := exec.CommandContext(ctx, "ar", "t", path)
	out, err := cmd.CombinedOutput()
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

func GetFileReader(ctx context.Context, path, file string) (r io.Reader, err error) {
	cmd := exec.CommandContext(ctx, "ar", "p", path)

	// We might need this if we want to display a better error to the user.
	// TODO acwrenn
	cmd.Stderr = io.Discard

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Run()
	if err != nil {
		return nil, attemptToConvertError(err)
	}
	return bufio.NewReader(stdout), nil
}
