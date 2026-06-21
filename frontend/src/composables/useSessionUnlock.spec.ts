import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  UNLOCK_NEEDS_RELOGIN_MARKER,
  unlockFailureRequiresRelogin,
} from "./useSessionUnlock";

const cryptoMocks = vi.hoisted(() => ({
  wrapToken: vi.fn(),
  unwrapBlob: vi.fn(),
  deleteWrappingKey: vi.fn(),
}));

vi.mock("@/services/credentialCrypto", () => ({
  WRAPPED_BLOB_MAGIC: "VRCTWKV1:",
  wrapToken: cryptoMocks.wrapToken,
  unwrapBlob: cryptoMocks.unwrapBlob,
  deleteWrappingKey: cryptoMocks.deleteWrappingKey,
}));

vi.mock("@/wails/app", () => ({
  App: {
    hasStoredCredential: vi.fn().mockResolvedValue(false),
    getCredentialBlob: vi.fn().mockResolvedValue(""),
    unlockVRChatSession: vi.fn().mockResolvedValue(undefined),
    clearStoredCredential: vi.fn().mockResolvedValue(undefined),
    persistWrappedCredential: vi.fn().mockResolvedValue(undefined),
  },
}));

async function loadUnlock() {
  vi.resetModules();
  const mod = await import("./useSessionUnlock");
  mod.resetSessionUnlockForStorybook();
  return mod;
}

async function appMocks() {
  const { App } = await import("@/wails/app");
  return {
    hasStoredCredential: vi.mocked(App.hasStoredCredential),
    getCredentialBlob: vi.mocked(App.getCredentialBlob),
    unlockVRChatSession: vi.mocked(App.unlockVRChatSession),
    clearStoredCredential: vi.mocked(App.clearStoredCredential),
    persistWrappedCredential: vi.mocked(App.persistWrappedCredential),
  };
}

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
  beforeEach(() => {
    vi.clearAllMocks();
    cryptoMocks.wrapToken.mockReset();
    cryptoMocks.unwrapBlob.mockReset();
    cryptoMocks.deleteWrappingKey.mockReset();
  });

  it("returns the same promise when called twice", async () => {
    const { useSessionUnlock } = await loadUnlock();
    const app = await appMocks();
    app.hasStoredCredential.mockResolvedValue(false);
    const u = useSessionUnlock();
    const p1 = u.beginStartupUnlock();
    const p2 = u.beginStartupUnlock();
    expect(p1).toBe(p2);
    await p1;
    expect(app.hasStoredCredential).toHaveBeenCalledTimes(1);
  });

  it("sets needs-relogin when stored blob is empty", async () => {
    const { useSessionUnlock } = await loadUnlock();
    const app = await appMocks();
    app.hasStoredCredential.mockResolvedValue(true);
    app.getCredentialBlob.mockResolvedValue("");
    const u = useSessionUnlock();
    await u.beginStartupUnlock();
    expect(u.state.value).toBe("needs-relogin");
  });

  it("unlocks session when wrapped blob decrypts successfully", async () => {
    const { useSessionUnlock } = await loadUnlock();
    const app = await appMocks();
    app.hasStoredCredential.mockResolvedValue(true);
    app.getCredentialBlob.mockResolvedValue("VRCTWKV1:wrapped");
    cryptoMocks.unwrapBlob.mockResolvedValue("plain-token");
    app.unlockVRChatSession.mockResolvedValue(undefined);
    const u = useSessionUnlock();
    await u.beginStartupUnlock();
    expect(u.state.value).toBe("unlocked");
    expect(app.unlockVRChatSession).toHaveBeenCalledWith("plain-token");
  });

  it("clears credential and sets needs-relogin when unwrap fails", async () => {
    const { useSessionUnlock } = await loadUnlock();
    const app = await appMocks();
    app.hasStoredCredential.mockResolvedValue(true);
    app.getCredentialBlob.mockResolvedValue("VRCTWKV1:bad");
    cryptoMocks.unwrapBlob.mockRejectedValue(new Error("decrypt failed"));
    const u = useSessionUnlock();
    await u.beginStartupUnlock();
    expect(u.state.value).toBe("needs-relogin");
    expect(app.clearStoredCredential).toHaveBeenCalled();
    expect(u.errorMessage.value).toContain("再ログイン");
  });

  it("migrates legacy plaintext blob before unlock", async () => {
    const { useSessionUnlock } = await loadUnlock();
    const app = await appMocks();
    app.hasStoredCredential.mockResolvedValue(true);
    app.getCredentialBlob.mockResolvedValue("legacy-plain-token");
    cryptoMocks.unwrapBlob.mockResolvedValue("legacy-plain-token");
    cryptoMocks.wrapToken.mockResolvedValue("VRCTWKV1:new-wrap");
    app.unlockVRChatSession.mockResolvedValue(undefined);
    const u = useSessionUnlock();
    await u.beginStartupUnlock();
    expect(cryptoMocks.wrapToken).toHaveBeenCalledWith("legacy-plain-token");
    expect(app.persistWrappedCredential).toHaveBeenCalledWith(
      "VRCTWKV1:new-wrap",
    );
    expect(u.state.value).toBe("unlocked");
  });

  it("clears credential on auth failure during unlock", async () => {
    const { useSessionUnlock } = await loadUnlock();
    const app = await appMocks();
    app.hasStoredCredential.mockResolvedValue(true);
    app.getCredentialBlob.mockResolvedValue("VRCTWKV1:wrapped");
    cryptoMocks.unwrapBlob.mockResolvedValue("token");
    app.unlockVRChatSession.mockRejectedValue(
      new Error(`wrapped: ${UNLOCK_NEEDS_RELOGIN_MARKER}`),
    );
    const u = useSessionUnlock();
    await u.beginStartupUnlock();
    expect(u.state.value).toBe("needs-relogin");
    expect(app.clearStoredCredential).toHaveBeenCalled();
  });

  it("keeps credential and sets error on transient unlock failure", async () => {
    const { useSessionUnlock } = await loadUnlock();
    const app = await appMocks();
    app.hasStoredCredential.mockResolvedValue(true);
    app.getCredentialBlob.mockResolvedValue("VRCTWKV1:wrapped");
    cryptoMocks.unwrapBlob.mockResolvedValue("token");
    app.unlockVRChatSession.mockRejectedValue(
      new Error("dial tcp: connection refused"),
    );
    const u = useSessionUnlock();
    await u.beginStartupUnlock();
    expect(u.state.value).toBe("error");
    expect(app.clearStoredCredential).not.toHaveBeenCalled();
    expect(u.errorMessage.value).toContain("connection refused");
  });
});

