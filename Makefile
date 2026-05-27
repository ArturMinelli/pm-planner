.PHONY: all cli frontend-build desktop tidy test

CLI_BIN := bin/pm
DESKTOP_BIN := bin/pm-desktop

# On amd64 distro with wrongly installed "linux/386" Go, CGO/WebKit libs on the system are
# 64‑bit → force GOARCH=amd64 when building the desktop binary (Makefile only). For
# native 386 machines, override: DESKTOP_FORCE_GOHOST=1 make desktop
UNAME_M := $(shell uname -m)
GOHOSTARCH ?= $(shell go env GOARCH)
DESKTOP_GOARCH ?= $(GOHOSTARCH)
ifneq ($(DESKTOP_FORCE_GOHOST),1)
ifeq ($(GOHOSTARCH),386)
ifneq ($(filter x86_64 amd64,$(UNAME_M)),)
  DESKTOP_GOARCH := amd64
endif
endif
endif

all: cli desktop

tidy:
	go mod tidy

frontend-build:
	cd desktop/frontend && npm ci && npm run build

# Wails needs CGO + gtk/webkit headers on Linux, and `-tags production` so the real
# webview backend is compiled (otherwise the binary exits with a “build tags” error).
# Ubuntu 24+ ships WebKitGTK 4.1 only; tag `webkit2_41` makes pkg-config use webkit2gtk-4.1.
desktop: frontend-build
	@mkdir -p bin
	cd desktop && CGO_ENABLED=1 GOARCH=$(DESKTOP_GOARCH) GOOS=linux \
		go build -tags production,webkit2_41 -o ../$(DESKTOP_BIN) .

cli:
	@mkdir -p bin
	go build -o $(CLI_BIN) .

test:
	go test ./pkg/... .
