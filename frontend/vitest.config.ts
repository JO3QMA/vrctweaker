import { defineConfig } from "vitest/config";
import vue from "@vitejs/plugin-vue";
import path from "node:path";

export default defineConfig({
  plugins: [vue()],
  test: {
    environment: "jsdom",
    globals: true,
    include: ["src/**/*.{test,spec}.{js,ts,vue}"],
    setupFiles: ["src/test/setupVitest.ts"],
    coverage: {
      provider: "v8",
      reporter: ["text", "html"],
      all: false,
      exclude: [
        "node_modules/**",
        "src/test/**",
        // Vitest already omits test files from coverage; listed for clarity in reviews
        "src/**/*.{spec,test}.{js,ts,vue}",
        "**/*.stories.ts",
        "**/*.config.*",
        "e2e/**",
        "wailsjs/**",
        ".storybook/**",
      ],
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
});
