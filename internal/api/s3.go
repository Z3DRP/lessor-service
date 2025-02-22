package api

import (
	"context"
	"fmt"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	s3Config    *aws.Config
	S3Client    *s3.Client
	bucket      = os.Getenv("S3_BUCKET")
	region      = os.Getenv("AWS_REGION")
	secretId    = os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey   = os.Getenv("AWS_SECRET_ACCESS_KEY")
	userDir     = os.Getenv("USERS_DIR")
	proeprtyDir = os.Getenv("PROPERTIES_DIR")
	taskDir     = os.Getenv("TASKS_DIR")
	SetupErr    error
	maxSize     = int64(1024000)
	once        sync.Once
)

type Uploader interface {
	Upload(*http.Request) (string, error)
}

type Getter interface {
	Get(context.Context, string) ([]string, error)
}

type S3Actor struct {
	dir       string
	subdir    string // will be identifier like uuid of struct
	client    *s3.Client
	StartedAt time.Time
	StoppedAt time.Time
}

func NewS3Actor(ctx context.Context, dir, subdir string) (*S3Actor, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))

	if err != nil {
		log.Printf("failed to setup s3, %v", err)
		return nil, err
	}

	return &S3Actor{
		client:    s3.NewFromConfig(cfg),
		StartedAt: time.Now(),
	}, nil
}

func (a *S3Actor) UplaodDir() string {
	objDir := determineDir(a.dir)
	if objDir == "" {
		return ""
	}

	return filepath.Join(objDir, a.subdir)
}

func (a S3Actor) Upload(r *http.Request) (string, error) {
	err := r.ParseMultipartForm(maxSize)

	if err != nil {
		return "", ErrMaxSize{Err: err}
	}

	file, header, err := r.FormFile("image")

	if err != nil {
		return "", ErrFileRead{Err: err}
	}

	defer file.Close()
	objDir := a.UplaodDir()

	if objDir == "" {
		return "", ErrInvalidBucketDir{InvalidDir: a.dir}
	}

	tstampFilename := fmt.Sprintf("%v%v", header.Filename, time.Now().UnixNano())
	fileNameKey := filepath.Join(objDir, tstampFilename)
	_, err = a.client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileNameKey),
		Body:   file,
		ACL:    "public-read",
	})

	if err != nil {
		return "", ErrFileObjUpload{Err: err}
	}

	return fileNameKey, nil
}

func (a S3Actor) Getter(ctx context.Context, fileName string) ([]string, error) {
	// said to do full key name ? bucket /folder/fileName.ext ?
	// result, err := a.client.GetObject(ctx, &s3.GetObjectInput{
	// 	Bucket: aws.String(bucket),
	// 	Key:    aws.String(fileName),
	// })
	objDir := a.UplaodDir()

	if objDir == "" {
		return nil, ErrInvalidBucketDir{InvalidDir: a.dir}
	}

	res, err := a.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(objDir),
	})

	if err != nil {
		return nil, err
	}

	var imgs []string
	for _, item := range res.Contents {
		presignedClient := s3.NewPresignClient(a.client)
		presignedUrl, err := presignedClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(*item.Key),
		}, s3.WithPresignExpires(24*time.Hour))

		if err != nil {
			continue
		}

		imgs = append(imgs, presignedUrl.URL)
	}

	return imgs, nil
}

func determineDir(loc string) string {
	switch strings.ToLower(loc) {
	case "property":
		return proeprtyDir
	case "task":
		return taskDir
	case "user":
		return userDir
	default:
		return ""
	}
}
