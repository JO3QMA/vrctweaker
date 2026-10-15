import { describe, expect, it } from "vitest";
import {
  isSettingsNotInitializedError,
  SETTINGS_NOT_INITIALIZED_MARKER,
} from "../firstLaunchGuideErrors";

describe("isSettingsNotInitializedError", () => {
  it("detects the Go sentinel", () => {
    expect(
      isSettingsNotInitializedError(new Error(SETTINGS_NOT_INITIALIZED_MARKER)),
    ).toBe(true);
  });

  it("rejects unrelated errors", () => {
    expect(isSettingsNotInitializedError(new Error("network down"))).toBe(false);
  });
});
