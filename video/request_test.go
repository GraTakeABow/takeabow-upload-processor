package video

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetSource(t *testing.T) {
	type TestCase struct {
		Url    string
		Source Source
	}

	testcases := []TestCase{
		{
			Url:    "http://www.youtube.com/watch?v=-wtIMTCHWuI",
			Source: SourceYoutube,
		},
		{
			Url:    "https://takeabow.s3.amazonaws.com/upload/foo.mp4",
			Source: SourceS3,
		},
	}

	for _, tc := range testcases {
		r := VideoRequest{
			Url: tc.Url,
		}
		actual := r.GetSource()
		assert.Equal(t, tc.Source, actual)
	}
}
