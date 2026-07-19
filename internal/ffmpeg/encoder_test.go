package ffmpeg

import (
	"testing"
)

func TestOutputPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/videos/clip.mp4", "/videos/clip_optimized.mp4"},
		{"/videos/clip.mov", "/videos/clip_optimized.mp4"},
		{"/videos/clip.mkv", "/videos/clip_optimized.mp4"},
		{"/videos/clip.avi", "/videos/clip_optimized.mp4"},
		{"/videos/clip.webm", "/videos/clip_optimized.mp4"},
		{"/videos/clip.MP4", "/videos/clip_optimized.mp4"},
		{"/videos/my video.mp4", "/videos/my video_optimized.mp4"},
		{"/deep/nested/path/clip.mp4", "/deep/nested/path/clip_optimized.mp4"},
		{"clip.mp4", "clip_optimized.mp4"},
		// file with dots in name
		{"/videos/my.film.2024.mp4", "/videos/my.film.2024_optimized.mp4"},
	}

	for _, tt := range tests {
		got := OutputPath(tt.input)
		if got != tt.want {
			t.Errorf("OutputPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
