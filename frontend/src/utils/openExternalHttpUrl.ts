import { App } from "@/wails/app";
import { isWailsRuntime } from "@/wails/app";
import { parseSafeHttpUrl, type SafeHttpHref } from "./safeHttpUrl";

/**
 * Opens an href produced by `parseSafeHttpUrl` in the OS browser (Wails) or a new tab
 * (Storybook/tests).
 */
export async function openValidatedExternalHttpUrl(
  href: SafeHttpHref,
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
