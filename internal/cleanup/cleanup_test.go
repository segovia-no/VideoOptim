package cleanup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCandidatePaths(t *testing.T) {
	t.Run("returns nil for non-optimized filename", func(t *testing.T) {
		got := candidatePaths("/videos/clip.mp4")
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("returns nil when suffix absent", func(t *testing.T) {
		got := candidatePaths("/videos/clip_optimized_extra.mp4")
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("contains all expected extensions in order", func(t *testing.T) {
		dir := "/videos"
		got := candidatePaths(filepath.Join(dir, "clip_optimized.mp4"))
		wantExts := []string{"mp4", "MP4", "mov", "MOV", "mkv", "MKV", "avi", "AVI", "webm", "WEBM"}
		if len(got) != len(wantExts) {
			t.Fatalf("expected %d candidates, got %d", len(wantExts), len(got))
		}
		for i, ext := range wantExts {
			want := filepath.Join(dir, "clip."+ext)
			if got[i] != want {
				t.Errorf("candidate[%d] = %q, want %q", i, got[i], want)
			}
		}
	})

	t.Run("preserves directory", func(t *testing.T) {
		got := candidatePaths("/deep/nested/path/clip_optimized.mp4")
		for _, p := range got {
			if filepath.Dir(p) != "/deep/nested/path" {
				t.Errorf("wrong directory in candidate: %q", p)
			}
		}
	})

	t.Run("handles stem with dots", func(t *testing.T) {
		got := candidatePaths("/videos/my.film.2024_optimized.mp4")
		if len(got) == 0 {
			t.Fatal("expected candidates, got none")
		}
		if filepath.Base(got[0]) != "my.film.2024.mp4" {
			t.Errorf("unexpected first candidate: %q", got[0])
		}
	})

	t.Run("handles spaces in path", func(t *testing.T) {
		got := candidatePaths("/my videos/family clip_optimized.mp4")
		if len(got) == 0 {
			t.Fatal("expected candidates, got none")
		}
		if filepath.Base(got[0]) != "family clip.mp4" {
			t.Errorf("unexpected first candidate: %q", got[0])
		}
	})
}

// ── originalFor (filesystem) ───────────────────────────────────────────────

func TestOriginalFor(t *testing.T) {
	t.Run("returns empty for non-optimized path", func(t *testing.T) {
		got := originalFor("/videos/clip.mp4")
		if got != "" {
			t.Errorf("expected empty, got %q", got)
		}
	})

	t.Run("returns empty when no original exists on disk", func(t *testing.T) {
		dir := t.TempDir()
		got := originalFor(filepath.Join(dir, "clip_optimized.mp4"))
		if got != "" {
			t.Errorf("expected empty, got %q", got)
		}
	})

	// Case-insensitive filesystems (macOS APFS) match lowercase candidates first,
	// so we only assert a non-empty result — not the exact path string.
	for _, ext := range []string{"mp4", "MP4", "mov", "MOV", "mkv", "MKV", "avi", "AVI", "webm", "WEBM"} {
		ext := ext
		t.Run("finds ."+ext+" original", func(t *testing.T) {
			dir := t.TempDir()
			orig := filepath.Join(dir, "clip."+ext)
			if err := os.WriteFile(orig, []byte{}, 0644); err != nil {
				t.Fatal(err)
			}
			got := originalFor(filepath.Join(dir, "clip_optimized.mp4"))
			if got == "" {
				t.Errorf("expected non-empty result for .%s original, got empty", ext)
			}
		})
	}

	t.Run("returns first match when multiple extensions exist", func(t *testing.T) {
		dir := t.TempDir()
		// create both mp4 and mov originals — mp4 should win (first in list)
		os.WriteFile(filepath.Join(dir, "clip.mp4"), []byte{}, 0644)
		os.WriteFile(filepath.Join(dir, "clip.mov"), []byte{}, 0644)
		got := originalFor(filepath.Join(dir, "clip_optimized.mp4"))
		want := filepath.Join(dir, "clip.mp4")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

// ── FindPairs (filesystem) ─────────────────────────────────────────────────

func TestFindPairs(t *testing.T) {
	t.Run("returns nil for nil input", func(t *testing.T) {
		if pairs := FindPairs(nil); pairs != nil {
			t.Errorf("expected nil, got %v", pairs)
		}
	})

	t.Run("skips empty optimized paths", func(t *testing.T) {
		pairs := FindPairs([]string{"", ""})
		if len(pairs) != 0 {
			t.Errorf("expected 0 pairs, got %d", len(pairs))
		}
	})

	t.Run("returns pair when original exists", func(t *testing.T) {
		dir := t.TempDir()
		orig := filepath.Join(dir, "clip.mp4")
		opt := filepath.Join(dir, "clip_optimized.mp4")
		os.WriteFile(orig, []byte{}, 0644)

		pairs := FindPairs([]string{opt})
		if len(pairs) != 1 {
			t.Fatalf("expected 1 pair, got %d", len(pairs))
		}
		if pairs[0].Original != orig || pairs[0].Optimized != opt {
			t.Errorf("unexpected pair: %+v", pairs[0])
		}
	})

	t.Run("skips when original missing from disk", func(t *testing.T) {
		dir := t.TempDir()
		opt := filepath.Join(dir, "clip_optimized.mp4")
		// no original created
		pairs := FindPairs([]string{opt})
		if len(pairs) != 0 {
			t.Errorf("expected 0 pairs, got %d", len(pairs))
		}
	})

	t.Run("handles mixed found and not-found", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "a.mp4"), []byte{}, 0644)
		optA := filepath.Join(dir, "a_optimized.mp4")
		optB := filepath.Join(dir, "b_optimized.mp4") // no original

		pairs := FindPairs([]string{optA, optB})
		if len(pairs) != 1 {
			t.Fatalf("expected 1 pair, got %d", len(pairs))
		}
		if pairs[0].Optimized != optA {
			t.Errorf("wrong pair returned: %+v", pairs[0])
		}
	})

	t.Run("returns multiple pairs", func(t *testing.T) {
		dir := t.TempDir()
		for _, name := range []string{"a.mp4", "b.mov", "c.mkv"} {
			os.WriteFile(filepath.Join(dir, name), []byte{}, 0644)
		}
		opts := []string{
			filepath.Join(dir, "a_optimized.mp4"),
			filepath.Join(dir, "b_optimized.mp4"),
			filepath.Join(dir, "c_optimized.mp4"),
		}
		pairs := FindPairs(opts)
		if len(pairs) != 3 {
			t.Errorf("expected 3 pairs, got %d", len(pairs))
		}
	})
}
