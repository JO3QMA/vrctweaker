/** Canvas buffer size fields mutated by syncCanvasBuffer. */
export type CanvasBuffer = {
  width: number;
  height: number;
};

/**
 * Clamp tip center X so a translateX(-50%) tip of tipWidth stays inside the chart.
 */
export function clampCenteredTipX(
  x: number,
  cssW: number,
  tipWidth: number,
  padLeft: number,
  padRight: number,
): number {
  const half = tipWidth / 2;
  const min = padLeft + half;
  const max = cssW - padRight - half;
  if (max < min) return cssW / 2;
  return Math.max(min, Math.min(max, x));
}

/**
 * Resize canvas backing store only when CSS size × DPR actually changes.
 * Assigning width/height always clears the bitmap, even to the same values.
 */
export function syncCanvasBuffer(
  canvas: CanvasBuffer,
  cssW: number,
  cssH: number,
  dpr: number,
): { resized: boolean; width: number; height: number } {
  const width = Math.floor(cssW * dpr);
  const height = Math.floor(cssH * dpr);
  const resized = canvas.width !== width || canvas.height !== height;
  if (resized) {
    canvas.width = width;
    canvas.height = height;
  }
  return { resized, width, height };
}
