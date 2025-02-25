package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Z3DRP/lessor-service/internal/ztype"
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
	Upload(context.Context, string, string, *ztype.FileUploadDto) (string, error)
}

type Getter interface {
	Get(context.Context, string, string, string) (string, error)
	List(context.Context, string) (map[string]string, error)
}

// TODO update this to be more testable and use these smaller interfaces

type FilePersister interface {
	Uploader
	Getter
}

type S3ObjectGetter interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type S3ObjectUploader interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type S3ObjectLister interface {
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

type S3Actor struct {
	dir       string
	client    *s3.Client
	StartedAt time.Time
	StoppedAt time.Time
}

func NewS3Actor(ctx context.Context, objDir string) (S3Actor, error) {
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
		dir:       objDir,
		client:    s3.NewFromConfig(cfg),
		StartedAt: time.Now(),
	}, nil
}

func (a *S3Actor) UplaodDir(ownerId, objId string) (string, error) {
	objDir := determineDir(a.dir)
	if objDir == "" {
		return "", ErrInvalidBucketDir{InvalidDir: a.dir}
	}
	// prefix for bucket dir: obj/owner-id/obj-id
	return filepath.Join(objDir, ownerId, objId), nil
}

func (a S3Actor) Upload(ctx context.Context, ownerId, objId string, file *ztype.FileUploadDto) (string, error) {
	objDir, err := a.UplaodDir(ownerId, objId)

	if err != nil {
		return "", err
	}

	// final bucket dir obj/lessor-id/obj-id/tstamp-Filename.ext
	tstampFilename := fmt.Sprintf("%v-%v", time.Now().UnixNano(), file.Header.Filename)
	fileNameKey := filepath.Join(objDir, tstampFilename)
	_, err = a.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileNameKey),
		Body:   file.File,
		ACL:    "public-read",
	})

	if err != nil {
		return "", ErrFileObjUpload{Err: err}
	}

	return tstampFilename, nil
}

func (a S3Actor) List(ctx context.Context, ownerId string) (map[string]string, error) {
	// said to do full key name ? bucket /folder/fileName.ext ?
	// result, err := a.client.GetObject(ctx, &s3.GetObjectInput{
	// 	Bucket: aws.String(bucket),
	// 	Key:    aws.String(fileName),
	// })

	// in theory this should get everything say under properties/ownerId/
	objDir := determineDir(a.dir)
	if objDir == "" {
		return nil, ErrInvalidBucketDir{InvalidDir: a.dir}
	}

	fileKey := filepath.Join(objDir, ownerId)

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

func (s S3Actor) Get(ctx context.Context, ownerId, objId string, fileKey string) (string, error) {
	psClient := s3.NewPresignClient(s.client)

	// the obj dir path is obj/owner-id/obj-id/filename
	// which is saved in the db field for obj which is passed into func
	objDir, err := s.UplaodDir(ownerId, objId)

	if err != nil {
		return "", err
	}

	fileObjKey := filepath.Join(objDir, fileKey)
	psUrl, err := psClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileObjKey),
	}, s3.WithPresignExpires(expireyHrs*time.Hour))

	if err != nil {
		return "", fmt.Errorf("failed to create image url for file %v", err)
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
