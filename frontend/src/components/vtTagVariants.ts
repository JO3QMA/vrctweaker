export const VT_TAG_VARIANTS = [
  "success",
  "warning",
  "danger",
  "info",
  "neutral",
  "primary",
] as const;

export type VtTagVariant = (typeof VT_TAG_VARIANTS)[number];

export type VtTagSize = "large" | "default" | "small";

export type VtTagProps = {
  variant: VtTagVariant;
  size?: VtTagSize;
};

/** Maps VtTag variant to Element Plus el-tag `type` (neutral → undefined; EP binding uses info+plain). */
export function vtTagElementType(
  variant: VtTagVariant,
): "success" | "warning" | "danger" | "info" | "primary" | undefined {
  if (variant === "neutral") return undefined;
  return variant;
}
