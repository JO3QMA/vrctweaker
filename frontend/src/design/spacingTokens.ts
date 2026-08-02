/** Spacing scale px values (ADR 0018). */
export const SPACING_SCALE_PX = [4, 8, 12, 16, 24, 32, 48, 64] as const;

export type SpacingScalePx = (typeof SPACING_SCALE_PX)[number];

/** CSS custom property for a numeric spacing step, e.g. `--space-16`. */
export function spacingScaleVar(
  px: SpacingScalePx,
): `--space-${SpacingScalePx}` {
  return `--space-${px}`;
}

export type SpacingPatternName =
  "inline-tight" | "action-group" | "form-field" | "block" | "section" | "page";

export type SpacingPattern = {
  name: SpacingPatternName;
  varName: `--space-${SpacingPatternName}`;
  px: SpacingScalePx;
  scaleVar: ReturnType<typeof spacingScaleVar>;
};

/** Spacing pattern catalog (v1). Each pattern delegates to a scale token. */
export const SPACING_PATTERNS: readonly SpacingPattern[] = [
  {
    name: "inline-tight",
    varName: "--space-inline-tight",
    px: 4,
    scaleVar: "--space-4",
  },
  {
    name: "action-group",
    varName: "--space-action-group",
    px: 8,
    scaleVar: "--space-8",
  },
  {
    name: "form-field",
    varName: "--space-form-field",
    px: 12,
    scaleVar: "--space-12",
  },
  {
    name: "block",
    varName: "--space-block",
    px: 16,
    scaleVar: "--space-16",
  },
  {
    name: "section",
    varName: "--space-section",
    px: 24,
    scaleVar: "--space-24",
  },
  {
    name: "page",
    varName: "--space-page",
    px: 32,
    scaleVar: "--space-32",
  },
] as const;
