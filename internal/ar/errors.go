package ar

import (
	"errors"
)

var (
	ErrIsDirectory = errors.New("Requested deb file a directory.")
)
