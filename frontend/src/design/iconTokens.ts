/** Icon size scale px values (ADR 0023). */
export const ICON_SIZE_SCALE_PX = [12, 16, 20, 24] as const;

export type IconSizeScalePx = (typeof ICON_SIZE_SCALE_PX)[number];

/** CSS custom property for a numeric icon size step, e.g. `--icon-size-16`. */
export function iconSizeScaleVar<const P extends IconSizeScalePx>(
  px: P,
): `--icon-size-${P}` {
  return `--icon-size-${px}`;
}

export type IconSizePatternName = "compact" | "default" | "emphasis" | "large";

export type IconSizePattern = {
  name: IconSizePatternName;
  varName: `--icon-size-${IconSizePatternName}`;
  px: IconSizeScalePx;
  scaleVar: ReturnType<typeof iconSizeScaleVar>;
};

const ICON_SIZE_PATTERN_SCALE = {
  compact: 12,
  default: 16,
  emphasis: 20,
  large: 24,
} as const satisfies Record<IconSizePatternName, IconSizeScalePx>;

/** Icon size pattern catalog (v1). Each pattern delegates to a scale token. */
export const ICON_SIZE_PATTERNS = (
  Object.entries(ICON_SIZE_PATTERN_SCALE) as [
    IconSizePatternName,
    IconSizeScalePx,
  ][]
).map(([name, px]) => ({
  name,
  varName: `--icon-size-${name}`,
  px,
  scaleVar: iconSizeScaleVar(px),
})) satisfies readonly IconSizePattern[];

export type VtIconSize = IconSizePatternName;

/** VtIcon `size` prop values (1:1 with Icon size pattern names). */
export const VT_ICON_SIZES: readonly VtIconSize[] = ICON_SIZE_PATTERNS.map(
  (pattern) => pattern.name,
);

/** Legacy icon size outside scale (v1 visual policy). */
export const ICON_SIZE_LEGACY = {
  toggle: {
    varName: "--icon-size-legacy-toggle",
    delegatesTo: "14px",
    px: 14,
    adoptionTarget: "--icon-size-default",
  },
} as const;
