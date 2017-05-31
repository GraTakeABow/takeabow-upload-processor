package video

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"net/url"
	"os"
)

// S3Video denotes a VideoRequest that you can perform S3 specific things on
type S3Video struct {
	*VideoRequest
	s3     *s3.S3
	sess   *session.Session
	bucket string
}

func NewS3Video(r *VideoRequest, s3 *s3.S3, sess *session.Session, bucket string) *S3Video {
	return &S3Video{r, s3, sess, bucket}
}

// HasVideo checks to see if the S3 key exists
func (v *S3Video) HasVideo() (bool, error) {
	url, err := url.Parse(v.GetRequest().Url)

	if err != nil {
		return false, err
	}

	params := s3.HeadObjectInput{
		Bucket: aws.String(v.bucket),
		Key:    aws.String(url.Path),
	}

	_, err = v.s3.HeadObject(&params)

	if err != nil {
		return false, err
	}

	return true, err
}

func (v *S3Video) GetVideo(dir string) (string, error) {
	dest := dir + "/" + v.Id
	url, err := url.Parse(v.GetRequest().Url)

	if err != nil {
		return "", err
	}

	// Create a downloader with the session and default options
	downloader := s3manager.NewDownloader(v.sess)

	// Create a file to write the S3 Object contents to.
	f, err := os.Create(dest)
	defer f.Close()

	if err != nil {
		return "", fmt.Errorf("failed to create file %q, %v", dest, err)
	}

	// Write the contents of S3 Object to the file
	_, err = downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(v.bucket),
		Key:    aws.String(url.Path),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}

	return dest, nil
}

func (v *S3Video) GetRequest() *VideoRequest {
	return v.VideoRequest
}
