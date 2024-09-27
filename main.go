package main

import (
	"context"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var client *s3.Client

func createS3Client() error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}

	var s3endpoint string = os.Getenv("AWS_S3_ENDPOINT")

	client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s3endpoint)
	})

	return nil
}

func uploadFile(objectKey string, r io.Reader) error {
	var bucketName string = os.Getenv("AWS_S3_BUCKET_NAME")

	var objectKeyParts []string = strings.Split(objectKey, ".")
	var ext string = "." + objectKeyParts[len(objectKeyParts)-1]
	var contentType string = mime.TypeByExtension(ext)

	_, err := client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        r,
		ContentType: aws.String(contentType),
	})

	return err
}

func main() {
	err := createS3Client()
	if err != nil {
		panic(err)
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/file", postFileHandler)

	e.Logger.Fatal(e.Start(":1323"))
}

func postFileHandler(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	err = uploadFile(file.Filename, src)
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, "File uploaded")
}
