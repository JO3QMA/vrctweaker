/**
 * useSessionUnlock – manages the credential unlock lifecycle.
 *
 * Startup flow:
 *   1. `tryUnlockOnStartup()` is called from App.vue once the Wails runtime is ready.
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

// Module-level shared state so all consumers (App.vue, SettingsView.vue, etc.)
// observe the same unlock lifecycle without prop drilling or provide/inject.
const state = ref<UnlockState>("idle");
const errorMessage = ref("");

export function useSessionUnlock() {
  /**
   * Called at app startup. Fetches the blob from Go and attempts to unlock the session.
   * Sets `state` to `"unlocked"`, `"needs-relogin"`, or `"error"` depending on outcome.
   */
  async function tryUnlockOnStartup(): Promise<void> {
    state.value = "unlocking";
    errorMessage.value = "";

    try {
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

      await App.unlockVRChatSession(token);
      state.value = "unlocked";
    } catch (e) {
      state.value = "error";
      errorMessage.value = e instanceof Error ? e.message : String(e);
    }
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
    persistAfterLogin,
    handleLogout,
  };
}
