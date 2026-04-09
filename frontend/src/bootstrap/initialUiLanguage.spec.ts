import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  getInitialUILanguageCode,
  FALLBACK_UI_LANGUAGE_CODE,
  GET_UI_LANGUAGE_TIMEOUT_MS,
} from "./initialUiLanguage";

describe("getInitialUILanguageCode", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("returns code when fetch succeeds", async () => {
    const p = getInitialUILanguageCode(() => Promise.resolve("en"));
    await expect(p).resolves.toBe("en");
  });

  it("trims whitespace", async () => {
    const p = getInitialUILanguageCode(() => Promise.resolve("  zh-CN  "));
    await expect(p).resolves.toBe("zh-CN");
  });

  it("falls back when fetch rejects", async () => {
    await expect(
      getInitialUILanguageCode(() => Promise.reject(new Error("ipc"))),
    ).resolves.toBe(FALLBACK_UI_LANGUAGE_CODE);
  });

  it("falls back when fetch resolves empty", async () => {
    await expect(
      getInitialUILanguageCode(() => Promise.resolve("   ")),
    ).resolves.toBe(FALLBACK_UI_LANGUAGE_CODE);
  });

  it("falls back after timeout when fetch never settles", async () => {
    const hanging = () => new Promise<string>(() => undefined);
    const p = getInitialUILanguageCode(hanging);
    await vi.advanceTimersByTimeAsync(GET_UI_LANGUAGE_TIMEOUT_MS);
    await expect(p).resolves.toBe(FALLBACK_UI_LANGUAGE_CODE);
  });
});
