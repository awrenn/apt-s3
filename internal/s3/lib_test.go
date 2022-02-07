package s3

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/awrenn/apt-s3/internal/deb"
	"github.com/awrenn/apt-s3/internal/debug"
)

func TestS3Upload(t *testing.T) {
	bucket := "dummy-apt"
	region := "us-west-2"
	wd, err := os.Getwd()
	if err != nil {
		t.Skipf("Need WD to run this test...")
	}
	dummy, err := debug.FindDummy(wd)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cls := context.WithTimeout(context.Background(), time.Minute)
	defer cls()
	deb, err := deb.ExtractDeb(ctx, dummy)
	if err != nil {
		t.Fatal(err)
	}
	err = PublishDebFile(ctx, deb, region, bucket)
	if err != nil {
		t.Fatal(err)
	}
}
