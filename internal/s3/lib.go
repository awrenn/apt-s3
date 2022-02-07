package s3

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/awrenn/apt-s3/internal/deb"
)

type PackageHashes struct {
	Regular HashSet
	Zipped  HashSet
}

type HashSet struct {
	MD5    string
	SHA1   string
	SHA256 string
	Size   string
	Path   string
}

func PublishDebFile(ctx context.Context, debFile deb.DebPackage, region, bucket, key string) error {
	sess := session.Must(session.NewSession(
		&aws.Config{
			Region: aws.String(region),
		},
	))
	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)

	// Adding our raw deb to the pool is always fine - someone could download it manually, but apt wouldn't find it.
	path, err := uploadDebFile(ctx, uploader, debFile, bucket)
	if err != nil {
		return err
	}
	debFile.Control.SetFilename(path)

	// These two files need to be completed atomically.
	// If not, we should consider how we rollback.
	// We also need to add locking
	hashes, err := syncPackageFile(ctx, uploader, downloader, debFile, bucket)
	if err != nil {
		return err
	}
	err = syncReleaseFile(ctx, uploader, downloader, debFile, bucket, key, hashes)
	if err != nil {
		return err
	}

	return nil
}

func uploadDebFile(ctx context.Context, uploader *s3manager.Uploader, debFile deb.DebPackage, bucket string) (string, error) {
	r, err := debFile.GetRawDebReader()
	if err != nil {
		return "", err
	}
	defer r.Close()
	fp := fmt.Sprintf("pool/%s%s/%s", debFile.Distribution(), debFile.PathPrefix(2), debFile.RawFileName)
	_, err = uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: &bucket,
		Key:    aws.String(fp),
		Body:   bufio.NewReader(r),
	})
	if err != nil {
		return "", err
	}
	return fp, nil
}

func syncPackageFile(ctx context.Context, uploader *s3manager.Uploader, downloader *s3manager.Downloader, debFile deb.DebPackage, bucket string) (PackageHashes, error) {
	h := PackageHashes{}
	packageKey := fmt.Sprintf("dists/%s/%s/binary-%s/Packages", debFile.Distribution(), debFile.Repo(), debFile.Arch())
	b := aws.NewWriteAtBuffer(make([]byte, 0))
	_, err := downloader.DownloadWithContext(ctx, b, &s3.GetObjectInput{
		Key:    aws.String(packageKey),
		Bucket: aws.String(bucket),
	})
	previous := make([]*deb.ControlManifest, 0)
	if err != nil {
		e, ok := err.(awserr.RequestFailure)
		if ok && e.Code() == s3.ErrCodeNoSuchKey {
			goto Eliminate
		}
		return h, err
	}
	previous, err = deb.ParsePackageList(b.Bytes())
	if err != nil {
		return h, err
	}
Eliminate:
	builder := new(strings.Builder)
	for _, p := range previous {
		if p.Package == "" || p.Version == "" {
			fmt.Println("WARNING - found previously indexed package with no package name or version number. Skipping.")
			continue
		}
		if p.Package == debFile.Control.Package &&
			p.Version == debFile.Control.Version {
			continue
		} else {
			builder.WriteString(p.Serialize())
		}
		builder.WriteString("\n")
	}
	builder.WriteString(debFile.Control.Serialize())
	many := builder.String()
	_, err = uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Key:    aws.String(packageKey),
		Bucket: aws.String(bucket),
		Body:   bytes.NewBuffer([]byte(many)),
	})
	if err != nil {
		return h, err
	}
	buf := make([]byte, 0)

	h.Regular.Size = strconv.Itoa(len(many))
	hasher := sha256.New()
	n, _ := io.Copy(hasher, bytes.NewBuffer([]byte(many)))
	h.Regular.Size = strconv.Itoa(int(n))
	h.Regular.Path = packageKey
	h.Regular.SHA256 = hex.EncodeToString(hasher.Sum(buf))

	hasher = sha1.New()
	io.Copy(hasher, bytes.NewBuffer([]byte(many)))
	h.Regular.SHA1 = hex.EncodeToString(hasher.Sum(buf))

	hasher = md5.New()
	io.Copy(hasher, bytes.NewBuffer([]byte(many)))
	h.Regular.MD5 = hex.EncodeToString(hasher.Sum(buf))

	r := getZipReader([]byte(many))
	_, err = uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Key:    aws.String(packageKey + ".gz"),
		Bucket: aws.String(bucket),
		Body:   bufio.NewReader(r),
	})

	// TODO acwrenn
	// We don't need to re-compress this file over and over and over again
	hasher = sha256.New()
	n, _ = io.Copy(hasher, getZipReader([]byte(many)))
	h.Zipped.Size = strconv.Itoa(int(n))
	h.Zipped.Path = packageKey + ".gz"
	h.Zipped.SHA256 = hex.EncodeToString(hasher.Sum(buf))

	hasher = sha1.New()
	io.Copy(hasher, getZipReader([]byte(many)))
	h.Zipped.SHA1 = hex.EncodeToString(hasher.Sum(buf))

	hasher = md5.New()
	io.Copy(hasher, getZipReader([]byte(many)))
	h.Zipped.MD5 = hex.EncodeToString(hasher.Sum(buf))

	return h, err
}

