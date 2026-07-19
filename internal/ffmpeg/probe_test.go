package ffmpeg

import (
	"testing"
)

// Real ffmpeg -i stderr output, trimmed to relevant sections.
const fixtureH264 = `ffmpeg version 7.1.1 Copyright (c) 2000-2025 the FFmpeg developers
Input #0, mov,mp4,m4a,3gp,3g2,mj2, from 'clip.mp4':
  Metadata:
    major_brand     : isom
  Duration: 00:01:30.50, start: 0.000000, bitrate: 5000 kb/s
    Stream #0:0(und): Video: h264 (High) (avc1 / 0x31637661), yuv420p, 1920x1080 [SAR 1:1 DAR 16:9], 4800 kb/s, 30 fps
    Stream #0:1(und): Audio: aac (LC), 48000 Hz, stereo, fltp, 192 kb/s`

const fixtureHEVC = `ffmpeg version 7.1.1
Input #0, mov,mp4,m4a,3gp,3g2,mj2, from 'clip.mov':
  Duration: 00:00:45.00, start: 0.000000, bitrate: 8000 kb/s
    Stream #0:0: Video: hevc (Main), yuv420p, 3840x2160, 25 fps
    Stream #0:1: Audio: aac, 48000 Hz, stereo`

const fixtureVP9 = `ffmpeg version 7.1.1
Input #0, matroska,webm, from 'clip.mkv':
  Duration: 00:02:00.00, start: 0.000000, bitrate: 3000 kb/s
    Stream #0:0: Video: vp9, yuv420p, 1280x720, 24 fps`

const fixtureNoVideo = `ffmpeg version 7.1.1
Input #0, mp3, from 'audio.mp3':
  Duration: 00:03:00.00, start: 0.000000, bitrate: 128 kb/s
    Stream #0:0: Audio: mp3, 44100 Hz, stereo`

const fixtureHoursLong = `ffmpeg version 7.1.1
Input #0, matroska, from 'long.mkv':
  Duration: 02:30:00.00, start: 0.000000, bitrate: 1000 kb/s
    Stream #0:0: Video: h264, yuv420p, 1920x1080`

func TestParseProbeText(t *testing.T) {
	t.Run("parses h264 codec and dimensions", func(t *testing.T) {
		info, err := parseProbeText(fixtureH264)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Codec != "h264" {
			t.Errorf("Codec = %q, want %q", info.Codec, "h264")
		}
		if info.Width != 1920 || info.Height != 1080 {
			t.Errorf("dims = %dx%d, want 1920x1080", info.Width, info.Height)
		}
	})

	t.Run("parses duration correctly", func(t *testing.T) {
		info, err := parseProbeText(fixtureH264)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// 00:01:30.50 = 90.5 seconds
		if info.Duration != 90.5 {
			t.Errorf("Duration = %v, want 90.5", info.Duration)
		}
	})

	t.Run("parses hevc codec and 4K dimensions", func(t *testing.T) {
		info, err := parseProbeText(fixtureHEVC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Codec != "hevc" {
			t.Errorf("Codec = %q, want %q", info.Codec, "hevc")
		}
		if info.Width != 3840 || info.Height != 2160 {
			t.Errorf("dims = %dx%d, want 3840x2160", info.Width, info.Height)
		}
		// 00:00:45.00 = 45 seconds
		if info.Duration != 45.0 {
			t.Errorf("Duration = %v, want 45.0", info.Duration)
		}
	})

	t.Run("parses vp9 and 720p dimensions", func(t *testing.T) {
		info, err := parseProbeText(fixtureVP9)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Codec != "vp9" {
			t.Errorf("Codec = %q, want %q", info.Codec, "vp9")
		}
		if info.Width != 1280 || info.Height != 720 {
			t.Errorf("dims = %dx%d, want 1280x720", info.Width, info.Height)
		}
	})

	t.Run("parses multi-hour duration", func(t *testing.T) {
		info, err := parseProbeText(fixtureHoursLong)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// 02:30:00.00 = 9000 seconds
		if info.Duration != 9000.0 {
			t.Errorf("Duration = %v, want 9000.0", info.Duration)
		}
	})

	t.Run("errors when no video stream", func(t *testing.T) {
		_, err := parseProbeText(fixtureNoVideo)
		if err == nil {
			t.Error("expected error for audio-only file, got nil")
		}
	})

	t.Run("errors on empty input", func(t *testing.T) {
		_, err := parseProbeText("")
		if err == nil {
			t.Error("expected error for empty input, got nil")
		}
	})
}
