import { App } from "@/wails/app";
import { isWailsRuntime } from "@/wails/app";
import { parseSafeHttpUrl } from "./safeHttpUrl";

/**
 * Opens a validated http(s) URL in the OS browser (Wails) or a new tab (Storybook/tests).
 * Call only with an href produced by `parseSafeHttpUrl`.
 */
export async function openValidatedExternalHttpUrl(
  href: string,
): Promise<boolean> {
  if (isWailsRuntime()) {
    await App.openExternalURL(href);
    return true;
  }
  if (typeof window !== "undefined") {
    window.open(href, "_blank", "noopener,noreferrer");
    return true;
  }
  return false;
}

/**
 * Parses and validates a raw URL, then opens it in the OS browser (Wails) or a new tab.
 */
export async function openExternalHttpUrl(raw: string): Promise<boolean> {
  const parsed = parseSafeHttpUrl(raw);
  if (!parsed.ok) {
    return false;
  }
  return openValidatedExternalHttpUrl(parsed.href);
}
