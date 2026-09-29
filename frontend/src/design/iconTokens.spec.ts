import { describe, expect, it } from "vitest";
import {
  ICON_SIZE_PATTERNS,
  VT_ICON_SIZES,
  iconSizeScaleVar,
} from "./iconTokens";

describe("iconTokens", () => {
  it("maps scale px to CSS variable names", () => {
    expect(iconSizeScaleVar(16)).toBe("--icon-size-16");
  });

  it("exposes VtIcon size props matching patterns", () => {
    expect([...VT_ICON_SIZES]).toEqual([
      "compact",
      "default",
      "emphasis",
      "large",
    ]);
  });

  it("keeps each pattern name, varName, and px aligned", () => {
    const pxByName = {
      compact: 12,
      default: 16,
      emphasis: 20,
      large: 24,
    } as const;
    for (const pattern of ICON_SIZE_PATTERNS) {
      expect(pattern.px).toBe(pxByName[pattern.name]);
      expect(pattern.varName).toBe(`--icon-size-${pattern.name}`);
      expect(pattern.scaleVar).toBe(`--icon-size-${pattern.px}`);
    }
  });
});
