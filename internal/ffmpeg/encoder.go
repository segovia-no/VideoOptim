package ffmpeg

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"VideoOptim/internal/settings"
)

type ProgressUpdate struct {
	Percent float64
	Elapsed time.Duration
	FPS     float64
}

type EncodeResult struct {
	OutputPath   string // empty if no gain (original kept)
	OriginalSize int64
	OutputSize   int64
}

func OutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	stem := strings.TrimSuffix(inputPath, ext)
	return stem + "_optimized.mp4"
}

func (d *Detector) Encode(
	inputPath string,
	info *VideoInfo,
	s settings.Settings,
	onProgress func(ProgressUpdate),
	onPid func(int),
	cancelCh <-chan struct{},
) (*EncodeResult, error) {
	outputPath := OutputPath(inputPath)

	args := []string{
		"-i", inputPath,
		"-progress", "pipe:1",
		"-nostats",
		"-y",
	}

	switch s.Encoder {
	case settings.EncoderHevcVideoToolbox:
		args = append(args,
			"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2,format=yuv420p",
			"-c:v", "hevc_videotoolbox", "-q:v", "65",
		)
	default:
		args = append(args, "-c:v", "libx265", "-crf", strconv.Itoa(s.CRF))
	}

	if s.KeepAudio {
		args = append(args, "-c:a", "copy")
	} else {
		args = append(args, "-an")
	}

	args = append(args, outputPath)

	cmd := exec.Command(d.FFmpegPath, args...)

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("ffmpeg start: %w", err)
	}
	if onPid != nil {
		onPid(cmd.Process.Pid)
	}

	start := time.Now()
	doneCh := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(stdout)
		var fps float64
		for scanner.Scan() {
			line := scanner.Text()
			kv := strings.SplitN(line, "=", 2)
			if len(kv) != 2 {
				continue
			}
			key, val := kv[0], strings.TrimSpace(kv[1])
			switch key {
			case "out_time_ms":
				ms, _ := strconv.ParseFloat(val, 64)
				if info.Duration > 0 && ms > 0 {
					pct := (ms / 1e6) / info.Duration * 100
					if pct > 100 {
						pct = 100
					}
					onProgress(ProgressUpdate{
						Percent: pct,
						Elapsed: time.Since(start),
						FPS:     fps,
					})
				}
			case "fps":
				fps, _ = strconv.ParseFloat(val, 64)
			}
		}
	}()

	go func() {
		doneCh <- cmd.Wait()
	}()

	select {
	case err := <-doneCh:
		if err != nil {
			os.Remove(outputPath)
			return nil, fmt.Errorf("ffmpeg encode: %w\n%s", err, lastLines(stderrBuf.String(), 5))
		}
	case <-cancelCh:
		cmd.Process.Kill()
		os.Remove(outputPath)
		return nil, fmt.Errorf("cancelled")
	}

	origStat, err := os.Stat(inputPath)
	if err != nil {
		return nil, err
	}
	outStat, err := os.Stat(outputPath)
	if err != nil {
		return nil, err
	}

	if s.DiscardIfNoGain && outStat.Size() >= origStat.Size() {
		os.Remove(outputPath)
		return &EncodeResult{
			OriginalSize: origStat.Size(),
			OutputSize:   origStat.Size(),
		}, nil
	}

	return &EncodeResult{
		OutputPath:   outputPath,
		OriginalSize: origStat.Size(),
		OutputSize:   outStat.Size(),
	}, nil
}

func lastLines(s string, n int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
