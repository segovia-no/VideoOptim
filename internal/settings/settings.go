package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Encoder string

const (
	EncoderLibx265          Encoder = "libx265"
	EncoderHevcVideoToolbox Encoder = "hevc_videotoolbox"
)

type Settings struct {
	Encoder         Encoder  `json:"encoder"`
	CRF             int      `json:"crf"`
	KeepAudio       bool     `json:"keepAudio"`
	DiscardIfNoGain bool     `json:"discardIfNoGain"`
	AcceptedFormats []string `json:"acceptedFormats"`
	OutputFolder    string   `json:"outputFolder,omitempty"`
}

func Default() Settings {
	return Settings{
		Encoder:         EncoderHevcVideoToolbox,
		CRF:             24,
		KeepAudio:       true,
		DiscardIfNoGain: true,
		AcceptedFormats: []string{"mp4", "mov", "mkv", "avi", "webm"},
	}
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "VideoOptim", "settings.json"), nil
}

func Load() (Settings, error) {
	path, err := configPath()
	if err != nil {
		return Default(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Default(), nil
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return Default(), nil
	}
	return s, nil
}

func Save(s Settings) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
