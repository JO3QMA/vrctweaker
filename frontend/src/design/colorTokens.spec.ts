import { describe, expect, it } from "vitest";
import {
  BRAND_COLOR_TOKENS,
  COLOR_LEGACY_ALIASES,
  hasLegacyAlias,
  NEUTRAL_COLOR_TOKENS,
  PRESENCE_COLOR_TOKENS,
  SEMANTIC_COLOR_TOKENS,
  SERVER_STATUS_COLOR_TOKENS,
  type SimpleColorToken,
} from "./colorTokens";

describe("colorTokens", () => {
  it("defines brand tokens with --color- prefix", () => {
    expect(BRAND_COLOR_TOKENS.map((t) => t.varName)).toEqual([
      "--color-brand",
      "--color-brand-hover",
    ]);
  });

  it("uses name for all simple token rows", () => {
    for (const row of [
      ...BRAND_COLOR_TOKENS,
      ...NEUTRAL_COLOR_TOKENS,
      ...SEMANTIC_COLOR_TOKENS,
      ...SERVER_STATUS_COLOR_TOKENS,
    ]) {
      expect(row.name.length).toBeGreaterThan(0);
      expect(row.varName.startsWith("--color-")).toBe(true);
    }
  });

  it("defines neutral surface and text tokens", () => {
    const names = NEUTRAL_COLOR_TOKENS.map((t) => t.varName);
    expect(names).toContain("--color-bg-base");
    expect(names).toContain("--color-text-muted");
    expect(names).toContain("--color-text-inverse");
    expect(names).toContain("--color-border");
  });

  it("defines semantic catalog danger through info", () => {
    expect(SEMANTIC_COLOR_TOKENS.map((t) => t.name)).toEqual([
      "danger",
      "success",
      "warning",
      "info",
    ]);
  });

  it("defines server status domain names", () => {
    expect(SERVER_STATUS_COLOR_TOKENS.map((t) => t.name)).toEqual([
      "operational",
      "degraded",
      "partial",
      "major",
      "maintenance",
      "unknown",
    ]);
  });

  it("defines presence domain bg and border pairs", () => {
    for (const row of PRESENCE_COLOR_TOKENS) {
      expect(row.name.length).toBeGreaterThan(0);
      expect(row.borderVar).toBe(`${row.bgVar}-border`);
    }
  });

  it("maps legacy aliases to canonical tokens only", () => {
    for (const { legacy, target } of COLOR_LEGACY_ALIASES) {
      expect(legacy.startsWith("--")).toBe(true);
      expect(target.startsWith("--color-")).toBe(true);
    }
  });

  it("derives COLOR_LEGACY_ALIASES from token legacyAlias fields", () => {
    const tokens: SimpleColorToken[] = [
      ...BRAND_COLOR_TOKENS,
      ...NEUTRAL_COLOR_TOKENS,
      ...SEMANTIC_COLOR_TOKENS,
    ];
    const expected = tokens.filter(hasLegacyAlias).map((token) => ({
      legacy: token.legacyAlias,
      target: token.varName,
    }));
    expect(COLOR_LEGACY_ALIASES).toEqual(expected);
  });
});
