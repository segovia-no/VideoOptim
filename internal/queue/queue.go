package queue

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"VideoOptim/internal/ffmpeg"
	"VideoOptim/internal/settings"

	"github.com/google/uuid"
)

type ProgressEvent struct {
	ID     string
	Update ffmpeg.ProgressUpdate
}

type CompleteEvent struct {
	ID         string
	Result     *ffmpeg.EncodeResult
	SkipReason string
}

type ErrorEvent struct {
	ID  string
	Err error
}

type Queue struct {
	mu            sync.Mutex
	jobs          []*Job
	detector      *ffmpeg.Detector
	getSettings   func() settings.Settings
	onProgress    func(ProgressEvent)
	onComplete    func(CompleteEvent)
	onError       func(ErrorEvent)
	onStart       func(id string)
	cancelCurrent chan struct{}
	running       bool
	startTimer    *time.Timer
}

func New(
	detector *ffmpeg.Detector,
	getSettings func() settings.Settings,
	onProgress func(ProgressEvent),
	onComplete func(CompleteEvent),
	onError func(ErrorEvent),
	onStart func(id string),
) *Queue {
	return &Queue{
		detector:    detector,
		getSettings: getSettings,
		onProgress:  onProgress,
		onComplete:  onComplete,
		onError:     onError,
		onStart:     onStart,
	}
}

func (q *Queue) Add(paths []string) []*Job {
	q.mu.Lock()
	existing := make(map[string]bool)
	for _, j := range q.jobs {
		existing[j.Path] = true
	}
	var added []*Job
	for _, path := range paths {
		if existing[path] {
			continue
		}
		job := &Job{
			ID:       uuid.New().String(),
			Path:     path,
			Filename: filepath.Base(path),
			Status:   StatusWaiting,
			AddedAt:  time.Now(),
		}
		q.jobs = append(q.jobs, job)
		added = append(added, job)
		existing[path] = true
	}
	if len(added) > 0 {
		if q.startTimer != nil {
			q.startTimer.Stop()
		}
		q.startTimer = time.AfterFunc(600*time.Millisecond, func() {
			go q.run()
		})
	}
	q.mu.Unlock()

	return added
}

func (q *Queue) Jobs() []*Job {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]*Job, len(q.jobs))
	copy(out, q.jobs)
	return out
}

func (q *Queue) CancelCurrent() {
	q.mu.Lock()
	if q.cancelCurrent != nil {
		close(q.cancelCurrent)
		q.cancelCurrent = nil
	}
	q.mu.Unlock()
}

func (q *Queue) ClearCompleted() {
	q.mu.Lock()
	defer q.mu.Unlock()
	var remaining []*Job
	for _, j := range q.jobs {
		if j.Status == StatusWaiting || j.Status == StatusProcessing {
			remaining = append(remaining, j)
		}
	}
	q.jobs = remaining
}

func (q *Queue) run() {
	q.mu.Lock()
	if q.running {
		q.mu.Unlock()
		return
	}
	q.running = true
	q.mu.Unlock()

	defer func() {
		q.mu.Lock()
		q.running = false
		q.mu.Unlock()
	}()

	for {
		job := q.nextWaiting()
		if job == nil {
			return
		}
		q.process(job)
	}
}

func (q *Queue) nextWaiting() *Job {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, j := range q.jobs {
		if j.Status == StatusWaiting {
			return j
		}
	}
	return nil
}

func (q *Queue) process(job *Job) {
	cancelCh := make(chan struct{})
	q.mu.Lock()
	job.Status = StatusProcessing
	q.cancelCurrent = cancelCh
	q.mu.Unlock()

	q.onStart(job.ID)

	s := q.getSettings()

	info, err := q.detector.Probe(job.Path)
	if err != nil {
		q.setError(job, err)
		return
	}

	if info.Codec == "hevc" {
		origStat, statErr := os.Stat(job.Path)
		var origSize int64
		if statErr == nil {
			origSize = origStat.Size()
		}
		q.mu.Lock()
		job.Status = StatusSkipped
		job.SkipReason = "hevc"
		job.OriginalSize = origSize
		job.Progress = 100
		q.mu.Unlock()
		q.onComplete(CompleteEvent{
			ID:         job.ID,
			Result:     &ffmpeg.EncodeResult{OriginalSize: origSize},
			SkipReason: "hevc",
		})
		return
	}

	result, err := q.detector.Encode(
		job.Path,
		info,
		s,
		func(u ffmpeg.ProgressUpdate) {
			q.mu.Lock()
			job.Progress = u.Percent
			job.Elapsed = formatDuration(u.Elapsed)
			job.FPS = u.FPS
			q.mu.Unlock()
			q.onProgress(ProgressEvent{ID: job.ID, Update: u})
		},
		cancelCh,
	)

	if err != nil {
		if err.Error() == "cancelled" {
			q.mu.Lock()
			job.Status = StatusCancelled
			q.mu.Unlock()
			return
		}
		q.setError(job, err)
		return
	}

	q.mu.Lock()
	job.OriginalSize = result.OriginalSize
	job.OutputSize = result.OutputSize
	job.OutputPath = result.OutputPath
	if result.OutputPath == "" {
		job.Status = StatusSkipped
		job.Savings = 0
		job.Progress = 100
	} else {
		job.Status = StatusDone
		job.Savings = (1 - float64(result.OutputSize)/float64(result.OriginalSize)) * 100
		job.Progress = 100
	}
	q.mu.Unlock()

	q.onComplete(CompleteEvent{ID: job.ID, Result: result})
}

func (q *Queue) setError(job *Job, err error) {
	q.mu.Lock()
	job.Status = StatusError
	job.Error = err.Error()
	q.mu.Unlock()
	q.onError(ErrorEvent{ID: job.ID, Err: err})
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}
