/** Icon size scale px values (ADR 0023). */
export const ICON_SIZE_SCALE_PX = [12, 16, 20, 24] as const;

export type IconSizeScalePx = (typeof ICON_SIZE_SCALE_PX)[number];

/** CSS custom property for a numeric icon size step, e.g. `--icon-size-16`. */
export function iconSizeScaleVar(
  px: IconSizeScalePx,
): `--icon-size-${IconSizeScalePx}` {
  return `--icon-size-${px}`;
}

export type IconSizePatternName = "compact" | "default" | "emphasis" | "large";

export type IconSizePattern = {
  name: IconSizePatternName;
  varName: `--icon-size-${IconSizePatternName}`;
  px: IconSizeScalePx;
  scaleVar: ReturnType<typeof iconSizeScaleVar>;
};

/** Icon size pattern catalog (v1). Each pattern delegates to a scale token. */
export const ICON_SIZE_PATTERNS = [
  {
    name: "compact",
    varName: "--icon-size-compact",
    px: 12,
    scaleVar: iconSizeScaleVar(12),
  },
  {
    name: "default",
    varName: "--icon-size-default",
    px: 16,
    scaleVar: iconSizeScaleVar(16),
  },
  {
    name: "emphasis",
    varName: "--icon-size-emphasis",
    px: 20,
    scaleVar: iconSizeScaleVar(20),
  },
  {
    name: "large",
    varName: "--icon-size-large",
    px: 24,
    scaleVar: iconSizeScaleVar(24),
  },
] as const satisfies readonly IconSizePattern[];

export type VtIconSize = IconSizePatternName;

/** VtIcon `size` prop values (1:1 with Icon size pattern names). */
export const VT_ICON_SIZES: readonly VtIconSize[] = ICON_SIZE_PATTERNS.map(
  (pattern) => pattern.name,
);

/** Legacy icon size outside scale (v1 visual policy). */
export const ICON_SIZE_LEGACY = {
  toggle: {
    varName: "--icon-size-legacy-toggle",
    delegatesTo: "--font-size-14",
    px: 14,
    adoptionTarget: "--icon-size-default",
  },
} as const;
