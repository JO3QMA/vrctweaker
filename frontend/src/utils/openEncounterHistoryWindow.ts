import type { Router } from "vue-router";
import { isWailsRuntime } from "../wails/app";

export type EncounterHistoryKind = "user" | "world";

const POPUP_FEATURES = "popup=yes,width=560,height=720";

/**
 * Opens encounter history. In Wails, `window.open` loads a separate WebView that
 * cannot reach wails.localhost (connection refused), so we always navigate in-app.
 * Outside Wails, tries a popup and falls back to router.push if blocked.
 */
export function openEncounterHistoryWindow(
  router: Router,
  kind: EncounterHistoryKind,
  id: string,
): void {
  const query: Record<string, string> = { kind };
  if (kind === "user") {
    query.vrcUserId = id;
  } else {
    query.worldId = id;
  }
  if (isWailsRuntime()) {
    void router.push({ name: "encounter-history", query });
    return;
  }
  const base = window.location.href.split("#")[0];
  const qs = new URLSearchParams(query).toString();
  const url = `${base}#/activity/encounter-history?${qs}`;
  const win = window.open(url, "_blank", POPUP_FEATURES);
  if (!win) {
    void router.push({ name: "encounter-history", query });
  }
}
