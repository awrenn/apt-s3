package s3

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/awrenn/apt-s3/internal/deb"
	"github.com/awrenn/apt-s3/internal/debug"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

var S3TestConfig S3Config

func testMain(m *testing.M) int {
	flag.Parse()
	if testing.Short() {
		return 0
	}

	port := uint16(10000 + rand.Intn(20000))
	mc, err := debug.NewMinio("dummy-apt", "apt-s3_minio_tester", port)
	if err != nil {
		log.Println(err)
		return 1
	}
	defer mc.Close()
	S3TestConfig = S3Config{
		Endpoint:         fmt.Sprintf("http://localhost:%d", port),
		DisableSSL:       true,
		S3ForcePathStyle: true,

		Bucket: "dummy-apt",
		Region: "apt-s3",

		// Tests that test GPG signing should clone this struct and set these themselves.
		Sign:     false,
		GPGKeyID: "",

		UseStaticCreds: true,
		AccessKey:      mc.AccessKey,
		SecretKey:      mc.SecretKey,
	}
	return m.Run()
}

func TestS3Upload(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Skipf("Need WD to run this test...")
	}
	dummy, err := debug.FindDummy(wd)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cls := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cls()
	deb, err := deb.ExtractDeb(ctx, dummy)
	if err != nil {
		t.Fatal(err)
	}
	err = PublishDebFileWithConfig(ctx, deb, S3TestConfig)
	if err != nil {
		t.Fatal(err)
	}
	c := S3TestConfig.GenerateS3Config()
	sess := session.Must(session.NewSession(c))
	downloader := s3manager.NewDownloader(sess)

	fp := fmt.Sprintf("pool/%s%s/%s", deb.Distribution(), deb.PathPrefix(2), deb.RawFileName)
	b := aws.NewWriteAtBuffer(make([]byte, 0))
	_, err = downloader.DownloadWithContext(ctx, b, &s3.GetObjectInput{
		Bucket: &S3TestConfig.Bucket,
		Key:    aws.String(fp),
	})
	if err != nil {
		t.Fatal(err)
	}
	packages, err := fetchOldPackageFile(ctx, downloader, deb, S3TestConfig.Bucket)
	if err != nil {
		t.Fatal(err)
	}
	ourIndex := -1
	t.Logf("Packages found: %+v", packages)
	for i, p := range packages {
		if p.Version == deb.Control.Version && p.Package == deb.Control.Package {
			ourIndex = i
			break
		}
	}
	if ourIndex == -1 {
		t.Fatal("Could not find new package manifest")
	}
	if deb.Control.MD5sum != packages[ourIndex].MD5sum {
		t.Error("Incorrect MD5Sum")
	}
	if deb.Control.SHA1 != packages[ourIndex].SHA1 {
		t.Error("Incorrect SHA1")
	}
	if deb.Control.SHA256 != packages[ourIndex].SHA256 {
		t.Error("Incorrect SHA256")
	}

}
