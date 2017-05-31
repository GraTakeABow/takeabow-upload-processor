package video

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"os/exec"
)

// Video is an interface that allows different sources of videos to say how to get a video file
type Video interface {
	HasVideo() (bool, error)
	GetVideo(dir string) (string, error)
	GetRequest() *VideoRequest
}

func New(b []byte, s3 *s3.S3, sess *session.Session, bucket string) (Video, error) {
	r, err := NewVideoRequest(b)
	if err != nil {
		return nil, err
	}

	switch r.GetSource() {
	case SourceS3:
		return NewS3Video(r, s3, sess, bucket), nil
	case SourceYoutube:
		return NewYoutubeVideo(r), nil
	case SourceVimeo:
		return NewVimeoVideo(r), nil
	}
	return nil, errors.New(fmt.Sprintf("VideoRequest %s does not have a valid source", r.Id))
}

func runCommand(cmd *exec.Cmd) ([]byte, error) {
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Println(cmd.Args)
		log.Println(err)
	}
	return output, err
}
