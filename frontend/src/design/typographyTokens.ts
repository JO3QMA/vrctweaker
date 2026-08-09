/** Font size scale px values (ADR 0020). */
export const FONT_SIZE_SCALE_PX = [10, 12, 14, 16, 18, 20, 24] as const;

export type FontSizeScalePx = (typeof FONT_SIZE_SCALE_PX)[number];

/** CSS custom property for a numeric font-size step, e.g. `--font-size-14`. */
export function fontSizeScaleVar(
  px: FontSizeScalePx,
): `--font-size-${FontSizeScalePx}` {
  return `--font-size-${px}`;
}

export type LineHeightPatternName = "tight" | "normal" | "relaxed";

export type LineHeightPattern = {
  name: LineHeightPatternName;
  varName: `--line-height-${LineHeightPatternName}`;
  value: number;
};

/** Line height pattern catalog (v1). Unitless ratios. */
export const LINE_HEIGHT_PATTERNS = [
  { name: "tight", varName: "--line-height-tight", value: 1.25 },
  { name: "normal", varName: "--line-height-normal", value: 1.5 },
  { name: "relaxed", varName: "--line-height-relaxed", value: 1.75 },
] as const satisfies readonly LineHeightPattern[];

/** Font weight scale values (ADR 0020). */
export const FONT_WEIGHT_SCALE = [400, 500, 600, 700] as const;

export type FontWeightScale = (typeof FONT_WEIGHT_SCALE)[number];

/** CSS custom property for a font-weight step, e.g. `--font-weight-600`. */
export function fontWeightScaleVar(
  weight: FontWeightScale,
): `--font-weight-${FontWeightScale}` {
  return `--font-weight-${weight}`;
}

export const FONT_FAMILY_UI_VAR = "--font-family-ui" as const;

export type FontSizeDerivative = {
  name: string;
  varName: `--font-size-${string}`;
  cssValue: string;
  note: string;
};

/**
 * Font sizes outside Typography scale (ADR 0020).
 * Preserves legacy computed sizes without adding scale steps.
 */
export const FONT_SIZE_DERIVATIVES = [
  {
    name: "h1",
    varName: "--font-size-h1",
    cssValue: "calc(var(--font-size-14) * 1.4)",
    note: "Legacy page title (19.6px at 14px root); outside scale by design",
  },
] as const satisfies readonly FontSizeDerivative[];

export type TextStyleName =
  "h1" | "h2" | "h3" | "h4" | "body" | "body-sm" | "caption";

export type TextStyle = {
  name: TextStyleName;
  className: `text-${TextStyleName}`;
  fontSize: string;
  lineHeightVar: LineHeightPattern["varName"];
  fontWeightVar: ReturnType<typeof fontWeightScaleVar>;
};

/** Text style catalog (v1). Colors are not included (see Color tokens). */
export const TEXT_STYLES = [
  {
    name: "h1",
    className: "text-h1",
    fontSize: "var(--font-size-h1)",
    lineHeightVar: "--line-height-tight",
    fontWeightVar: fontWeightScaleVar(600),
  },
  {
    name: "h2",
    className: "text-h2",
    fontSize: "var(--font-size-18)",
    lineHeightVar: "--line-height-tight",
    fontWeightVar: fontWeightScaleVar(600),
  },
  {
    name: "h3",
    className: "text-h3",
    fontSize: "var(--font-size-16)",
    lineHeightVar: "--line-height-tight",
    fontWeightVar: fontWeightScaleVar(600),
  },
  {
    name: "h4",
    className: "text-h4",
    fontSize: "var(--font-size-14)",
    lineHeightVar: "--line-height-normal",
    fontWeightVar: fontWeightScaleVar(600),
  },
  {
    name: "body",
    className: "text-body",
    fontSize: "var(--font-size-14)",
    lineHeightVar: "--line-height-normal",
    fontWeightVar: fontWeightScaleVar(400),
  },
  {
    name: "body-sm",
    className: "text-body-sm",
    fontSize: "var(--font-size-12)",
    lineHeightVar: "--line-height-normal",
    fontWeightVar: fontWeightScaleVar(400),
  },
  {
    name: "caption",
    className: "text-caption",
    fontSize: "var(--font-size-10)",
    lineHeightVar: "--line-height-normal",
    fontWeightVar: fontWeightScaleVar(400),
  },
] as const satisfies readonly TextStyle[];

export type ElementPlusTypographyDerivative = {
  elVar: `--el-font-size-${string}`;
  valuePx: number;
};

/** EP font sizes outside Typography scale (mapping block only). */
export const ELEMENT_PLUS_TYPOGRAPHY_DERIVATIVES = [
  { elVar: "--el-font-size-small", valuePx: 13 },
] as const satisfies readonly ElementPlusTypographyDerivative[];

export type ElementPlusTypographyMapping = {
  elVar: `--el-font-size-${string}`;
  appVar: `--font-size-${FontSizeScalePx}`;
};

/** Element Plus typography mapping (ADR 0020). */
export const ELEMENT_PLUS_TYPOGRAPHY_MAPPING = [
  { elVar: "--el-font-size-extra-large", appVar: fontSizeScaleVar(20) },
  { elVar: "--el-font-size-large", appVar: fontSizeScaleVar(18) },
  { elVar: "--el-font-size-medium", appVar: fontSizeScaleVar(16) },
  { elVar: "--el-font-size-base", appVar: fontSizeScaleVar(14) },
  { elVar: "--el-font-size-extra-small", appVar: fontSizeScaleVar(12) },
] as const satisfies readonly ElementPlusTypographyMapping[];
