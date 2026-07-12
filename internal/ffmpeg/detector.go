package ffmpeg

import (
	"errors"
	"os/exec"
	"strings"
)

var ErrNotFound = errors.New("ffmpeg not found — install with: brew install ffmpeg")

type Detector struct {
	FFmpegPath  string
	FFprobePath string
}

func Detect() (*Detector, error) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, ErrNotFound
	}
	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, ErrNotFound
	}
	return &Detector{FFmpegPath: ffmpegPath, FFprobePath: ffprobePath}, nil
}

func (d *Detector) Version() string {
	out, err := exec.Command(d.FFmpegPath, "-version").Output()
	if err != nil {
		return "unknown"
	}
	lines := strings.SplitN(string(out), "\n", 2)
	if len(lines) > 0 {
		fields := strings.Fields(lines[0])
		if len(fields) >= 3 {
			return fields[2]
		}
	}
	return "unknown"
}
