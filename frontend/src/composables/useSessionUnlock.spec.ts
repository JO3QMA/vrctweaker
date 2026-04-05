import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  UNLOCK_NEEDS_RELOGIN_MARKER,
  unlockFailureRequiresRelogin,
} from "./useSessionUnlock";

vi.mock("@/wails/app", () => ({
  App: {
    hasStoredCredential: vi.fn().mockResolvedValue(false),
    getCredentialBlob: vi.fn().mockResolvedValue(""),
    unlockVRChatSession: vi.fn().mockResolvedValue(undefined),
    clearStoredCredential: vi.fn().mockResolvedValue(undefined),
    persistWrappedCredential: vi.fn().mockResolvedValue(undefined),
  },
}));

describe("unlockFailureRequiresRelogin", () => {
  it.each([
    [`wrapped: ${UNLOCK_NEEDS_RELOGIN_MARKER}: session expired`, true],
    ["session expired: GET /auth/user", true],
    ["Session Expired: foo", true],
    ["not authenticated", true],
    ["not authenticated: extra", true],
    ["dial tcp: connection refused", false],
    ["context deadline exceeded", false],
    ["", false],
  ])("%s → %s", (msg, want) => {
    expect(unlockFailureRequiresRelogin(new Error(msg))).toBe(want);
  });

  it("handles non-Error throws", () => {
    expect(unlockFailureRequiresRelogin("session expired")).toBe(true);
    expect(unlockFailureRequiresRelogin(123)).toBe(false);
  });
});

describe("useSessionUnlock beginStartupUnlock", () => {
  beforeEach(async () => {
    vi.resetModules();
    const { App } = await import("@/wails/app");
    vi.mocked(App.hasStoredCredential).mockResolvedValue(false);
  });

  it("returns the same promise when called twice", async () => {
    const { useSessionUnlock } = await import("./useSessionUnlock");
    const u = useSessionUnlock();
    const p1 = u.beginStartupUnlock();
    const p2 = u.beginStartupUnlock();
    expect(p1).toBe(p2);
    await p1;
    const { App } = await import("@/wails/app");
    expect(vi.mocked(App.hasStoredCredential)).toHaveBeenCalledTimes(1);
  });
});
