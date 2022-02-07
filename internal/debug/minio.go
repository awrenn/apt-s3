package debug

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"
)

const (
	MinioAccessKey = "minioadmin"
	MinioSecretKey = "minioadmin"
)

type MinioContainer struct {
	Name      string
	Port      uint16
	Bucket    string
	AccessKey string
	SecretKey string
}

func NewMinio(bucket, name string, port uint16) (mc MinioContainer, err error) {
	mc = MinioContainer{
		Name:      name,
		Port:      port,
		Bucket:    bucket,
		AccessKey: MinioAccessKey,
		SecretKey: MinioSecretKey,
	}
	// Attempt to destroy an existing container if one exists, but don't get hung up if it fails.
	mc.Close()

	cmd := exec.Command("docker", "run",
		"-d",
		"--name", name,
		"-e", fmt.Sprintf("MINIO_ROOT_USER=%s", MinioAccessKey),
		"-e", fmt.Sprintf("MINIO_ROOT_PASSWORD=%s", MinioSecretKey),
		"-e", fmt.Sprintf("MINIO_ACCESS_KEY=%s", MinioAccessKey),
		"-e", fmt.Sprintf("MINIO_SECRET_KEY=%s", MinioSecretKey),
		"-p", fmt.Sprintf("127.0.0.1:%d:9000", port),
		"minio/minio:latest",
		"server", "/data")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return mc, err
	}

	cmd = exec.Command("docker", "exec", name, "mkdir", "-p", fmt.Sprintf("/data/%s", bucket))
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return mc, err
	}

	ctx, cls := context.WithTimeout(context.Background(), time.Minute)
	defer cls()
	c := &http.Client{}
	for {
		select {
		case <-ctx.Done():
			return mc, errors.New("Minio Container never ready.")
		default:
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://localhost:%d/minio/health/live", port), nil)
		if err != nil {
			log.Fatal(err)
		}
		resp, err := c.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			continue
		}
		break
	}

	return mc, nil

}

func (m MinioContainer) Close() error {
	cmd := exec.Command("docker", "rm", "-f", m.Name)
	return cmd.Run()
}
