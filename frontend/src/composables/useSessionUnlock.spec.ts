import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/wails/app", () => ({
  App: {
    hasStoredCredential: vi.fn().mockResolvedValue(false),
    getCredentialBlob: vi.fn().mockResolvedValue(""),
    unlockVRChatSession: vi.fn().mockResolvedValue(undefined),
    clearStoredCredential: vi.fn().mockResolvedValue(undefined),
    persistWrappedCredential: vi.fn().mockResolvedValue(undefined),
  },
}));

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
