/**
 * バックエンド（Go）で使用している主な OSS のライセンス情報。
 * go.mod の直接依存関係に基づいています。
 */
export const goLicenses: Array<{
  path: string;
  license: string;
  repository?: string;
}> = [
  {
    path: "github.com/gen2brain/beeep",
    license: "Apache-2.0",
    repository: "https://github.com/gen2brain/beeep",
  },
  {
    path: "github.com/google/uuid",
    license: "BSD-3-Clause",
    repository: "https://github.com/google/uuid",
  },
  {
    path: "github.com/rwcarlsen/goexif",
    license: "MIT",
    repository: "https://github.com/rwcarlsen/goexif",
  },
  {
    path: "github.com/wailsapp/wails/v2",
    license: "MIT",
    repository: "https://github.com/wailsapp/wails",
  },
  {
    path: "github.com/zalando/go-keyring",
    license: "MIT",
    repository: "https://github.com/zalando/go-keyring",
  },
  {
    path: "modernc.org/sqlite",
    license: "BSD-3-Clause",
    repository: "https://gitlab.com/cznic/sqlite",
  },
];
