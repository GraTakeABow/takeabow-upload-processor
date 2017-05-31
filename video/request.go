package video

import (
	"database/sql"
	"encoding/json"
	"regexp"
	"time"
)

// Source represents the current location of the video
type Source string

const SourceYoutube Source = "youtube"
const SourceVimeo Source = "vimeo"
const SourceS3 Source = "s3"

// VideoRequest is the minimal information we need to perform processing
type VideoRequest struct {
	Id       string `json:"id"`
	Url      string `json:"url"`
	Status   string `json:"string"`
	Duration int    `json:"duration"`
}

// NewVideoRequest creates a VideoRequest object from a byte array. It attempts to get the source of the video
func NewVideoRequest(b []byte) (*VideoRequest, error) {
	v := VideoRequest{}
	err := json.Unmarshal(b, &v)

	if err != nil {
		return nil, err
	}

	return &v, nil
}

// GetSource sets a source of the video, or error if it cannot be decide
func (v *VideoRequest) GetSource() Source {
	if v.Url == "" {
		return ""
	}

	var match bool

	match, _ = regexp.MatchString(`s3.*\.*amazonaws\.com`, v.Url)

	if match {
		return SourceS3
	}

	match, _ = regexp.MatchString(
		`http(?:s?):\/\/(?:www\.)?youtu(?:be\.com\/watch\?v=|\.be\/)([\w\-\_]*)(&(amp;)?‌​[\w\?‌​=]*)?`,
		v.Url)

	if match {
		return SourceYoutube
	}

	match, _ = regexp.MatchString(
		`vimeo\.com`,
		v.Url)

	if match {
		return SourceVimeo
	}

	return ""
}

// SetStatus sets the status of the video and saves it
func (v *VideoRequest) SetStatus(status string, db *sql.DB) error {
	query := `UPDATE videos SET status = ?, updated_at = ? WHERE id = ?`
	_, err := db.Exec(query, status, time.Now(), v.Id)

	return err
}

func (v *VideoRequest) SetOriginalUrl(db *sql.DB) error {
	query := `UPDATE videos SET original_url = ? WHERE id = ?`
	_, err := db.Exec(query, v.Url, v.Id)

	return err
}

func (v *VideoRequest) SaveDuration(db *sql.DB) error {
	query := `UPDATE videos SET duration = ? WHERE id = ?`
	_, err := db.Exec(query, v.Duration, v.Id)

	return err
}
