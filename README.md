# VideoOptim

A macOS video compression app inspired by [ImageOptim](https://imageoptim.com). Drop videos in, get smaller videos out.

Uses ffmpeg under the hood with HEVC/H.265 encoding. Supports batch processing via drag-and-drop or folder scanning. Audio is untouched.

---

## Features

- Drag & drop videos or folders onto the window
- HEVC encoding via **libx265** (best compression) or **VideoToolbox** (hardware accelerated)
- Pause, resume, and stop the queue at any time
- Real-time progress bar with elapsed time per file
- No-gain discard — if the output is larger, the original is kept (configurable)
- Cleanup button — moves originals to Trash after reviewing results
- Supports MP4, MOV, MKV, AVI, WebM

## Installation

Download `VideoOptim.dmg` from [Releases](../../releases), open it, drag the app to Applications.

**First launch:** macOS may block the app since it isn't notarized. Right-click → Open to bypass, or run:
```bash
xattr -cr /Applications/VideoOptim.app
```

## Requirements

- macOS 12+, Apple Silicon (M1 or later)
- No runtime dependencies — ffmpeg is bundled inside the app

---

## Usage

1. Launch VideoOptim
2. Drop video files or folders onto the window (or use **Add Files…** / **Add Folder…**)
3. Compression runs automatically, one file at a time
4. Use the **Pause / Resume / Stop** controls while the queue is running
5. When done, click **Clean up originals** to move source files to Trash

Output files are saved alongside the originals with an `_optimized.mp4` suffix:
```
holiday.mov  →  holiday_optimized.mp4
```

---

## Settings

Open via the gear icon or **VideoOptim → Settings** (⌘,).

| Setting | Default | Notes |
|---|---|---|
| Encoder | hevc_videotoolbox | Hardware encoder, fastest |
| — | libx265 | Software encoder, best compression ratio |
| CRF | 24 | Quality level 18–35 (libx265 only) |
| Keep audio | On | Copies audio stream without re-encoding |
| Discard if no gain | On | Keeps original when compressed file is larger |
| Accepted formats | All | Toggle which extensions are picked up |

---

## Development

### Prerequisites

- [Go](https://go.dev) 1.26+
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

The build pipeline compiles a minimal static ffmpeg (with `hevc_videotoolbox` + `libx265`) and bundles it inside the `.app`. No Homebrew runtime dependency for end users.

**Prerequisites:** `brew install x265` (one-time, for the static build)

```bash
# 1. Build bundled ffmpeg (one-time, ~5 min)
make ffmpeg

# 2. Build app + DMG
make dmg
```

Output: `build/bin/VideoOptim.dmg` (~6.5 MB), `build/bin/VideoOptim.app` (~14 MB, self-contained)

### Testing

```bash
# Go tests
go test ./...

# JS tests
cd frontend && npm test
```

---

## Project Structure

```
VideoOptim/
├── main.go                    # Wails entry + window options
├── app.go                     # Go methods exposed to the frontend (IPC)
├── Makefile                   # ffmpeg build + app bundle + DMG
├── internal/
│   ├── ffmpeg/                # Binary detection, video probe, encoder
│   ├── queue/                 # Sequential job runner with pause/resume/stop
│   ├── cleanup/               # Move originals to Trash
│   └── settings/              # User preferences (JSON)
└── frontend/
    └── src/
        ├── App.svelte          # Root layout, toolbar, modals
        ├── stores/queue.js     # Shared job state + derived stats
        ├── utils/
        │   ├── events.js       # Wails event subscriptions
        │   └── format.js       # formatBytes utility
        └── components/
            ├── DragDropZone.svelte  # Drag-and-drop wrapper + overlay
            ├── FileList.svelte      # Scrollable job list with header
            ├── FileRow.svelte       # Per-job row with context menu
            └── Settings.svelte     # Settings modal
```

---

## License

MIT
