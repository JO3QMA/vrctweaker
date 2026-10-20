/** Matches Go wailsapp.errSettingsNotInitialized (backend not ready yet). */
export const SETTINGS_NOT_INITIALIZED_MARKER =
  "settings repository is not initialized";

function errorMessage(e: unknown): string {
  if (e instanceof Error) {
    return e.message;
  }
  if (typeof e === "string") {
    return e;
  }
  if (e && typeof e === "object" && "message" in e) {
    return String((e as { message: unknown }).message);
  }
  return "";
}

export function isSettingsNotInitializedError(e: unknown): boolean {
  return errorMessage(e).includes(SETTINGS_NOT_INITIALIZED_MARKER);
}
