import { describe, expect, it } from "vitest";
import ja from "./ja.json";

/** Hiragana, katakana, CJK unified ideographs */
const JP = /[\u3040-\u309F\u30A0-\u30FF\u4E00-\u9FFF]/;

/** Half-width space between Latin/digit/} and Japanese */
const LATIN_THEN_JP =
  /[A-Za-z0-9}][ \u3000]+[\u3040-\u309F\u30A0-\u30FF\u4E00-\u9FFF]/;
/**
 * Japanese + space + Latin letter/digit.
 * Does not flag `{` after a label (intentional 「現在 {version}」 / 「最終確認 {when}」 separators).
 */
const JP_THEN_LATIN =
  /[\u3040-\u309F\u30A0-\u30FF\u4E00-\u9FFF][ \u3000]+[A-Za-z0-9]/;

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
    hits.push("Japanese + space + Latin/digit");
  }
  return hits;
}

describe("ja.json Japanese UI spacing", () => {
  it("has no half-width space between Latin/digits/placeholders and Japanese (either order)", () => {
    const violations: { path: string; text: string; reasons: string[] }[] = [];

    for (const { path, text } of collectStringPaths(ja)) {
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

  it("allows intentional label–value space before a placeholder", () => {
    expect(spacingViolations("現在 {version}")).toEqual([]);
    expect(spacingViolations("または Offline")).toContain(
      "Japanese + space + Latin/digit",
    );
    expect(JP.test("あ")).toBe(true);
  });
});