func syncReleaseFile(ctx context.Context, uploader *s3manager.Uploader, downloader *s3manager.Downloader, debFile deb.DebPackage, bucket, gpgKey string, hashes PackageHashes) error {
	releaseName := fmt.Sprintf("dists/%s/Release", debFile.Distribution())

	b := aws.NewWriteAtBuffer(make([]byte, 0))
	_, err := downloader.DownloadWithContext(ctx, b, &s3.GetObjectInput{
		Key:    aws.String(releaseName),
		Bucket: aws.String(bucket),
	})
	manifest := debFile.NewReleaseManifest()
	if err != nil {
		e, ok := err.(awserr.RequestFailure)
		if ok && e.Code() == s3.ErrCodeNoSuchKey {
			goto Send
		}
		return err
	}
	manifest, err = deb.ParseRelease(string(b.Bytes()))
	if err != nil {
		return err
	}
Send:
	manifest.AddArch(debFile.Arch())
	manifest.AddComponent(debFile.Repo())
	manifest.UpdateDate()
	// OK, we would have just updated our Packages and Packages.gz
	// files.
	manifest.AddHash(hashes.Regular.MD5, hashes.Regular.Size, hashes.Regular.Path)
	manifest.AddHash(hashes.Regular.SHA1, hashes.Regular.Size, hashes.Regular.Path)
	manifest.AddHash(hashes.Regular.SHA256, hashes.Regular.Size, hashes.Regular.Path)
	manifest.AddHash(hashes.Zipped.MD5, hashes.Zipped.Size, hashes.Zipped.Path)
	manifest.AddHash(hashes.Zipped.SHA1, hashes.Zipped.Size, hashes.Zipped.Path)
	manifest.AddHash(hashes.Zipped.SHA256, hashes.Zipped.Size, hashes.Zipped.Path)

	_, err = uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Key:    aws.String(releaseName),
		Bucket: aws.String(bucket),
		Body:   bytes.NewBuffer([]byte(manifest.Serialize())),
	})
	if err != nil {
		return err
	}

	// We don't have a key passed to us, we're done.
	if gpgKey == "" {
		return nil
	}
	// Otherwise, let's upload an InRelease file as well.
	man, err := manifest.SerializeAndSign(gpgKey)
	if err != nil {
		return err
	}
	inReleaseName := fmt.Sprintf("dists/%s/InRelease", debFile.Distribution())
	_, err = uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Key:    aws.String(inReleaseName),
		Bucket: aws.String(bucket),
		Body:   bytes.NewBuffer([]byte(man)),
	})
	return err
}

func getZipReader(b []byte) io.Reader {
	r, w := io.Pipe()
	go func() {
		b := bytes.NewReader(b)
		gw := gzip.NewWriter(w)
		defer w.Close()
		defer gw.Close()

		_, err := io.Copy(gw, b)
		if err != nil {
			fmt.Println(err)
		}
	}()
	return r
}
