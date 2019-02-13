package upload

import (
	"bufio"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/zemirco/uid"
)

var (
	uploader *s3manager.Uploader
)

// Init initialises the uploader.
func Init() (err error) {
	session, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-2"),
	})
	if err != nil {
		return
	}

	uploader = s3manager.NewUploader(session)
	return
}

// Image uploads an image to Amazon S3.
func Image(file *multipart.FileHeader) (location string, err error) {
	image, err := file.Open()
	defer image.Close()
	if err != nil {
		return
	}

	imageID := uid.New(32)
	fileName := imageID + filepath.Ext(file.Filename)

	// Upload file to S3
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("froogo-ap"),        // Bucket name to upload to.
		Key:    aws.String("post/" + fileName), // Directory to upload to.
		Body:   bufio.NewReader(image),         // Body to upload (just bytes).
		ACL:    aws.String("public-read"),      // Set to public read (no key required to read).
	})
	if err != nil {
		return
	}

	url, err := url.Parse(result.Location)
	if err != nil {
		return
	}
	urlSplit := strings.Split(url.Path, "/")

	location = urlSplit[len(urlSplit)-1]
	return
}
