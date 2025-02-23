package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const expireyHrs = 3

var (
	bucket      = os.Getenv("S3_BUCKET")
	region      = os.Getenv("AWS_REGION")
	secretId    = os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey   = os.Getenv("AWS_SECRET_ACCESS_KEY")
	userDir     = os.Getenv("USERS_DIR")
	proeprtyDir = os.Getenv("PROPERTIES_DIR")
	taskDir     = os.Getenv("TASKS_DIR")
	SetupErr    error
	maxSize     = int64(1024000)
)

type Uploader interface {
	Upload(*http.Request) (string, error)
}

type Getter interface {
	Get(context.Context, string) ([]string, error)
	GetAll(context.Context, string) (string, error)
}

type S3Actor struct {
	dir       string
	client    *s3.Client
	StartedAt time.Time
	StoppedAt time.Time
}

func NewS3Actor(ctx context.Context, dir string) (S3Actor, error) {
	// loadDefaultConfig load all env variables it checks the aws shared dir then env vars but
	// i want to only use env vars so use credProvider
	// if this doesnt work then do this
	// err := godotenv.Load()
	// if err != nil {
	// return err
	//}
	// cfg, err := config.LoadDefaulConfig(ctx, config.WithRegion(region))
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			secretId,
			secretKey,
			"",
		))),
		config.WithRegion(region),
	)

	if err != nil {
		log.Printf("failed to setup s3, %v", err)
		return S3Actor{}, err
	}

	return S3Actor{
		dir:       dir,
		client:    s3.NewFromConfig(cfg),
		StartedAt: time.Now(),
	}, nil
}

func (a *S3Actor) UplaodDir(ownerId, objId string) string {
	objDir := determineDir(a.dir)
	if objDir == "" {
		return ""
	}
	// prefix for bucket dir: obj/owner-id/obj-id
	return filepath.Join(objDir, ownerId, objId)
}

func (a S3Actor) Upload(r *http.Request, ownerId, objId string) (string, error) {
	err := r.ParseMultipartForm(maxSize)

	if err != nil {
		return "", ErrMaxSize{Err: err}
	}

	file, header, err := r.FormFile("image")

	if err != nil {
		return "", ErrFileRead{Err: err}
	}

	defer file.Close()
	objDir := a.UplaodDir(ownerId, objId)

	if objDir == "" {
		return "", ErrInvalidBucketDir{InvalidDir: a.dir}
	}

	// final bucket dir obj/lessor-id/obj-id/tstamp-Filename.ext
	tstampFilename := fmt.Sprintf("%v-%v", time.Now().UnixNano(), header.Filename)
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

func (a S3Actor) GetAll(ctx context.Context, lessorId string) (map[string]string, error) {
	// said to do full key name ? bucket /folder/fileName.ext ?
	// result, err := a.client.GetObject(ctx, &s3.GetObjectInput{
	// 	Bucket: aws.String(bucket),
	// 	Key:    aws.String(fileName),
	// })

	// in theory this should get everything say under properties/lessorId/
	objDir := determineDir(a.dir)
	if objDir == "" {
		return nil, ErrInvalidBucketDir{InvalidDir: a.dir}
	}

	fileKey := filepath.Join(objDir, lessorId)

	res, err := a.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(fileKey),
	})

	if err != nil {
		return nil, err
	}

	var imgs = make(map[string]string)
	for _, item := range res.Contents {
		presignedClient := s3.NewPresignClient(a.client)
		presignedUrl, err := presignedClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(*item.Key),
		}, s3.WithPresignExpires(expireyHrs*time.Hour))

		if err != nil {
			continue
		}

		imgs[*item.Key] = presignedUrl.URL
	}

	return imgs, nil
}

func (s *S3Actor) Get(ctx context.Context, fileKey string) (string, error) {
	psClient := s3.NewPresignClient(s.client)

	// the obj dir path is obj/owner-id/obj-id/filename
	// which is saved in the db field for obj which is passed into func
	psUrl, err := psClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileKey),
	}, s3.WithPresignExpires(expireyHrs*time.Hour))

	if err != nil {
		return "", fmt.Errorf("failed to create image url %v", err)
	}

	return psUrl.URL, nil
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
