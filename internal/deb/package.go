package deb

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type ControlManifest struct {
	Package       string
	Version       string
	License       string
	Vendor        string
	Architecture  string
	Maintainer    string
	InstalledSize string `flat:"Installed-Size"`
	Section       string
	Priority      string
	Filename      string
	Size          string
	SHA1          string
	SHA256        string
	MD5sum        string
	Description   string
	Essential     string
	extra         map[string]string
}

func (d DebPackage) GeneragePackageFile(ctx context.Context) (packageManifest string) {
	return d.Control.Serialize()
}

func (c *ControlManifest) Serialize() string {
	out := ""
	v := reflect.ValueOf(c).Elem()
	fields := v.Type().NumField()
	for i := 0; i < fields; i++ {
		fieldName := v.Type().Field(i)
		field := v.Field(i)
		if !field.CanInterface() {
			// This is our extra field
			continue
		}
		next := fmt.Sprintf("%s: %s\n", fieldName.Name, field)
		out += next
	}
	for k, v := range c.extra {
		next := fmt.Sprintf("%s: %s\n", k, v)
		out += next
	}
	return out
}

func (c *ControlManifest) SetFilename(f string) {
	c.Filename = f
}

func ParsePackageList(raw []byte) ([]*ControlManifest, error) {
	parts := bytes.Split(raw, []byte("\n\n"))
	out := make([]*ControlManifest, 0, len(parts))
	for _, p := range parts {
		o, err := parseControl(p)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, nil
}

func parseControl(rawControl []byte) (*ControlManifest, error) {
	final := &ControlManifest{
		extra: make(map[string]string),
	}
	lineRe := regexp.MustCompile("^([^:]+):( ?.*)$")
	for _, line := range strings.Split(string(rawControl), "\n") {
		match := lineRe.FindStringSubmatch(line)
		if len(match) != 3 {
			continue
		}
		name := match[1]
		field := reflect.ValueOf(final).Elem().FieldByName(name)
		value := strings.Trim(match[2], " \n\t")
		if !field.IsValid() {
			final.extra[name] = value
			continue
		}
		field.Set(reflect.ValueOf(value))
	}
	return final, nil
}
