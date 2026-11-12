import { defineConfig, devices } from "@playwright/test";

/**
 * 閾値は意図的に小さくしている。Playwright公式Dockerイメージ上のChromiumと
 * Noto CJKで生成した基準画像との比較であり、フォント・アンチエイリアス以外の差分は許さない。
 */
const visualScreenshotExpect = {
  animations: "disabled" as const,
  caret: "hide" as const,
  maxDiffPixelRatio: 0.001,
};

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI ? [["github"], ["html"]] : "html",
  use: {
    baseURL: "http://localhost:5173",
    trace: "on-first-retry",
  },
  projects: [
    {
      name: "chromium",
      testIgnore: /\/visual\//,
      use: { ...devices["Desktop Chrome"] },
    },
    {
      name: "visual",
      testMatch: /\/visual\/.*\.visual\.spec\.ts/,
      retries: 0,
      use: {
        ...devices["Desktop Chrome"],
        locale: "ja-JP",
      },
      expect: {
        toHaveScreenshot: visualScreenshotExpect,
      },
    },
  ],
  webServer: {
    command: "pnpm run dev:e2e",
    url: "http://localhost:5173",
    // Always spawn the e2e-mode dev server below; a hand-started `pnpm run dev` on :5173 prevents bind and breaks tests.
    reuseExistingServer: false,
  },
});
