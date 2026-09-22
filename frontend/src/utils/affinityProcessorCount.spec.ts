import { describe, expect, it, vi, beforeEach } from "vitest";
import { DEFAULT_LOGICAL_PROCESSOR_COUNT } from "./affinityMask";

vi.mock("../wails/app", () => ({
  App: {
    getLogicalProcessorCount: vi.fn(),
  },
  logicalProcessorCountFallback: vi.fn(() => DEFAULT_LOGICAL_PROCESSOR_COUNT),
}));

import { App, logicalProcessorCountFallback } from "../wails/app";
import {
  normalizeLogicalProcessorCount,
  resolveLogicalProcessorCount,
} from "./affinityProcessorCount";

describe("normalizeLogicalProcessorCount", () => {
  it("accepts positive integers up to 512", () => {
    expect(normalizeLogicalProcessorCount(24)).toBe(24);
    expect(normalizeLogicalProcessorCount(512)).toBe(512);
  });

  it("falls back for invalid values", () => {
    vi.mocked(logicalProcessorCountFallback).mockReturnValue(600);
    expect(normalizeLogicalProcessorCount(0)).toBe(512);
    expect(normalizeLogicalProcessorCount(3.5)).toBe(512);
    expect(normalizeLogicalProcessorCount(Number.NaN)).toBe(512);
    expect(normalizeLogicalProcessorCount(513)).toBe(512);
  });
});

describe("resolveLogicalProcessorCount", () => {
  beforeEach(() => {
    vi.mocked(App.getLogicalProcessorCount).mockReset();
    vi.mocked(logicalProcessorCountFallback).mockReturnValue(
      DEFAULT_LOGICAL_PROCESSOR_COUNT,
    );
  });

  it("returns normalized count from App when valid", async () => {
    vi.mocked(App.getLogicalProcessorCount).mockResolvedValue(24);
    await expect(resolveLogicalProcessorCount()).resolves.toBe(24);
  });

  it("falls back when count is invalid", async () => {
    vi.mocked(App.getLogicalProcessorCount).mockResolvedValue(0);
    vi.mocked(logicalProcessorCountFallback).mockReturnValue(12);
    await expect(resolveLogicalProcessorCount()).resolves.toBe(12);
  });

  it("falls back when App throws", async () => {
    vi.mocked(App.getLogicalProcessorCount).mockRejectedValue(new Error("ipc"));
    vi.mocked(logicalProcessorCountFallback).mockReturnValue(600);
    await expect(resolveLogicalProcessorCount()).resolves.toBe(512);
  });
});
