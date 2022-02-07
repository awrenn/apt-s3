package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/awrenn/apt-s3/internal/deb"
	"github.com/awrenn/apt-s3/internal/s3"
)

func main() {
	args := os.Args
	if len(args) <= 1 {
		os.Exit(0)
	}

	regionFlag := flag.String("region", "", "Which region is the target s3 bucket in?")
	bucketFlag := flag.String("bucket", "", "Which bucket should we extract our deb package into?")
	debFlag := flag.String("deb", "", "What is the path to the desired deb file?")
	keyFlag := flag.String("key", "", "What GPG key should be used to sign this Release? Leave blank for no signature.")

	flag.Parse()

	if *debFlag == "" {
		fmt.Println(".deb file path not provided")
		flag.PrintDefaults()
		os.Exit(1)
	}
	ctx, cls := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cls()

	debFile, err := deb.ExtractDeb(ctx, *debFlag)
	if err != nil {
		os.Exit(handleError(err))
	}
	os.Exit(handleError(s3.PublishDebFile(ctx, debFile, *regionFlag, *bucketFlag, *keyFlag)))
}

func handleError(err error) int {
	if err == nil {
		return 0
	}
	fmt.Println(err)
	return 1
}
