import { describe, expect, it } from "vitest";
import { spacingScaleVar } from "./spacingTokens";

describe("spacingTokens", () => {
  it("maps scale px to CSS variable names", () => {
    expect(spacingScaleVar(16)).toBe("--space-16");
  });
});
