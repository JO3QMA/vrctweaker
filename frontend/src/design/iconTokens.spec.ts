import { describe, expect, it } from "vitest";
import {
  ICON_SIZE_LEGACY,
  ICON_SIZE_PATTERNS,
  ICON_SIZE_SCALE_PX,
  VT_ICON_SIZES,
  iconSizeScaleVar,
} from "./iconTokens";

describe("iconTokens", () => {
  it("defines the agreed icon size scale", () => {
    expect([...ICON_SIZE_SCALE_PX]).toEqual([12, 16, 20, 24]);
  });

  it("maps scale px to CSS variable names", () => {
    expect(iconSizeScaleVar(16)).toBe("--icon-size-16");
  });

  it("delegates each pattern to the matching scale token", () => {
    for (const pattern of ICON_SIZE_PATTERNS) {
      expect(pattern.scaleVar).toBe(`--icon-size-${pattern.px}`);
    }
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

  it("documents legacy toggle size for v1 visual policy", () => {
    expect(ICON_SIZE_LEGACY.toggle.px).toBe(14);
    expect(ICON_SIZE_LEGACY.toggle.adoptionTarget).toBe("--icon-size-default");
  });
});
