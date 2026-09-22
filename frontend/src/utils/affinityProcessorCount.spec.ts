import { describe, expect, it, vi, beforeEach } from "vitest";
import { DEFAULT_LOGICAL_PROCESSOR_COUNT } from "./affinityMask";

vi.mock("../wails/app", () => ({
  App: {
    getLogicalProcessorCount: vi.fn(),
  },
}));

import { App } from "../wails/app";
import { resolveLogicalProcessorCount } from "./affinityProcessorCount";

describe("resolveLogicalProcessorCount", () => {
  beforeEach(() => {
    vi.mocked(App.getLogicalProcessorCount).mockReset();
  });

  it("returns count from App when positive", async () => {
    vi.mocked(App.getLogicalProcessorCount).mockResolvedValue(24);
    await expect(resolveLogicalProcessorCount()).resolves.toBe(24);
  });

  it("falls back when count is zero", async () => {
    vi.mocked(App.getLogicalProcessorCount).mockResolvedValue(0);
    await expect(resolveLogicalProcessorCount()).resolves.toBe(
      DEFAULT_LOGICAL_PROCESSOR_COUNT,
    );
  });

  it("falls back when App throws", async () => {
    vi.mocked(App.getLogicalProcessorCount).mockRejectedValue(new Error("ipc"));
    await expect(resolveLogicalProcessorCount()).resolves.toBe(
      DEFAULT_LOGICAL_PROCESSOR_COUNT,
    );
  });
});
