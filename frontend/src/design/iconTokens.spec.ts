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

  it("documents legacy toggle size for v1 visual policy", () => {
    expect(ICON_SIZE_LEGACY.toggle.px).toBe(14);
    expect(ICON_SIZE_LEGACY.toggle.adoptionTarget).toBe("--icon-size-default");
  });
});
