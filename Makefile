# vrchat-tweaker Makefile
# フルビルド、front/backendビルド、lint、fmt、test、e2e を実行

.PHONY: all build build-front build-back lint fmt test test-e2e clean help

# デフォルトターゲット
all: build

# --- ビルド ---

## フルビルド（frontend + backend、Wails アプリ全体）
## wails build が内部で frontend:build を実行する
build:
	wails build

## フロントエンドのみビルド
build-front:
	cd frontend && pnpm run build

## バックエンド（Go）のみビルド
build-back:
	go build ./...

# --- Lint ---

## 全体の Lint（Go + Frontend）
lint: lint-back lint-front

## バックエンド Lint（golangci-lint）
lint-back:
	golangci-lint run ./...

## フロントエンド Lint（ESLint + vue-tsc）
lint-front:
	cd frontend && pnpm run lint
	cd frontend && pnpm exec vue-tsc --noEmit

# --- フォーマット ---

## 全体のフォーマット（Go + Frontend）
fmt: fmt-back fmt-front

## バックエンドフォーマット（go fmt）
fmt-back:
	go fmt ./...

## フロントエンドフォーマット（Prettier）
fmt-front:
	cd frontend && pnpm run format

# --- テスト ---

## 全体のユニットテスト（Go + Frontend）
test: test-back test-front

## バックエンドテスト（go test）
test-back:
	go test -v -race -cover ./internal/...

## フロントエンドテスト（Vitest）
test-front:
	cd frontend && pnpm run test

# --- E2E テスト ---

## E2Eテスト（Playwright）
## 初回は `make test-e2e-install` でブラウザをインストール
test-e2e:
	cd frontend && pnpm run test:e2e

## Playwright ブラウザ（chromium）のインストール
test-e2e-install:
	cd frontend && pnpm exec playwright install --with-deps chromium

# --- ユーティリティ ---

## フロントエンド依存関係のインストール
install-front:
	cd frontend && pnpm install --frozen-lockfile

## クリーン（ビルド成果物の削除）
clean:
	rm -rf frontend/dist build/bin
	go clean -cache

## ヘルプ
help:
	@echo "vrchat-tweaker Makefile"
	@echo ""
	@echo "ビルド:"
	@echo "  make build         - フルビルド（frontend + backend）"
	@echo "  make build-front   - フロントエンドのみビルド"
	@echo "  make build-back    - バックエンド（Go）のみビルド"
	@echo ""
	@echo "品質チェック:"
	@echo "  make lint         - 全体の Lint（golangci-lint + ESLint + vue-tsc）"
	@echo "  make lint-back    - バックエンド Lint"
	@echo "  make lint-front   - フロントエンド Lint"
	@echo ""
	@echo "フォーマット:"
	@echo "  make fmt          - 全体のフォーマット（go fmt + Prettier）"
	@echo "  make fmt-back     - バックエンドフォーマット"
	@echo "  make fmt-front    - フロントエンドフォーマット"
	@echo ""
	@echo "テスト:"
	@echo "  make test         - 全体のユニットテスト"
	@echo "  make test-back    - バックエンドテスト"
	@echo "  make test-front   - フロントエンドテスト"
	@echo "  make test-e2e     - E2Eテスト（Playwright）"
	@echo "  make test-e2e-install - Playwright ブラウザのインストール"
	@echo ""
	@echo "その他:"
	@echo "  make install-front - フロントエンド依存関係のインストール"
	@echo "  make clean        - ビルド成果物の削除"
	@echo "  make help         - このヘルプを表示"
