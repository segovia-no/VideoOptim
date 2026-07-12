package queue

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
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
	currentPid    int
	paused        bool
	stopAll       bool
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

func (q *Queue) MarkOriginalDeleted(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, j := range q.jobs {
		if j.ID == id {
			j.OriginalDeleted = true
			return
		}
	}
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

func (q *Queue) Pause() {
	q.mu.Lock()
	pid := q.currentPid
	q.paused = true
	q.mu.Unlock()
	if pid > 0 {
		if p, err := os.FindProcess(pid); err == nil {
			p.Signal(syscall.SIGSTOP)
		}
	}
}

func (q *Queue) Resume() {
	q.mu.Lock()
	pid := q.currentPid
	q.paused = false
	q.mu.Unlock()
	if pid > 0 {
		if p, err := os.FindProcess(pid); err == nil {
			p.Signal(syscall.SIGCONT)
		}
	}
}

func (q *Queue) Stop() {
	q.mu.Lock()
	q.stopAll = true
	q.paused = false
	ch := q.cancelCurrent
	q.cancelCurrent = nil
	q.mu.Unlock()
	if ch != nil {
		close(ch)
	}
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
		q.mu.Lock()
		if q.stopAll {
			q.stopAll = false
			q.mu.Unlock()
			return
		}
		q.mu.Unlock()
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
	q.beginJob(job, cancelCh)
	q.onStart(job.ID)

	s := q.getSettings()

	info, err := q.detector.Probe(job.Path)
	if err != nil {
		q.setError(job, err)
		return
	}

	if info.Codec == "hevc" {
		q.skipHevc(job)
		return
	}

	result, err := q.runEncode(job, info, s, cancelCh)

	q.mu.Lock()
	q.currentPid = 0
	q.mu.Unlock()

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

	q.applyResult(job, result)
	q.onComplete(CompleteEvent{ID: job.ID, Result: result})
}

func (q *Queue) beginJob(job *Job, cancelCh chan struct{}) {
	q.mu.Lock()
	defer q.mu.Unlock()
	job.Status = StatusProcessing
	q.cancelCurrent = cancelCh
}

func (q *Queue) skipHevc(job *Job) {
	var origSize int64
	if fi, err := os.Stat(job.Path); err == nil {
		origSize = fi.Size()
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
}

func (q *Queue) runEncode(job *Job, info *ffmpeg.VideoInfo, s settings.Settings, cancelCh chan struct{}) (*ffmpeg.EncodeResult, error) {
	return q.detector.Encode(
		job.Path,
		info,
		s,
		func(u ffmpeg.ProgressUpdate) {
			q.mu.Lock()
			job.Progress = u.Percent
			job.Elapsed = FormatDuration(u.Elapsed)
			job.FPS = u.FPS
			q.mu.Unlock()
			q.onProgress(ProgressEvent{ID: job.ID, Update: u})
		},
		func(pid int) {
			q.mu.Lock()
			q.currentPid = pid
			q.mu.Unlock()
		},
		cancelCh,
	)
}

func (q *Queue) applyResult(job *Job, result *ffmpeg.EncodeResult) {
	q.mu.Lock()
	defer q.mu.Unlock()
	job.OriginalSize = result.OriginalSize
	job.OutputSize = result.OutputSize
	job.OutputPath = result.OutputPath
	job.Progress = 100
	if result.OutputPath == "" {
		job.Status = StatusSkipped
		job.Savings = 0
	} else {
		job.Status = StatusDone
		job.Savings = calcSavings(result.OriginalSize, result.OutputSize)
	}
}

func (q *Queue) setError(job *Job, err error) {
	q.mu.Lock()
	job.Status = StatusError
	job.Error = err.Error()
	q.mu.Unlock()
	q.onError(ErrorEvent{ID: job.ID, Err: err})
}

func calcSavings(origSize, outSize int64) float64 {
	if origSize == 0 {
		return 0
	}
	return (1 - float64(outSize)/float64(origSize)) * 100
}

func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}
