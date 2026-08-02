export const VT_BUTTON_VARIANTS = [
  "primary",
  "secondary",
  "tertiary",
  "danger",
] as const;

export type VtButtonVariant = (typeof VT_BUTTON_VARIANTS)[number];

export type VtButtonProps =
  | { variant: Exclude<VtButtonVariant, "danger"> }
  | {
      variant: "danger";
      /** Outline when a Primary button is in the same action group. */
      plain?: boolean;
    };
