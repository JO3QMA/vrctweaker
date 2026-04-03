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
    xvfb \
    xauth \
  && rm -rf /var/lib/apt/lists/*

# Wails は webkit2gtk-4.0 を参照するが、Debian Bookworm+ では 4.1 のみ提供されるため互換用シンボリックリンクを作成
RUN arch=$(dpkg-architecture -qDEB_HOST_MULTIARCH 2>/dev/null || echo "x86_64-linux-gnu") \
  && ln -sf /usr/include/webkitgtk-4.1 /usr/include/webkitgtk-4.0 \
  && ln -sf /usr/lib/${arch}/pkgconfig/webkit2gtk-4.1.pc /usr/lib/${arch}/pkgconfig/webkit2gtk-4.0.pc

# Wails CLI (global install)
RUN go install github.com/wailsapp/wails/v2/cmd/wails@latest

# golangci-lint v2 (must match .golangci.yml version: "2" and CI action)
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

