import { describe, expect, it } from "vitest";
import {
  ELEMENT_PLUS_TYPOGRAPHY_DERIVATIVES,
  ELEMENT_PLUS_TYPOGRAPHY_MAPPING,
  FONT_FAMILY_UI_VAR,
  FONT_SIZE_DERIVATIVES,
  FONT_SIZE_SCALE_PX,
  FONT_WEIGHT_SCALE,
  LINE_HEIGHT_PATTERNS,
  TEXT_STYLES,
  fontSizeScaleVar,
  fontWeightScaleVar,
} from "./typographyTokens";

describe("typographyTokens", () => {
  it("defines the agreed font-size scale", () => {
    expect([...FONT_SIZE_SCALE_PX]).toEqual([10, 12, 14, 16, 18, 20, 24]);
  });

  it("maps scale px to CSS variable names", () => {
    expect(fontSizeScaleVar(14)).toBe("--font-size-14");
  });

  it("defines line-height patterns", () => {
    expect(LINE_HEIGHT_PATTERNS.map((p) => p.value)).toEqual([1.25, 1.5, 1.75]);
  });

  it("defines font-weight scale", () => {
    expect([...FONT_WEIGHT_SCALE]).toEqual([400, 500, 600, 700]);
    expect(fontWeightScaleVar(600)).toBe("--font-weight-600");
  });

  it("defines font-size derivatives outside the scale", () => {
    expect(FONT_SIZE_DERIVATIVES[0]?.varName).toBe("--font-size-h1");
    expect(FONT_SIZE_DERIVATIVES[0]?.cssValue).toBe(
      "calc(var(--font-size-14) * 1.4)",
    );
  });

  it("exposes UI font family token for Storybook catalog", () => {
    expect(FONT_FAMILY_UI_VAR).toBe("--font-family-ui");
  });

  it("maps each text style to catalog tokens", () => {
    expect(TEXT_STYLES.map((s) => s.className)).toEqual([
      "text-h1",
      "text-h2",
      "text-h3",
      "text-h4",
      "text-body",
      "text-body-sm",
      "text-caption",
    ]);
    expect(TEXT_STYLES[0]?.fontSize).toBe("var(--font-size-h1)");
    expect(TEXT_STYLES[0]?.fontWeightVar).toBe(fontWeightScaleVar(600));
  });

  it("maps Element Plus font sizes to app tokens", () => {
    for (const row of ELEMENT_PLUS_TYPOGRAPHY_MAPPING) {
      expect(row.appVar).toMatch(/^--font-size-\d+$/);
    }
    expect(ELEMENT_PLUS_TYPOGRAPHY_DERIVATIVES[0]?.valuePx).toBe(13);
  });
});
