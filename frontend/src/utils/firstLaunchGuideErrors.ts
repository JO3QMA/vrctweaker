/** Matches Go wailsapp.errSettingsNotInitialized (backend not ready yet). */
export const SETTINGS_NOT_INITIALIZED_MARKER =
  "settings repository is not initialized";

export function isSettingsNotInitializedError(e: unknown): boolean {
  const msg =
    e instanceof Error
      ? e.message
      : typeof e === "string"
        ? e
        : e && typeof e === "object" && "message" in e
          ? String((e as { message: unknown }).message)
          : "";
  return msg.includes(SETTINGS_NOT_INITIALIZED_MARKER);
}
