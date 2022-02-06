package s3

import (
	"context"
	"errors"

	"github.com/awrenn/apt-s3/internal/deb"
)

func PublishDebFile(ctx context.Context, debFile deb.DebPackage, region, bucket string) error {
	return errors.New("Not yet impl")
}
