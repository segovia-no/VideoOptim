package ffmpeg

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

type VideoInfo struct {
	Duration float64 // seconds
	Width    int
	Height   int
	Codec    string
	Size     int64 // bytes
}

type probeOutput struct {
	Streams []probeStream `json:"streams"`
	Format  probeFormat  `json:"format"`
}

type probeStream struct {
	CodecType string `json:"codec_type"`
	CodecName string `json:"codec_name"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Duration  string `json:"duration"`
}

type probeFormat struct {
	Duration string `json:"duration"`
	Size     string `json:"size"`
}

func (d *Detector) Probe(path string) (*VideoInfo, error) {
	out, err := exec.Command(d.FFprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"-show_format",
		path,
	).Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe: %w", err)
	}

	var p probeOutput
	if err := json.Unmarshal(out, &p); err != nil {
		return nil, fmt.Errorf("ffprobe parse: %w", err)
	}

	info := &VideoInfo{}
	dur, _ := strconv.ParseFloat(p.Format.Duration, 64)
	info.Duration = dur
	size, _ := strconv.ParseInt(p.Format.Size, 10, 64)
	info.Size = size

	for _, s := range p.Streams {
		if s.CodecType == "video" {
			info.Width = s.Width
			info.Height = s.Height
			info.Codec = s.CodecName
			if info.Duration == 0 {
				d, _ := strconv.ParseFloat(s.Duration, 64)
				info.Duration = d
			}
			break
		}
	}

	return info, nil
}
