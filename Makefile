SHELL        := /bin/bash
FFMPEG_VER   := 7.1.1
NCPU         := $(shell sysctl -n hw.logicalcpu 2>/dev/null || echo 4)
X265_PREFIX  := $(shell brew --prefix x265 2>/dev/null)

BUILD_DIR    := $(CURDIR)/build
FFMPEG_BIN   := $(BUILD_DIR)/ffmpeg
APP_BUNDLE   := $(BUILD_DIR)/bin/VideoOptim.app
APP_BIN      := $(APP_BUNDLE)/Contents/MacOS
DMG_PATH     := $(BUILD_DIR)/bin/VideoOptim.dmg

.PHONY: all app dmg ffmpeg clean

all: dmg

# ── Full app build ────────────────────────────────────────────────────────────
app: $(FFMPEG_BIN)/ffmpeg
	~/go/bin/wails build -ldflags="-s -w" -platform darwin/arm64
	rm -f $(APP_BIN)/ffprobe
	install -m755 $(FFMPEG_BIN)/ffmpeg $(APP_BIN)/ffmpeg
	codesign --force --sign - $(APP_BIN)/ffmpeg
	codesign --force --sign - "$(APP_BUNDLE)"
	@echo "==> Bundle: $$(du -sh '$(APP_BUNDLE)' | cut -f1)"

# ── DMG ───────────────────────────────────────────────────────────────────────
dmg: app
	rm -f "$(DMG_PATH)"
	hdiutil create \
	  -volname "VideoOptim" \
	  -srcfolder "$(APP_BUNDLE)" \
	  -ov -format UDZO \
	  "$(DMG_PATH)"
	@echo "==> DMG: $$(du -sh '$(DMG_PATH)' | cut -f1)  →  $(DMG_PATH)"

# ── ffmpeg ───────────────────────────────────────────────────────────────────
ffmpeg: $(FFMPEG_BIN)/ffmpeg

$(FFMPEG_BIN)/ffmpeg:
	@[ -n "$(X265_PREFIX)" ] || { echo "ERROR: x265 not found. Run: brew install x265"; exit 1; }
	@echo "==> Building ffmpeg $(FFMPEG_VER) (x265 from $(X265_PREFIX))..."
	@set -euo pipefail; \
	mkdir -p $(FFMPEG_BIN); \
	TMP=$$(mktemp -d); \
	trap "rm -rf $$TMP" EXIT; \
	curl -fsSL "https://ffmpeg.org/releases/ffmpeg-$(FFMPEG_VER).tar.bz2" \
	  | tar xj -C $$TMP --strip-components=1; \
	cd $$TMP && \
	PKG_CONFIG_PATH="$(X265_PREFIX)/lib/pkgconfig" \
	./configure \
	  --prefix=$(BUILD_DIR)/ffmpeg-prefix \
	  --bindir=$(FFMPEG_BIN) \
	  --disable-everything \
	  --disable-network \
	  --disable-doc \
	  --disable-htmlpages \
	  --disable-manpages \
	  --disable-podpages \
	  --disable-txtpages \
	  --enable-gpl \
	  --enable-version3 \
	  --enable-libx265 \
	  --enable-videotoolbox \
	  --enable-static \
	  --disable-shared \
	  --pkg-config-flags="--static" \
	  --extra-cflags="-I$(X265_PREFIX)/include" \
	  --extra-ldflags="-L$(X265_PREFIX)/lib -lc++ \
	    -framework VideoToolbox \
	    -framework CoreFoundation \
	    -framework CoreMedia \
	    -framework CoreVideo \
	    -framework CoreServices \
	    -framework AudioToolbox" \
	  --enable-protocol=file,pipe \
	  --enable-demuxer=mov,matroska,avi \
	  --enable-muxer=mp4 \
	  --enable-decoder=h264,hevc,vp8,vp9,mpeg4video,prores,av1 \
	  --enable-encoder=hevc_videotoolbox,libx265 \
	  --enable-parser=h264,hevc,vp8,vp9,mpeg4video,aac,mpegaudio \
	  --enable-bsf=hevc_mp4toannexb,h264_mp4toannexb,aac_adtstoasc \
	  --enable-filter=null,anull,format,scale \
	  --enable-avformat \
	  --enable-avcodec \
	  --enable-avutil \
	  --enable-swscale \
	  --enable-swresample && \
	$(MAKE) -j$(NCPU) install
	strip $(FFMPEG_BIN)/ffmpeg
	@echo "==> ffmpeg $$(du -sh $(FFMPEG_BIN)/ffmpeg | cut -f1)"

clean:
	rm -rf $(FFMPEG_BIN) $(BUILD_DIR)/ffmpeg-prefix
