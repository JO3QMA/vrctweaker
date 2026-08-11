export const VT_ALERT_VARIANTS = [
  "success",
  "warning",
  "danger",
  "info",
] as const;

export type VtAlertVariant = (typeof VT_ALERT_VARIANTS)[number];

export type VtAlertProps = {
  variant: VtAlertVariant;
};

/** Maps semantic danger to Element Plus alert type `error`. */
export function vtAlertElementType(
  variant: VtAlertVariant,
): "success" | "warning" | "error" | "info" {
  if (variant === "danger") return "error";
  return variant;
}
