/**
 * useSessionUnlock – manages the credential unlock lifecycle.
 *
 * Startup flow:
 *   1. `beginStartupUnlock()` is called from App.vue (and awaited from SettingsView) so all
 *      consumers share one in-flight unlock; it wraps `tryUnlockOnStartup()`.
 *   2. It fetches the blob from Go, decrypts it with Web Crypto, and calls UnlockVRChatSession.
 *   3. If the blob is legacy plaintext, it migrates by re-wrapping before unlocking.
 *   4. If decryption fails (IDB key lost), it clears the stored credential and sets
 *      `needsRelogin` so the UI can show a re-login prompt.
 *
 * Login flow (called after App.login() returns OK):
 *   `persistAfterLogin(plaintextToken)` wraps the token and calls PersistWrappedCredential.
 *
 * Logout flow:
 *   `handleLogout()` clears both the stored credential and the IDB wrapping key.
 */

import { ref } from "vue";
import { App } from "@/wails/app";
import {
  wrapToken,
  unwrapBlob,
  deleteWrappingKey,
  WRAPPED_BLOB_MAGIC,
} from "@/services/credentialCrypto";

export type UnlockState =
  | "idle"
  | "unlocking"
  | "unlocked"
  | "needs-relogin"
  | "error";

/** Stable substring in errors from Go UnlockSession when the stored credential is invalid. */
export const UNLOCK_NEEDS_RELOGIN_MARKER = "VRCTWK_UNLOCK_NEEDS_RELOGIN";

/**
 * True when startup unlock failed for auth/session reasons (clear stored blob, prompt re-login).
 * Also matches legacy Go errors without the marker for older backends.
 */
export function unlockFailureRequiresRelogin(e: unknown): boolean {
  const msg = e instanceof Error ? e.message : String(e);
  if (msg.includes(UNLOCK_NEEDS_RELOGIN_MARKER)) return true;
  if (/session expired/i.test(msg)) return true;
  const lower = msg.toLowerCase();
  if (lower === "not authenticated" || lower.startsWith("not authenticated:")) {
    return true;
  }
  return false;
}

// Module-level shared state so all consumers (App.vue, SettingsView.vue, etc.)
// observe the same unlock lifecycle without prop drilling or provide/inject.
const state = ref<UnlockState>("idle");
const errorMessage = ref("");

let startupUnlockPromise: Promise<void> | null = null;

/**
 * Clears module-level unlock state so each Storybook story (or test) starts isolated.
 * Does not touch Go or IndexedDB.
 */
export function resetSessionUnlockForStorybook(): void {
  state.value = "idle";
  errorMessage.value = "";
  startupUnlockPromise = null;
}

export function useSessionUnlock() {
  /**
   * Called at app startup. Fetches the blob from Go and attempts to unlock the session.
   * Sets `state` to `"unlocked"`, `"needs-relogin"`, or `"error"` depending on outcome.
   */
  async function tryUnlockOnStartup(): Promise<void> {
    state.value = "unlocking";
    errorMessage.value = "";

    const hasBlob = await App.hasStoredCredential();
    if (!hasBlob) {
      state.value = "needs-relogin";
      return;
    }

    const blob = await App.getCredentialBlob();
    if (!blob) {
      state.value = "needs-relogin";
      return;
    }

    let token: string;
    try {
      token = await unwrapBlob(blob);
    } catch {
      // IDB key loss or blob corruption – clear the unusable blob and request re-login.
      await App.clearStoredCredential().catch(() => undefined);
      state.value = "needs-relogin";
      errorMessage.value =
        "保存された認証情報を復号できませんでした。再ログインが必要です。";
      return;
    }

    // Migration: if the stored value was a legacy plaintext token, re-wrap it.
    if (!blob.startsWith(WRAPPED_BLOB_MAGIC)) {
      try {
        const wrapped = await wrapToken(token);
        await App.persistWrappedCredential(wrapped);
      } catch {
        // Migration failure is non-fatal; the session still unlocks this time.
      }
    }

    // Auth/session failure: clear blob so we do not loop on a dead token. Transient errors
    // (network, timeout): keep blob and surface "error" so a later retry can succeed.
    try {
      await App.unlockVRChatSession(token);
      state.value = "unlocked";
    } catch (e: unknown) {
      errorMessage.value = e instanceof Error ? e.message : String(e);
      if (unlockFailureRequiresRelogin(e)) {
        await App.clearStoredCredential().catch(() => undefined);
        state.value = "needs-relogin";
      } else {
        state.value = "error";
      }
      return;
    }
  }

  /**
   * Starts startup unlock once per app lifetime; further calls return the same promise.
   * Use from App.vue to kick off unlock and from SettingsView to wait before reading IsLoggedIn.
   */
  function beginStartupUnlock(): Promise<void> {
    if (!startupUnlockPromise) {
      startupUnlockPromise = tryUnlockOnStartup();
    }
    return startupUnlockPromise;
  }

  /**
   * Called immediately after a successful login with the plaintext token from LoginResultDTO.
   * Wraps the token and persists it so future startups can restore the session.
   */
  async function persistAfterLogin(plaintextToken: string): Promise<void> {
    if (!plaintextToken) return;
    try {
      const wrapped = await wrapToken(plaintextToken);
      await App.persistWrappedCredential(wrapped);
      state.value = "unlocked";
    } catch {
      // Persistence failure is logged but non-fatal; the session works until app restart.
    }
  }

  /**
   * Called when the user logs out. Clears the credential store and the IDB wrapping key.
   * After App.logout(), Go already deletes the credential; clearStoredCredential here is
   * redundant but idempotent and keeps this path correct when handleLogout is used alone.
   */
  async function handleLogout(): Promise<void> {
    await App.clearStoredCredential().catch(() => undefined);
    await deleteWrappingKey().catch(() => undefined);
    state.value = "needs-relogin";
  }

  return {
    state,
    errorMessage,
    tryUnlockOnStartup,
    beginStartupUnlock,
    persistAfterLogin,
    handleLogout,
  };
}
