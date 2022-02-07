package deb

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ulikunitz/xz"

	"github.com/awrenn/apt-s3/internal/ar"
)

type DebPackage struct {
	RawPath     string
	RawFileName string
	Files       []string

	Control *ControlManifest
}

func ExtractDeb(ctx context.Context, path string) (pkg DebPackage, err error) {
	files, err := ar.ListFiles(ctx, path)
	if err != nil {
		return pkg, err
	}
	_, f := filepath.Split(path)
	pkg = DebPackage{
		RawPath:     path,
		RawFileName: f,
		Files:       files,
	}
	err = pkg.installControl(ctx)
	if err != nil {
		return pkg, err
	}
	return pkg, nil
}

// What path prefix should we use?
// Normally apt package directory are made into a prefix tree.
// so dummy.deb would live in /d/du/dummy.deb in the directory structure.
// So this function should return /d/du for steps == 2
//
func (d DebPackage) PathPrefix(steps int) string {
	path := ""
	elem := ""
	for i := 0; i < steps; i++ {
		elem += string(d.RawFileName[i])
		path = fmt.Sprintf("%s/%s", path, elem)
	}
	return path
}

func (d DebPackage) GetRawDebReader() (io.ReadCloser, error) {
	f, err := os.OpenFile(d.RawPath, 0400, fs.FileMode(os.O_RDONLY))
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (d DebPackage) GetFileReader(ctx context.Context, file string) (io.ReadCloser, error) {
	r, err := ar.GetFileReader(ctx, d.RawPath, file)
	if err != nil {
		return nil, err
	}
	if strings.Contains(file, ".xz") {
		xr, err := xz.NewReader(r)
		if err != nil {
			return nil, err
		}
		r = xzReader{xr, r}
	}
	return r, nil
}

type xzReader struct {
	xr *xz.Reader
	c  io.Closer
}

func (r xzReader) Read(b []byte) (int, error) {
	return r.xr.Read(b)
}

func (r xzReader) Close() error {
	return r.c.Close()
}

func (d DebPackage) Distribution() string {
	return "stable"
}

func (d DebPackage) Size() (string, error) {
	r, err := d.GetRawDebReader()
	if err != nil {
		return "", err
	}
	defer r.Close()
	n, err := io.Copy(io.Discard, r)
	return strconv.Itoa(int(n)), err
}

func (d *DebPackage) installControl(ctx context.Context) error {
	control, err := d.GetFileReader(ctx, "control.tar.xz")
	if err != nil {
		return err
	}
	t := tar.NewReader(control)
	for {
		section, err := t.Next()
		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("Early EOF - did not find control file: %w", err)
			}
			return err
		}
		if filepath.Clean(section.Name) == "control" {
			break
		}
	}
	rawControl, err := io.ReadAll(t)
	if err != nil {
		return err
	}
	final, err := parseControl(rawControl)
	if err != nil {
		return err
	}

	final.Size, err = d.Size()
	if err != nil {
		return err
	}
	// We are going to over-ride the sha fields.
	final.SHA256, err = d.SHA256()
	if err != nil {
		return err
	}
	final.SHA1, err = d.SHA1()
	if err != nil {
		return err
	}
	final.MD5sum, err = d.MD5()
	if err != nil {
		return err
	}
	d.Control = final
	return nil
}

func (d DebPackage) Arch() string {
	return d.Control.Architecture
}

func (d DebPackage) Repo() string {
	return "main"
}
