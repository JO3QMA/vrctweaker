import type { VtIconSize } from "../design/iconTokens";

export type { VtIconSize };

export type VtIconProps = {
  size: VtIconSize;
  /** When true, sets aria-hidden for decorative icons (default in VtIcon.vue). */
  decorative?: boolean;
};
