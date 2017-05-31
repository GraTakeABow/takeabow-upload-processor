package processor

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/therealpenguin/takeabow-upload-processor/video"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Processor struct {
	sess        *session.Session
	dir         string
	bucket      string
	prefix      string
	smallPrefix string
}

func New(sess *session.Session, dir, bucket, prefix, smallPrefix string) *Processor {
	return &Processor{
		sess:        sess,
		dir:         dir,
		bucket:      bucket,
		prefix:      prefix,
		smallPrefix: smallPrefix,
	}
}

// Process gets the video into a file in a temporary directory, transcodes it into the format we want and uploads it to S3
func (p *Processor) Process(v video.Video) error {
	r := v.GetRequest()
	fmt.Printf("Processing %s video\n%+v\n", r.GetSource(), r)
	hasFile, err := v.HasVideo()
	if err != nil {
		return err
	}

	if !hasFile {
		return errors.New(fmt.Sprintf("Video %s has no video", r.Url))
	}

	location, err := v.GetVideo(p.dir)

	if err != nil {
		return errors.New(fmt.Sprintf("Error getting video %s: %s", r.Url, err.Error()))
	}

	f, err := os.Open(location)
	if err != nil {
		return errors.New(fmt.Sprintf("Error opening %s: %s", location, err))
	}
	defer f.Close()
	defer os.Remove(f.Name())

	err = p.processFile(f, r.Id)
	if err != nil {
		return err
	}

	r.Duration = p.getDurationInSeconds(f.Name())

	return nil
}

func (p *Processor) getFramerate(f *os.File) (string, error) {
	args := strings.Split(fmt.Sprintf("ffprobe -v error -select_streams v:0 -show_entries stream=avg_frame_rate -of default=noprint_wrappers=1:nokey=1 %s", f.Name()), " ")
	cmd := exec.Command(args[0], args[1:]...)
	output, err := runCommand(cmd)
	if err != nil {
		return "", err
	}
	framerate := strings.Replace(string(output), "\n", "", -1)

	return framerate, nil
}

// processFile performs all the transcoding and uploading of a video file
// It turns an input video into the same video format
// It uploads the out of that onto S3
// It processes that output video into a small video, ready for other tasks
func (p *Processor) processFile(f *os.File, id string) error {
	// Get the input framerate
	framerate, err := p.getFramerate(f)
	if err != nil {
		framerate = "25"
	}

	// Scale the video to 1080p, 16/9, 24fps, libx24, yuv420p and remove audio
	command := "ffmpeg -y -r %s -i %s -filter:v scale=1920:1080,setdar=16/9 -r 24 -c:v libx264 -pix_fmt yuv420p -an %s"
	destination := f.Name() + "-processed.mp4"
	args := strings.Split(fmt.Sprintf(command, framerate, f.Name(), destination), " ")

	cmd := exec.Command(args[0], args[1:]...)
	_, err = runCommand(cmd)

	if err != nil {
		return err
	}

	defer os.Remove(destination)

	processed, err := os.Open(destination)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s/%s.mp4", p.prefix, id)

	manager := s3manager.NewUploader(p.sess)
	_, err = manager.Upload(&s3manager.UploadInput{
		Key:    aws.String(key),
		Bucket: aws.String(p.bucket),
		Body:   processed,
	})

	if err != nil {
		return err
	}

	err = p.uploadSmallVideo(processed, id)

	if err != nil {
		return err
	}

	return nil
}

func (p *Processor) uploadSmallVideo(f *os.File, id string) error {
	// make the video small for other types of processing
	command := "ffmpeg -y -r 24 -i %s -filter:v scale=320:240 %s"
	destination := f.Name() + "-small.mp4"
	args := strings.Split(fmt.Sprintf(command, f.Name(), destination), " ")

	cmd := exec.Command(args[0], args[1:]...)
	_, err := runCommand(cmd)

	if err != nil {
		return err
	}

	defer os.Remove(destination)

	processed, err := os.Open(destination)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s/%s.mp4", p.smallPrefix, id)

	manager := s3manager.NewUploader(p.sess)
	_, err = manager.Upload(&s3manager.UploadInput{
		Key:    aws.String(key),
		Bucket: aws.String(p.bucket),
		Body:   processed,
	})

	if err != nil {
		return err
	}

	return nil
}

func (p *Processor) getDurationInSeconds(filename string) int {
	args := strings.Split(fmt.Sprintf(`ffprobe -i %s -show_entries format=duration -v quiet -of csv=p=0`, filename), " ")
	cmd := exec.Command(args[0], args[1:]...)
	output, err := runCommand(cmd)
	if err != nil {
		log.Println(err)
		return 0
	}

	f, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)

	if err != nil {
		log.Println(err)
		return 0
	}

	return int(f)
}

// runCommand runs a cmd and gets the output (if any) and error (if any)
func runCommand(cmd *exec.Cmd) ([]byte, error) {
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("%s returned an error: %s", cmd.Args, err)
	}
	return output, err
}
