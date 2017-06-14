package timecode

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

// Timecode represents a section of video that we should cut to
type Timecode struct {
	Length float64
}

// NewFromPath uses a reader and converts the first column into a Timecode
func NewFromFile(r io.Reader) (*[]Timecode, error) {
	reader := csv.NewReader(r)
	timecodes := make([]Timecode, 0)
	row := 0

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if len(record) < 1 {
			return nil, fmt.Errorf("Row %d of timecode csv is empty", row)
		}
		len, err := strconv.ParseFloat(record[0], 64)
		timecodes = append(timecodes, Timecode{
			Length: len,
		})
		row++
	}

	return &timecodes, nil
}
