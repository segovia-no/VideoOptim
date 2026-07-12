package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type VideoInfo struct {
	Duration float64 // seconds
	Width    int
	Height   int
	Codec    string
	Size     int64 // bytes
}

var (
	reDuration = regexp.MustCompile(`Duration:\s+(\d+):(\d+):(\d+(?:\.\d+)?)`)
	reVideo    = regexp.MustCompile(`Video:\s+(\w+)`)
	reDims     = regexp.MustCompile(`(\d{2,5})x(\d{2,5})`)
)

func (d *Detector) Probe(path string) (*VideoInfo, error) {
	// ffmpeg -i with no output prints stream info to stderr and exits non-zero.
	cmd := exec.Command(d.FFmpegPath, "-i", path)
	out, _ := cmd.CombinedOutput()

	info, err := parseProbeText(string(out))
	if err != nil {
		return nil, fmt.Errorf("ffmpeg probe: no video stream in %s", path)
	}

	if fi, err := os.Stat(path); err == nil {
		info.Size = fi.Size()
	}

	return info, nil
}

// parseProbeText parses the combined output of ffmpeg -i and extracts video
// stream metadata. Pure function — no filesystem or process access.
func parseProbeText(text string) (*VideoInfo, error) {
	info := &VideoInfo{}

	if m := reDuration.FindStringSubmatch(text); m != nil {
		h, _ := strconv.ParseFloat(m[1], 64)
		mn, _ := strconv.ParseFloat(m[2], 64)
		s, _ := strconv.ParseFloat(m[3], 64)
		info.Duration = h*3600 + mn*60 + s
	}

	for _, line := range strings.Split(text, "\n") {
		if !strings.Contains(line, "Video:") {
			continue
		}
		if m := reVideo.FindStringSubmatch(line); m != nil {
			info.Codec = m[1]
		}
		if m := reDims.FindStringSubmatch(line); m != nil {
			info.Width, _ = strconv.Atoi(m[1])
			info.Height, _ = strconv.Atoi(m[2])
		}
		break
	}

	if info.Codec == "" {
		return nil, fmt.Errorf("no video stream found")
	}

	return info, nil
}
