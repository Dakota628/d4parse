package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/karrick/godirwalk"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	workers    = 256
	maxRetries = 100
)

var r2DeployConfigs = []R2DeployConfig{
	{
		AccountId:  "4bc1086f05e53d1b4006b5edfb5b7732",
		BucketName: aws.String("docs-diablo-parse"),
		LocalPath:  "docs",
	},
	{
		AccountId:  "4bc1086f05e53d1b4006b5edfb5b7732",
		BucketName: aws.String("map-diablo-farm"),
		LocalPath:  filepath.Join("map", "dist"),
	},
}

type R2DeployConfig struct {
	AccountId  string
	BucketName *string
	LocalPath  string
}

func getAwsConfig(dc R2DeployConfig, accessKeyId string, secretAccessKey string) aws.Config {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", dc.AccountId),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, ""),
		),
		config.WithRetryMode(aws.RetryModeAdaptive),
		config.WithRetryMaxAttempts(maxRetries),
	)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

func getFilesToUpload(dc R2DeployConfig) chan string {
	c := make(chan string, workers*10)

	go func() {
		err := godirwalk.Walk(
			dc.LocalPath,
			&godirwalk.Options{
				Callback: func(path string, de *godirwalk.Dirent) error {
					if !de.IsRegular() {
						return nil
					}

					c <- path
					return nil
				},
				Unsorted: true,
			},
		)

		if err != nil {
			// NOTE: godirwalk is currently broken on windows and will EOF after the first dir
			log.Fatalf("Failed to walk directory tree: %s", err)
		}
		close(c)
	}()

	return c
}

func getObjectKey(dc R2DeployConfig, path string) string {
	name, err := filepath.Rel(dc.LocalPath, path)
	if err != nil {
		log.Fatalf("Failed to get object name for %q: %s", path, err)
	}
	parts := strings.Split(name, string(filepath.Separator))
	return strings.Join(parts, "/")
}

func main() {
	accessKeyId, ok := os.LookupEnv("AWS_ACCESS_KEY_ID")
	if !ok {
		log.Fatal("Missing AWS_ACCESS_KEY_ID")
	}

	secretAccessKey, ok := os.LookupEnv("AWS_SECRET_ACCESS_KEY")
	if !ok {
		log.Fatal("Missing AWS_SECRET_ACCESS_KEY")
	}

	if err := mime.AddExtensionType(".bin", "application/octet-stream"); err != nil {
		log.Fatalf("Failed to add .bin mime extension: %s", err)
	}
	if err := mime.AddExtensionType(".mpk", "application/msgpack"); err != nil {
		log.Fatalf("Failed to add .mpk mime extension: %s", err)
	}
	if err := mime.AddExtensionType(".binpb", "application/protobuf"); err != nil {
		log.Fatalf("Failed to add .binpb mime extension: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, dc := range r2DeployConfigs {
		cfg := getAwsConfig(dc, accessKeyId, secretAccessKey)
		files := getFilesToUpload(dc)

		var count atomic.Uint64
		wg := &sync.WaitGroup{}
		for i := uint(0); i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				client := s3.NewFromConfig(cfg)

				for {
					path, ok := <-files
					if !ok {
						return
					}

					objectKey := getObjectKey(dc, path)
					contentType := mime.TypeByExtension(filepath.Ext(path))

					file, err := os.Open(path)
					if err != nil {
						log.Fatalf("Failed to open file %q: %s", path, err)
					}

					fileStat, err := file.Stat()
					if err != nil {
						log.Fatalf("Failed to stat file %q: %s", path, err)
					}

					params := &s3.PutObjectInput{
						Bucket:        dc.BucketName,
						Key:           aws.String(objectKey),
						Body:          file,
						ContentLength: fileStat.Size(),
						ContentType:   aws.String(contentType),
					}

					for _, err = client.PutObject(ctx, params); err != nil; {
						log.Printf("Retrying PutObject %q with new client due to error: %s", objectKey, err)
						client = s3.NewFromConfig(cfg)
						time.Sleep(1 * time.Second)
					}

					newCount := count.Add(1)
					log.Printf("Uploaded %s->%s (#%d)", path, objectKey, newCount)

					// TODO: track uploaded successfully files to allow continuing; don't re-upload same file
				}
			}()
		}

		wg.Wait()
	}
}
