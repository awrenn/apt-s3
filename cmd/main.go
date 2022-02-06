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
		printHelp()
		os.Exit(0)
	}
	switch args[1] {
	case "pub", "publish", "p":
		os.Exit(publishCommand())
	default:
		fmt.Printf("Command \"%s\" not understood: \n", args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`This is a tool for syncing local deb files with s3 in order to create easy, hosted apt repos.
		Valid commands are:
		- pub (aliases publish, p)
			Publish a local deb file to a given s3 bucket.
			Required flags:
			-b (bucket), -r (region), -f (deb file)
	`)
}

func publishCommand() int {
	flagSet := flag.NewFlagSet("publish", flag.ContinueOnError)
	regionFlag := flagSet.String("region", "", "Which region is the target s3 bucket in?")
	bucketFlag := flagSet.String("bucket", "", "Which bucket should we extract our deb package into?")
	debFlag := flagSet.String("deb", "", "What is the path to the desired deb file?")
	flag.Parse()
	ctx, cls := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cls()

	debFile, err := deb.ExtractDeb(ctx, *debFlag)
	if err != nil {
		return handleError(err)
	}
	return handleError(s3.PublishDebFile(ctx, debFile, *regionFlag, *bucketFlag))
}

func handleError(err error) int {
	if err == nil {
		return 0
	}
	fmt.Println(err)
	return 1
}
