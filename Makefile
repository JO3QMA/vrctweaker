# vrchat-tweaker Makefile
# フルビルド、front/backendビルド、lint、fmt、test、e2e を実行

.PHONY: all build build-native build-windows build-front build-back dev-wails lint fmt test test-e2e clean help

# デフォルトターゲット
all: build

# --- ビルド ---

## フルビルド（native + Windows、Wails アプリ全体）
## Linux/WSL から Windows クロスコンパイルには mingw-w64 が必要
build: build-native build-windows

## ネイティブプラットフォームのみビルド
build-native:
	wails build

## Windows 版のみビルド（linux/WSL からは mingw-w64 が必要）
build-windows:
	wails build -platform windows/amd64

## フロントエンドのみビルド
build-front:
	cd frontend && pnpm run build

## バックエンド（Go）のみビルド
build-back:
	go build ./...

## Wails 開発サーバ（DISPLAY 無し環境向け: DevContainer 等では xvfb で仮想 X を用意）
## ブラウザは VSCode のポート転送で http://localhost:34115 を開く
dev-wails:
	xvfb-run -a wails dev

# --- Lint ---

## 全体の Lint（Go + Frontend）
lint: lint-back lint-front

## バックエンド Lint（golangci-lint）
## キャッシュをリポジトリ配下に置き、~/.cache が書けない環境での警告スパムを避ける
lint-back:
	GOLANGCI_LINT_CACHE=$(CURDIR)/.cache/golangci-lint golangci-lint run ./...

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
## 手元で通常の `pnpm run dev` が :5173 を占有していると E2E 用サーバーが起動できないため、失敗時は dev を止めてから再実行すること
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
	@echo "  make build         - フルビルド（native + Windows）"
	@echo "  make build-native  - ネイティブプラットフォームのみビルド"
	@echo "  make build-windows - Windows 版のみビルド"
	@echo "  make build-front   - フロントエンドのみビルド"
	@echo "  make build-back    - バックエンド（Go）のみビルド"
	@echo "  make dev-wails     - Wails dev（xvfb 付き・ヘッドレス向け）"
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
