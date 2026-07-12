package queue

import (
	"math"
	"testing"
	"time"
)

const floatEps = 1e-9

func TestCalcSavings(t *testing.T) {
	tests := []struct {
		origSize int64
		outSize  int64
		want     float64
	}{
		{1000, 500, 50.0},
		{1000, 1000, 0.0},
		{1000, 0, 100.0},
		{1000, 750, 25.0},
		{1000, 1200, -20.0}, // output larger than original
		{0, 500, 0.0},       // zero origSize — no division
	}

	for _, tt := range tests {
		got := calcSavings(tt.origSize, tt.outSize)
		if math.Abs(got-tt.want) > floatEps {
			t.Errorf("calcSavings(%d, %d) = %v, want %v", tt.origSize, tt.outSize, got, tt.want)
		}
	}
}

// stopTimer prevents the 600ms debounce from calling run() during tests.
func stopTimer(q *Queue) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.startTimer != nil {
		q.startTimer.Stop()
		q.startTimer = nil
	}
}

func TestQueueAdd(t *testing.T) {
	t.Run("adds single job", func(t *testing.T) {
		q := &Queue{}
		added := q.Add([]string{"/videos/a.mp4"})
		stopTimer(q)
		if len(added) != 1 {
			t.Errorf("expected 1 added, got %d", len(added))
		}
		if len(q.Jobs()) != 1 {
			t.Errorf("expected 1 job in queue, got %d", len(q.Jobs()))
		}
	})

	t.Run("deduplicates same path on second add", func(t *testing.T) {
		q := &Queue{}
		q.Add([]string{"/videos/a.mp4"})
		stopTimer(q)
		added := q.Add([]string{"/videos/a.mp4"})
		stopTimer(q)
		if len(added) != 0 {
			t.Errorf("expected 0 added on duplicate, got %d", len(added))
		}
		if len(q.Jobs()) != 1 {
			t.Errorf("expected 1 job total, got %d", len(q.Jobs()))
		}
	})

	t.Run("adds only new paths from a batch with duplicates", func(t *testing.T) {
		q := &Queue{}
		q.Add([]string{"/videos/a.mp4", "/videos/b.mp4"})
		stopTimer(q)
		// a.mp4 is duplicate, c.mp4 is new
		added := q.Add([]string{"/videos/a.mp4", "/videos/c.mp4"})
		stopTimer(q)
		if len(added) != 1 {
			t.Errorf("expected 1 new job, got %d", len(added))
		}
		if added[0].Path != "/videos/c.mp4" {
			t.Errorf("expected c.mp4, got %q", added[0].Path)
		}
		if len(q.Jobs()) != 3 {
			t.Errorf("expected 3 total jobs, got %d", len(q.Jobs()))
		}
	})

	t.Run("deduplicates within a single batch", func(t *testing.T) {
		q := &Queue{}
		added := q.Add([]string{"/videos/a.mp4", "/videos/a.mp4", "/videos/b.mp4"})
		stopTimer(q)
		if len(added) != 2 {
			t.Errorf("expected 2 added (dedup within batch), got %d", len(added))
		}
	})

	t.Run("returns empty when given empty input", func(t *testing.T) {
		q := &Queue{}
		added := q.Add([]string{})
		if len(added) != 0 {
			t.Errorf("expected 0 added, got %d", len(added))
		}
	})

	t.Run("job has correct initial fields", func(t *testing.T) {
		q := &Queue{}
		added := q.Add([]string{"/videos/clip.mp4"})
		stopTimer(q)
		job := added[0]
		if job.Status != StatusWaiting {
			t.Errorf("Status = %q, want %q", job.Status, StatusWaiting)
		}
		if job.Filename != "clip.mp4" {
			t.Errorf("Filename = %q, want %q", job.Filename, "clip.mp4")
		}
		if job.Path != "/videos/clip.mp4" {
			t.Errorf("Path = %q, want %q", job.Path, "/videos/clip.mp4")
		}
		if job.ID == "" {
			t.Error("ID should not be empty")
		}
	})
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "0:00"},
		{time.Second, "0:01"},
		{59 * time.Second, "0:59"},
		{60 * time.Second, "1:00"},
		{65 * time.Second, "1:05"},
		{90 * time.Second, "1:30"},
		{600 * time.Second, "10:00"},
		{3600 * time.Second, "60:00"},
		{3661 * time.Second, "61:01"},
		// sub-second rounds down
		{400 * time.Millisecond, "0:00"},
		// rounds up at 500ms
		{500 * time.Millisecond, "0:01"},
	}

	for _, tt := range tests {
		got := FormatDuration(tt.d)
		if got != tt.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
