# VideoOptim

A macOS video compression app inspired by [ImageOptim](https://imageoptim.com). Drop videos in, get smaller videos out.

Uses ffmpeg under the hood with HEVC/H.265 encoding. Supports batch processing via drag-and-drop or folder scanning.

---

## Features

- Drag & drop videos or folders
- HEVC encoding via **libx265** (best ratio) or **VideoToolbox** (hardware accelerated)
- Real-time progress bar with elapsed time
- No-gain discard — if the output is larger, the original is kept (configurable)
- Cleanup button — moves originals to Trash after reviewing results
- Supports MP4, MOV, MKV, AVI, WebM
- Dark and light mode (follows system)

## Requirements

- macOS 12+
- [ffmpeg](https://ffmpeg.org) installed via Homebrew

```bash
brew install ffmpeg
```

---

## Usage

1. Launch VideoOptim
2. Drop video files or folders onto the window (or click **Choose files…**)
3. Compression runs automatically, one file at a time
4. When done, click **Clean up originals** to move source files to Trash

Output files are saved alongside the originals with an `_optimized.mp4` suffix:
```
holiday.mov  →  holiday_optimized.mp4
```

---

## Settings

Open via the gear icon (bottom-right corner).

| Setting | Default | Notes |
|---|---|---|
| Encoder | hevc_videotoolbox | Hardware encoder, much faster (default on macOS) |
| — | libx265 | Software encoder, best compression ratio |
| CRF | 24 | Quality level 18–35 (libx265 only) |
| Keep audio | On | Copies audio stream without re-encoding |
| Discard if no gain | On | Keep original when compressed file is larger |
| Accepted formats | All | Toggle which extensions are picked up |

---

## Development

### Prerequisites

- [Go](https://go.dev) 1.21+
- [Wails CLI](https://wails.io) v2
- Node.js 22+ (via [nvm](https://github.com/nvm-sh/nvm))

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest
export PATH=$PATH:$(go env GOPATH)/bin

# Install frontend dependencies
cd frontend && nvm use && npm install && cd ..
```

### Run in dev mode

```bash
wails dev
```

Hot-reloads the frontend on save. Go changes require a restart.

### Production build

```bash
wails build
```

Output: `build/bin/VideoOptim.app`

---

## Project Structure

```
VideoOptim/
├── main.go                  # Wails app entry + window options
├── app.go                   # Go methods exposed to the frontend
├── internal/
│   ├── ffmpeg/              # Binary detection, video probe, encoder
│   ├── queue/               # Sequential job runner
│   ├── cleanup/             # Move originals to Trash
│   └── settings/            # User preferences (JSON)
├── frontend/
│   └── src/
│       ├── App.svelte        # Root layout, event wiring
│       ├── stores/queue.js   # Shared job state
│       └── components/       # FileList, FileRow, Settings
└── docs/                    # Architecture and internals
```

See [`docs/architecture.md`](docs/architecture.md) for a full breakdown.

---

## License

MIT
