export const VT_TAG_VARIANTS = [
  "success",
  "warning",
  "danger",
  "info",
  "neutral",
  "primary",
] as const;

export type VtTagVariant = (typeof VT_TAG_VARIANTS)[number];

export type VtTagProps = {
  variant: VtTagVariant;
};

/** Maps VtTag variant to Element Plus el-tag `type` (neutral omits type). */
export function vtTagElementType(
  variant: VtTagVariant,
): "success" | "warning" | "danger" | "info" | "primary" | undefined {
  if (variant === "neutral") return undefined;
  return variant;
}
