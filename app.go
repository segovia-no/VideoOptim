package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"VideoOptim/internal/cleanup"
	"VideoOptim/internal/ffmpeg"
	"VideoOptim/internal/queue"
	"VideoOptim/internal/settings"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx      context.Context
	detector *ffmpeg.Detector
	q        *queue.Queue
	settings settings.Settings
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	s, _ := settings.Load()
	a.settings = s

	detector, err := ffmpeg.Detect()
	if err != nil {
		wailsRuntime.EventsEmit(ctx, "ffmpeg:missing", map[string]string{
			"message":    err.Error(),
			"installCmd": "brew install ffmpeg",
		})
		return
	}
	a.detector = detector

	a.q = queue.New(
		detector,
		func() settings.Settings { return a.settings },
		func(e queue.ProgressEvent) {
			wailsRuntime.EventsEmit(ctx, "job:progress", map[string]interface{}{
				"id":      e.ID,
				"percent": e.Update.Percent,
				"elapsed": queue.FormatDuration(e.Update.Elapsed),
				"fps":     e.Update.FPS,
			})
		},
		func(e queue.CompleteEvent) {
			wailsRuntime.EventsEmit(ctx, "job:complete", map[string]interface{}{
				"id":           e.ID,
				"outputPath":   e.Result.OutputPath,
				"originalSize": e.Result.OriginalSize,
				"outputSize":   e.Result.OutputSize,
				"skipReason":   e.SkipReason,
			})
		},
		func(e queue.ErrorEvent) {
			wailsRuntime.EventsEmit(ctx, "job:error", map[string]string{
				"id":      e.ID,
				"message": e.Err.Error(),
			})
		},
		func(id string) {
			wailsRuntime.EventsEmit(ctx, "job:start", map[string]string{"id": id})
		},
	)
}

// AddFiles adds video files/directories to the queue and returns the created jobs.
func (a *App) AddFiles(paths []string) []*queue.Job {
	if a.q == nil {
		return nil
	}
	expanded := expandPaths(paths, a.settings.AcceptedFormats)
	if len(expanded) == 0 {
		return nil
	}
	return a.q.Add(expanded)
}

// GetJobs returns the current queue state.
func (a *App) GetJobs() []*queue.Job {
	if a.q == nil {
		return nil
	}
	return a.q.Jobs()
}

// GetSettings returns the current settings.
func (a *App) GetSettings() settings.Settings {
	return a.settings
}

// SaveSettings persists settings to disk.
func (a *App) SaveSettings(s settings.Settings) error {
	a.settings = s
	return settings.Save(s)
}

// Cleanup moves originals to Trash for all completed jobs where the optimized file is smaller.
func (a *App) Cleanup() cleanup.Result {
	if a.q == nil {
		return cleanup.Result{}
	}
	jobs := a.q.Jobs()
	var optimizedPaths []string
	for _, j := range jobs {
		if j.Status == queue.StatusDone && j.OutputPath != "" {
			optimizedPaths = append(optimizedPaths, j.OutputPath)
		}
	}
	pairs := cleanup.FindPairs(optimizedPaths)
	return cleanup.Run(pairs)
}

// ClearCompleted removes finished/error/skipped jobs from the list.
func (a *App) ClearCompleted() {
	if a.q != nil {
		a.q.ClearCompleted()
	}
}

// IsFFmpegAvailable returns whether ffmpeg was found on startup.
func (a *App) IsFFmpegAvailable() bool {
	return a.detector != nil
}

// OpenFilePicker opens a native file dialog and returns selected paths.
func (a *App) OpenFilePicker() []string {
	paths, err := wailsRuntime.OpenMultipleFilesDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Videos",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Videos (mp4, mov, mkv, avi, webm)", Pattern: "*.mp4;*.mov;*.mkv;*.avi;*.webm"},
		},
	})
	if err != nil || len(paths) == 0 {
		return nil
	}
	return paths
}

// OpenFolderPicker opens a native folder dialog and returns the selected path.
func (a *App) OpenFolderPicker() string {
	path, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Folder",
	})
	if err != nil || path == "" {
		return ""
	}
	return path
}

// RevealInFinder reveals the file in macOS Finder.
func (a *App) RevealInFinder(path string) error {
	return exec.Command("open", "-R", path).Run()
}

// OpenFile opens the file with its default application.
func (a *App) OpenFile(path string) error {
	return exec.Command("open", path).Run()
}

// MoveToTrash moves the file at path to the macOS Trash.
func (a *App) MoveToTrash(path string) error {
	script := fmt.Sprintf(`tell application "Finder" to delete POSIX file %q`, path)
	return exec.Command("osascript", "-e", script).Run()
}

func expandPaths(paths []string, formats []string) []string {
	extSet := make(map[string]bool)
	for _, f := range formats {
		extSet["."+strings.ToLower(f)] = true
	}
	var result []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if info.IsDir() {
			_ = filepath.Walk(p, func(path string, fi os.FileInfo, err error) error {
				if err != nil || fi.IsDir() {
					return nil
				}
				if extSet[strings.ToLower(filepath.Ext(path))] {
					result = append(result, path)
				}
				return nil
			})
		} else {
			if extSet[strings.ToLower(filepath.Ext(p))] {
				result = append(result, p)
			}
		}
	}
	return result
}

