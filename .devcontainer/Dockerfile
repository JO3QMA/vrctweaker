FROM mcr.microsoft.com/devcontainers/go:1

SHELL ["/bin/bash", "-lc"]

RUN apt-get update \
  && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    gnupg \
    build-essential \
    pkg-config \
    libglib2.0-dev \
    libgtk-3-dev \
    libwebkit2gtk-4.1-dev \
    mingw-w64 \
  && rm -rf /var/lib/apt/lists/*

# Wails CLI (global install)
RUN go install github.com/wailsapp/wails/v2/cmd/wails@latest

# golangci-lint (used in CI)
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

