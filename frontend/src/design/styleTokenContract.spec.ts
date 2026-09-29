import { readFileSync } from "node:fs";
import path from "node:path";
import { describe, expect, it } from "vitest";
import {
  BRAND_COLOR_TOKENS,
  COLOR_LEGACY_ALIASES,
  NEUTRAL_COLOR_TOKENS,
  PRESENCE_COLOR_TOKENS,
  SEMANTIC_COLOR_TOKENS,
  SERVER_STATUS_COLOR_TOKENS,
} from "./colorTokens";
import {
  ICON_SIZE_LEGACY,
  ICON_SIZE_PATTERNS,
  ICON_SIZE_SCALE_PX,
} from "./iconTokens";
import {
  parseCssPxLength,
  parseStyleCssRootDeclarations,
  resolveCssVarInMap,
  rootNumericScalePx,
} from "./parseStyleCssRoot";
import { SPACING_PATTERNS, SPACING_SCALE_PX } from "./spacingTokens";
import {
  FONT_SIZE_DERIVATIVES,
  FONT_SIZE_SCALE_PX,
  FONT_WEIGHT_SCALE,
  LINE_HEIGHT_PATTERNS,
} from "./typographyTokens";

const styleCssPath = path.join(process.cwd(), "src/assets/style.css");

function loadRootVars() {
  const css = readFileSync(styleCssPath, "utf8");
  return parseStyleCssRootDeclarations(css);
}

describe("style.css ↔ design token catalogs", () => {
  const root = loadRootVars();

  it("declares every catalog color token on :root", () => {
    const simple = [
      ...BRAND_COLOR_TOKENS,
      ...NEUTRAL_COLOR_TOKENS,
      ...SEMANTIC_COLOR_TOKENS,
      ...SERVER_STATUS_COLOR_TOKENS,
    ];
    for (const row of simple) {
      expect(root[row.varName], row.varName).toBeDefined();
    }
    for (const row of PRESENCE_COLOR_TOKENS) {
      expect(root[row.bgVar], row.bgVar).toBeDefined();
      expect(root[row.borderVar], row.borderVar).toBeDefined();
    }
  });

  it("maps legacy aliases to canonical vars in style.css", () => {
    for (const { legacy, target } of COLOR_LEGACY_ALIASES) {
      expect(root[legacy]).toBe(`var(${target})`);
    }
  });

  it("lists the same spacing scale in TS catalogs and style.css", () => {
    expect(rootNumericScalePx(root, "--space-")).toEqual([...SPACING_SCALE_PX]);
    for (const px of SPACING_SCALE_PX) {
      const varName = `--space-${px}` as const;
      const resolved = resolveCssVarInMap(root, varName);
      expect(parseCssPxLength(resolved)).toBe(px);
    }
  });

  it("matches spacing patterns to scale tokens in style.css", () => {
    for (const pattern of SPACING_PATTERNS) {
      expect(root[pattern.varName]).toBe(`var(${pattern.scaleVar})`);
      const resolved = resolveCssVarInMap(root, pattern.varName);
      expect(parseCssPxLength(resolved)).toBe(pattern.px);
    }
  });

  it("lists the same font-size scale in TS catalogs and style.css", () => {
    expect(rootNumericScalePx(root, "--font-size-")).toEqual([
      ...FONT_SIZE_SCALE_PX,
    ]);
    for (const px of FONT_SIZE_SCALE_PX) {
      const varName = `--font-size-${px}` as const;
      expect(parseCssPxLength(root[varName] ?? "")).toBe(px);
    }
  });

  it("matches font-size derivatives declared in style.css", () => {
    for (const row of FONT_SIZE_DERIVATIVES) {
      expect(root[row.varName]).toBe(row.cssValue);
    }
  });

  it("matches line height patterns in style.css", () => {
    for (const row of LINE_HEIGHT_PATTERNS) {
      expect(Number(root[row.varName])).toBe(row.value);
    }
  });

  it("matches font weight scale in style.css", () => {
    for (const weight of FONT_WEIGHT_SCALE) {
      const varName = `--font-weight-${weight}` as const;
      expect(Number(root[varName])).toBe(weight);
    }
  });

  it("lists the same icon size scale in TS catalogs and style.css", () => {
    expect(rootNumericScalePx(root, "--icon-size-")).toEqual([
      ...ICON_SIZE_SCALE_PX,
    ]);
    for (const px of ICON_SIZE_SCALE_PX) {
      const varName = `--icon-size-${px}` as const;
      expect(parseCssPxLength(root[varName] ?? "")).toBe(px);
    }
    for (const pattern of ICON_SIZE_PATTERNS) {
      expect(root[pattern.varName]).toBe(`var(${pattern.scaleVar})`);
      const resolved = resolveCssVarInMap(root, pattern.varName);
      expect(parseCssPxLength(resolved)).toBe(pattern.px);
    }
  });

  it("matches legacy icon size token in style.css (literal px, not a var() delegate)", () => {
    const legacy = ICON_SIZE_LEGACY.toggle;
    expect(root[legacy.varName]).toBe(legacy.delegatesTo);
    expect(parseCssPxLength(root[legacy.varName] ?? "")).toBe(legacy.px);
  });
});
