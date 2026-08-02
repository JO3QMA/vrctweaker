import { describe, expect, it } from "vitest";
import {
  SPACING_PATTERNS,
  SPACING_SCALE_PX,
  spacingScaleVar,
} from "./spacingTokens";

describe("spacingTokens", () => {
  it("defines the agreed spacing scale", () => {
    expect([...SPACING_SCALE_PX]).toEqual([4, 8, 12, 16, 24, 32, 48, 64]);
  });

  it("maps scale px to CSS variable names", () => {
    expect(spacingScaleVar(16)).toBe("--space-16");
  });

  it("delegates each pattern to the matching scale token", () => {
    for (const pattern of SPACING_PATTERNS) {
      expect(pattern.scaleVar).toBe(`--space-${pattern.px}`);
    }
  });
});
