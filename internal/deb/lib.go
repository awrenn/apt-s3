package deb

import "context"

type DebPackage struct {
}

func ExtractDeb(ctx context.Context, path string) (DebPackage, error) {

	return DebPackage{}, nil
}
