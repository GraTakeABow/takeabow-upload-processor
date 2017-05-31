package video

import (
	"fmt"
	"os/exec"
	"strings"
)

// YoutubeVideo denotes a VideoRequest that you can perform Youtube specific things on
type YoutubeVideo struct {
	*VideoRequest
}

func NewYoutubeVideo(r *VideoRequest) *YoutubeVideo {
	return &YoutubeVideo{r}
}

func (v *YoutubeVideo) HasVideo() (bool, error) {
	args := strings.Split(fmt.Sprintf("youtube-dl -f 137/136/22/mp4 -s %s", v.Url), " ")
	cmd := exec.Command(args[0], args[1:]...)

	_, err := runCommand(cmd)

	if err != nil {
		return false, err
	}

	return true, err
}

func (v *YoutubeVideo) GetVideo(dir string) (string, error) {
	dest := dir + "/" + v.Id + ".mp4"
	args := strings.Split(fmt.Sprintf("youtube-dl -f youtube-dl -f 137/136/22/mp4 -o %s %s", dest, v.Url), " ")
	cmd := exec.Command(args[0], args[1:]...)
	_, err := runCommand(cmd)

	if err != nil {
		return "", err
	}

	return dest, err
}

func (v *YoutubeVideo) GetRequest() *VideoRequest {
	return v.VideoRequest
}
