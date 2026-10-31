import { describe, expect, it } from "vitest";
import ja from "./ja.json";

/** Hiragana, katakana, CJK unified ideographs */
const JP = /[\u3040-\u309F\u30A0-\u30FF\u4E00-\u9FFF]/;
const LATIN_OR_CLOSE = /[A-Za-z0-9}]/;
const LATIN_OR_OPEN = /[A-Za-z0-9{]/;

/** Half-width space between Latin/digit/} and Japanese */
const LATIN_THEN_JP =
  /[A-Za-z0-9}][ \u3000]+[\u3040-\u309F\u30A0-\u30FF\u4E00-\u9FFF]/;
/** Half-width space between Japanese and Latin/digit/{ */
const JP_THEN_LATIN =
  /[\u3040-\u309F\u30A0-\u30FF\u4E00-\u9FFF][ \u3000]+[A-Za-z0-9{]/;

/**
 * yt-dlp video tab copy uses intentional label–value spaces (e.g. 現在 {version}).
 * Allowlisted until a dedicated policy covers placeholder spacing app-wide.
 */
const YTDLP_REPLACE_ALLOWLIST_PREFIXES = [
  "video.experimentalFeatures",
  "video.alwaysWarn",
  "video.replaceLabel",
  "video.statusPrefix",
  "video.effectiveOfficial",
  "video.effectiveBundled",
  "video.cacheVersion",
  "video.cacheMissing",
  "video.latest",
  "video.latestUnchecked",
  "video.checkLatest",
  "video.updateCache",
  "video.versionLineCurrent",
  "video.versionLineGithubLabel",
  "video.versionLineLastCheck",
  "video.githubCheckUnavailable",
  "video.actionFetch",
  "video.actionUpdate",
  "video.actionChecking",
  "video.actionUpdating",
  "video.currentNotDownloaded",
  "video.errorDetailsToggle",
  "video.openCacheFolder",
  "video.openToolsFolder",
  "video.riskAckTitle",
  "video.riskAckBody",
  "video.riskAckConfirm",
  "video.flashEnabled",
  "video.flashDisabled",
  "video.flashLatest",
  "video.flashUpdated",
  "video.resultFirstDownload",
  "video.resultUpdated",
  "video.resultUpToDate",
  "video.errorRiskAckRequired",
  "video.errorUnsupportedPlatform",
  "video.errorMaintenanceFailed",
  "video.notInitialized",
  "video.errors",
  "video.cookieLinkage",
] as const;

function isAllowlisted(path: string): boolean {
  return YTDLP_REPLACE_ALLOWLIST_PREFIXES.some(
    (prefix) => path === prefix || path.startsWith(`${prefix}.`),
  );
}

function collectStringPaths(
  value: unknown,
  prefix = "",
): { path: string; text: string }[] {
  if (typeof value === "string") {
    return prefix ? [{ path: prefix, text: value }] : [];
  }
  if (Array.isArray(value)) {
    return value.flatMap((item, i) =>
      collectStringPaths(item, `${prefix}[${i}]`),
    );
  }
  if (value && typeof value === "object") {
    return Object.entries(value).flatMap(([key, nested]) => {
      const path = prefix ? `${prefix}.${key}` : key;
      return collectStringPaths(nested, path);
    });
  }
  return [];
}

function spacingViolations(text: string): string[] {
  const hits: string[] = [];
  if (LATIN_THEN_JP.test(text)) {
    hits.push("Latin/digit/} + space + Japanese");
  }
  if (JP_THEN_LATIN.test(text)) {
    hits.push("Japanese + space + Latin/digit/{");
  }
  return hits;
}

describe("ja.json Japanese UI spacing", () => {
  it("has no half-width space between Latin/digits/placeholders and Japanese (either order)", () => {
    const violations: { path: string; text: string; reasons: string[] }[] = [];

    for (const { path, text } of collectStringPaths(ja)) {
      if (isAllowlisted(path)) {
        continue;
      }
      const reasons = spacingViolations(text);
      if (reasons.length > 0) {
        violations.push({ path, text, reasons });
      }
    }

    expect(
      violations,
      violations
        .map(
          (v) =>
            `${v.path}: ${v.reasons.join(", ")}\n  ${JSON.stringify(v.text)}`,
        )
        .join("\n\n"),
    ).toEqual([]);
  });

  it("documents allowlist entries for yt-dlp replace (#369)", () => {
    expect(YTDLP_REPLACE_ALLOWLIST_PREFIXES.length).toBeGreaterThan(0);
    expect(LATIN_OR_CLOSE.test("a")).toBe(true);
    expect(JP.test("あ")).toBe(true);
    expect(LATIN_OR_OPEN.test("{")).toBe(true);
  });
});
