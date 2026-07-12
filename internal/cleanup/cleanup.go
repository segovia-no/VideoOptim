package cleanup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Result struct {
	Moved   int      `json:"moved"`
	Deleted int      `json:"deleted"`
	Errors  []string `json:"errors,omitempty"`
}

type Pair struct {
	Original  string
	Optimized string
}

func FindPairs(optimizedPaths []string) []Pair {
	var pairs []Pair
	for _, opt := range optimizedPaths {
		if opt == "" {
			continue
		}
		orig := originalFor(opt)
		if orig == "" {
			continue
		}
		if _, err := os.Stat(orig); err == nil {
			pairs = append(pairs, Pair{Original: orig, Optimized: opt})
		}
	}
	return pairs
}

func originalFor(optimized string) string {
	base := filepath.Base(optimized)
	dir := filepath.Dir(optimized)
	if !strings.HasSuffix(base, "_optimized.mp4") {
		return ""
	}
	stem := strings.TrimSuffix(base, "_optimized.mp4")
	for _, ext := range []string{"mp4", "MP4", "mov", "MOV", "mkv", "MKV", "avi", "AVI", "webm", "WEBM"} {
		candidate := filepath.Join(dir, stem+"."+ext)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func Run(pairs []Pair) Result {
	r := Result{}
	for _, p := range pairs {
		origInfo, err := os.Stat(p.Original)
		if err != nil {
			r.Errors = append(r.Errors, fmt.Sprintf("stat %s: %v", p.Original, err))
			continue
		}
		optInfo, err := os.Stat(p.Optimized)
		if err != nil {
			r.Errors = append(r.Errors, fmt.Sprintf("stat %s: %v", p.Optimized, err))
			continue
		}

		if optInfo.Size() < origInfo.Size() {
			if err := moveToTrash(p.Original); err != nil {
				r.Errors = append(r.Errors, fmt.Sprintf("trash %s: %v", p.Original, err))
			} else {
				r.Moved++
			}
		} else {
			if err := os.Remove(p.Optimized); err != nil {
				r.Errors = append(r.Errors, fmt.Sprintf("remove %s: %v", p.Optimized, err))
			} else {
				r.Deleted++
			}
		}
	}
	return r
}

func moveToTrash(path string) error {
	script := fmt.Sprintf(`tell application "Finder" to delete POSIX file %q`, path)
	return exec.Command("osascript", "-e", script).Run()
}
