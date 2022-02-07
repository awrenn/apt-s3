package deb

import (
	"bytes"
	"fmt"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type ReleaseSum struct {
	Hash string
	Size string
	Path string
}

type ReleaseManfifest struct {
	Codename      string
	Date          string
	Architectures string
	Components    string
	Suite         string
	md5Sums       []ReleaseSum
	sha1Sums      []ReleaseSum
	sha256Sums    []ReleaseSum
	extra         map[string]string
}

func ParseRelease(raw string) (*ReleaseManfifest, error) {
	// First, let's grab the main parts.
	// No spaces or : allowed in field names here!
	// Thanks bad specs!
	init := &ReleaseManfifest{
		extra: make(map[string]string),
	}
	fieldRe := regexp.MustCompile("^([^: ]+): ?(.*)$")
	hashRe := regexp.MustCompile("^\\s*(\\S+)\\s+(\\d+)\\s+(.*)$")
	for _, line := range strings.Split(raw, "\n") {
		matches := fieldRe.FindStringSubmatch(line)
		var name string
		var value string
		var f reflect.Value
		if matches == nil || len(matches) < 3 {
			goto TryRE
		}
		name = matches[1]
		value = matches[2]
		f = reflect.ValueOf(init).Elem().FieldByName(name)
		if !f.IsValid() {
			init.extra[name] = value
			continue
		}
		if !f.CanInterface() {
			continue
		}
		f.Set(reflect.ValueOf(value))
		continue
	TryRE:
		matches = hashRe.FindStringSubmatch(line)
		if matches == nil || len(matches) < 4 {
			continue
		}
		init.AddHash(matches[1], matches[2], matches[3])
	}
	init.Date = time.Now().Format(time.RFC1123)
	return init, nil
}

func (deb DebPackage) NewReleaseManifest() *ReleaseManfifest {
	return &ReleaseManfifest{
		Codename:      deb.Distribution(),
		Date:          time.Now().Format(time.RFC1123),
		Architectures: deb.Arch(),
		Components:    deb.Repo(),
		Suite:         "",
		md5Sums:       make([]ReleaseSum, 0),
		sha1Sums:      make([]ReleaseSum, 0),
		sha256Sums:    make([]ReleaseSum, 0),
	}
}

func (r *ReleaseManfifest) Serialize() string {
	out := new(strings.Builder)
	out.WriteString(fmt.Sprintf("Codename: %s\n", r.Codename))
	out.WriteString(fmt.Sprintf("Date: %s\n", r.Date))
	out.WriteString(fmt.Sprintf("Architectures: %s\n", r.Architectures))
	out.WriteString(fmt.Sprintf("Components: %s\n", r.Components))
	out.WriteString(fmt.Sprintf("Suite: %s\n", r.Suite))
	out.WriteString(fmt.Sprintf("MD5Sum:\n"))
	for _, m := range r.md5Sums {
		out.WriteString(fmt.Sprintf("  %s\t%s\t%s\n", m.Hash, m.Size, m.Path))
	}
	out.WriteString(fmt.Sprintf("SHA1:\n"))
	for _, m := range r.sha1Sums {
		out.WriteString(fmt.Sprintf("  %s\t%s\t%s\n", m.Hash, m.Size, m.Path))
	}
	out.WriteString(fmt.Sprintf("SHA256:\n"))
	for _, m := range r.sha256Sums {
		out.WriteString(fmt.Sprintf("  %s\t%s\t%s\n", m.Hash, m.Size, m.Path))
	}
	return out.String()
}

func (r *ReleaseManfifest) SerializeAndSign(key string) (string, error) {
	raw := r.Serialize()
	cmd := exec.Command("gpg", "-a", "-s", "--clearsign", "-b", "-u", key)
	cmd.Stdin = bytes.NewBuffer([]byte(raw))

	eb := bytes.NewBuffer(make([]byte, 0))
	cmd.Stderr = eb

	sig, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error from GPG:\n%s", string(eb.Bytes()))
		return "", err
	}

	return string(sig), nil
}

func (r *ReleaseManfifest) AddArch(arch string) {
	parts := strings.Split(r.Architectures, " ")
	for _, p := range parts {
		if p == arch {
			return
		}
	}
	r.Architectures = r.Architectures + " " + arch
}

func (r *ReleaseManfifest) AddComponent(comp string) {
	parts := strings.Split(r.Components, " ")
	for _, p := range parts {
		if p == comp {
			return
		}
	}
	r.Components = r.Components + " " + comp
}

func (r *ReleaseManfifest) UpdateDate() {
	now := time.Now().UTC()
	r.Date = now.Format(time.RFC1123)
}

func (r *ReleaseManfifest) AddHash(hash, size, path string) {
	s := ReleaseSum{
		Hash: hash,
		Size: size,
		Path: path,
	}
	switch len(hash) {
	case 32:
		for i, s := range r.md5Sums {
			if s.Path == path {
				r.md5Sums[i] = s
				return
			}
		}
		r.md5Sums = append(r.md5Sums, s)
	case 40:
		for i, s := range r.sha1Sums {
			if s.Path == path {
				r.sha1Sums[i] = s
				return
			}
		}
		r.sha1Sums = append(r.sha1Sums, s)
	case 64:
		for i, s := range r.sha256Sums {
			if s.Path == path {
				r.sha256Sums[i] = s
				return
			}
		}
		r.sha256Sums = append(r.sha256Sums, s)
	default:
		fmt.Println("WARNING: Found ill-understood hash.")
	}
}