describe("useSessionUnlock persistAfterLogin", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    cryptoMocks.wrapToken.mockReset();
  });

  it("wraps token and persists credential on success", async () => {
    const { useSessionUnlock } = await loadUnlock();
    const app = await appMocks();
    cryptoMocks.wrapToken.mockResolvedValue("VRCTWKV1:stored");
    const u = useSessionUnlock();
    await u.persistAfterLogin("fresh-token");
    expect(cryptoMocks.wrapToken).toHaveBeenCalledWith("fresh-token");
    expect(app.persistWrappedCredential).toHaveBeenCalledWith(
      "VRCTWKV1:stored",
    );
    expect(u.state.value).toBe("unlocked");
  });

  it("ignores empty token", async () => {
    const { useSessionUnlock } = await loadUnlock();
    const app = await appMocks();
    const u = useSessionUnlock();
    await u.persistAfterLogin("");
    expect(cryptoMocks.wrapToken).not.toHaveBeenCalled();
    expect(app.persistWrappedCredential).not.toHaveBeenCalled();
  });

  it("does not throw when wrap fails", async () => {
    const { useSessionUnlock } = await loadUnlock();
    cryptoMocks.wrapToken.mockRejectedValue(new Error("idb unavailable"));
    const u = useSessionUnlock();
    await expect(u.persistAfterLogin("token")).resolves.toBeUndefined();
  });
});

describe("useSessionUnlock tryUnlockOnStartup migration", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    cryptoMocks.wrapToken.mockReset();
    cryptoMocks.unwrapBlob.mockReset();
  });

  it("still unlocks when legacy migration persist fails", async () => {
    const { useSessionUnlock } = await loadUnlock();
    const app = await appMocks();
    app.hasStoredCredential.mockResolvedValue(true);
    app.getCredentialBlob.mockResolvedValue("legacy-plain-token");
    cryptoMocks.unwrapBlob.mockResolvedValue("legacy-plain-token");
    cryptoMocks.wrapToken.mockRejectedValue(new Error("wrap failed"));
    app.unlockVRChatSession.mockResolvedValue(undefined);
    const u = useSessionUnlock();
    await u.tryUnlockOnStartup();
    expect(u.state.value).toBe("unlocked");
    expect(app.unlockVRChatSession).toHaveBeenCalledWith("legacy-plain-token");
  });
});

describe("useSessionUnlock handleLogout", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    cryptoMocks.deleteWrappingKey.mockReset();
  });

  it("clears stored credential and wrapping key", async () => {
    const { useSessionUnlock } = await loadUnlock();
    const app = await appMocks();
    cryptoMocks.deleteWrappingKey.mockResolvedValue(undefined);
    const u = useSessionUnlock();
    u.state.value = "unlocked";
    u.errorMessage.value = "old";
    await u.handleLogout();
    expect(app.clearStoredCredential).toHaveBeenCalled();
    expect(cryptoMocks.deleteWrappingKey).toHaveBeenCalled();
    expect(u.state.value).toBe("needs-relogin");
    expect(u.errorMessage.value).toBe("");
  });
});

describe("resetSessionUnlockForStorybook", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("clears shared state so beginStartupUnlock can run again", async () => {
    const { useSessionUnlock, resetSessionUnlockForStorybook } =
      await loadUnlock();
    const app = await appMocks();
    app.hasStoredCredential.mockResolvedValue(false);
    const u = useSessionUnlock();
    await u.beginStartupUnlock();
    expect(u.state.value).toBe("needs-relogin");
    resetSessionUnlockForStorybook();
    expect(u.state.value).toBe("idle");
    await u.beginStartupUnlock();
    expect(app.hasStoredCredential).toHaveBeenCalledTimes(2);
  });
});
