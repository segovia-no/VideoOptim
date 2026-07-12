package ffmpeg

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var ErrNotFound = errors.New("ffmpeg not found — bundle it via `make ffmpeg` or install with: brew install ffmpeg")

type Detector struct {
	FFmpegPath string
}

func Detect() (*Detector, error) {
	ffmpegPath := findBin("ffmpeg")
	if ffmpegPath == "" {
		return nil, ErrNotFound
	}
	return &Detector{FFmpegPath: ffmpegPath}, nil
}

// findBin checks alongside the running executable first (bundled .app),
// then falls back to PATH.
func findBin(name string) string {
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	return ""
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
