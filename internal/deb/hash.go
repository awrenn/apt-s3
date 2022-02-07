package deb

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"io"
)

func (d DebPackage) SHA1() (string, error) {
	h := sha1.New()
	f, err := d.GetRawDebReader()
	if err != nil {
		return "", err
	}
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}
	b := make([]byte, 0)
	b = h.Sum(b)
	return hex.EncodeToString(b), nil
}

func (d DebPackage) SHA256() (string, error) {
	h := sha256.New()
	f, err := d.GetRawDebReader()
	if err != nil {
		return "", err
	}
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}
	b := make([]byte, 0)
	b = h.Sum(b)
	return hex.EncodeToString(b), nil
}

func (d DebPackage) MD5() (string, error) {
	h := md5.New()
	f, err := d.GetRawDebReader()
	if err != nil {
		return "", err
	}
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}
	b := make([]byte, 0)
	b = h.Sum(b)
	return hex.EncodeToString(b), nil
}
