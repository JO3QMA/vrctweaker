import type { VtIconSize } from "../design/iconTokens";

export type { VtIconSize };

export type VtIconProps = {
  size: VtIconSize;
  /** When true (default), sets aria-hidden for decorative icons. */
  decorative?: boolean;
};
