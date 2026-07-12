package queue

import "time"

type JobStatus string

const (
	StatusWaiting    JobStatus = "waiting"
	StatusProcessing JobStatus = "processing"
	StatusDone       JobStatus = "done"
	StatusSkipped    JobStatus = "skipped"
	StatusError      JobStatus = "error"
	StatusCancelled  JobStatus = "cancelled"
)

type Job struct {
	ID           string    `json:"id"`
	Path         string    `json:"path"`
	Filename     string    `json:"filename"`
	Status       JobStatus `json:"status"`
	Progress     float64   `json:"progress"`
	Elapsed      string    `json:"elapsed"`
	FPS          float64   `json:"fps"`
	OriginalSize int64     `json:"originalSize"`
	OutputSize   int64     `json:"outputSize"`
	Savings      float64   `json:"savings"`
	Error        string    `json:"error,omitempty"`
	OutputPath   string    `json:"outputPath,omitempty"`
	SkipReason   string    `json:"skipReason,omitempty"`
	AddedAt      time.Time `json:"addedAt"`
}
