package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/karrick/godirwalk"
	"io"
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
	bucketName = "docs-diablo-parse"
	accountId  = "4bc1086f05e53d1b4006b5edfb5b7732"

	localBasePath = "docs"
	workers       = 256
	maxRetries    = 100
)

var (
	awsBucketName = aws.String(bucketName)
)

func getAwsConfig(accessKeyId string, secretAccessKey string) aws.Config {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId),
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

func getFilesToUpload() chan string {
	c := make(chan string, workers*10)

	go func() {
		err := godirwalk.Walk(
			localBasePath,
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

		if err != nil && err != io.EOF {
			log.Fatalf("Failed to walk directory tree: %s", err)
		}
		close(c)
	}()

	return c
}

func getObjectKey(path string) string {
	name, err := filepath.Rel(localBasePath, path)
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := getAwsConfig(accessKeyId, secretAccessKey)
	files := getFilesToUpload()

	if err := mime.AddExtensionType(".bin", "application/octet-stream"); err != nil {
		log.Fatalf("Failed to add .bin mime extension: %s", err)
	}

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

				objectKey := getObjectKey(path)
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
					Bucket:        awsBucketName,
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

				// TODO: track uploaded successfully files to allow continuing
			}
		}()
	}

	wg.Wait()
}
