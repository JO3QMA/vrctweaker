import { App } from "@/wails/app";
import { isWailsRuntime } from "@/wails/app";
import { parseSafeHttpUrl } from "./safeHttpUrl";

/**
 * Opens a validated http(s) URL in the OS browser (Wails) or a new tab (Storybook/tests).
 */
export async function openExternalHttpUrl(raw: string): Promise<boolean> {
  const parsed = parseSafeHttpUrl(raw);
  if (!parsed.ok) {
    return false;
  }
  if (isWailsRuntime()) {
    await App.openExternalURL(parsed.href);
    return true;
  }
  if (typeof window !== "undefined") {
    window.open(parsed.href, "_blank", "noopener,noreferrer");
    return true;
  }
  return false;
}
