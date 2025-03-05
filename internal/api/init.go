package api

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var (
	bucket    string
	region    string
	secretId  string
	secretKey string
	SetupErr  error
	maxSize   int64
	expirey   time.Duration
)

func init() {
	// load package level env
	if err := godotenv.Load(); err != nil {
		log.Printf("WARNING no .env file found this may cause problems")
	}
	bucket = os.Getenv("S3_BUCKET")
	region = os.Getenv("AWS_REGION")
	secretId = os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	maxSize = int64(1024000)
	expirey = 14400 * time.Second
}
