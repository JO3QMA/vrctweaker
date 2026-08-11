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

/** Maps VtTag variant to Element Plus el-tag `type` (neutral omits type). */
export function vtTagElementType(
  variant: VtTagVariant,
): "success" | "warning" | "danger" | "info" | "primary" | undefined {
  if (variant === "neutral") return undefined;
  return variant;
}

/** Omit a single attribute before forwarding to Element Plus. */
export function omitVtTagAttr(
  attrs: Record<string, unknown>,
  key: string,
): Record<string, unknown> {
  const copy = { ...attrs };
  delete copy[key];
  return copy;
}
