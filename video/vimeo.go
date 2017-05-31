package video

import (
	"fmt"
	"os/exec"
	"strings"
)

// VimeoVideo denotes a VideoRequest that you can perform S3 specific things on
type VimeoVideo struct {
	*VideoRequest
}

func NewVimeoVideo(r *VideoRequest) *VimeoVideo {
	return &VimeoVideo{r}
}

func (v *VimeoVideo) HasVideo() (bool, error) {
	args := strings.Split(fmt.Sprintf("youtube-dl -f http-1080p/http-720p/mp4 -s %s", v.Url), " ")
	cmd := exec.Command(args[0], args[1:]...)

	_, err := runCommand(cmd)

	if err != nil {
		return false, err
	}

	return true, err

}

func (v *VimeoVideo) GetVideo(dir string) (string, error) {
	dest := dir + "/" + v.Id + ".mp4"
	args := strings.Split(fmt.Sprintf("youtube-dl -f http-720p/mp4 -o %s %s", dest, v.Url), " ")
	cmd := exec.Command(args[0], args[1:]...)
	_, err := runCommand(cmd)

	if err != nil {
		return "", err
	}

	return dest, err
}

func (v *VimeoVideo) GetRequest() *VideoRequest {
	return v.VideoRequest
}
