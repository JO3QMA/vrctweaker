/** Minimum gap between adjacent visible X-axis date labels (px). */
export const PLAY_TIME_X_LABEL_MIN_GAP = 8;

/** Extra inset from canvas edge so a centered label is not clipped (px). */
export const PLAY_TIME_X_LABEL_EDGE_MARGIN = 4;

export type XLabelLayoutInput = {
  pointCount: number;
  plotLeft: number;
  plotWidth: number;
  /** Measured width per series index (same order as labels). */
  labelWidths: number[];
  minGap?: number;
};

/**
 * Choose which day indices get an X-axis label. Prefers showing every day; increases
 * stride until visible labels do not overlap. The newest day (last index) is always
 * preferred on the X axis; when it collides with an interior neighbor, that neighbor
 * is dropped. If the first and last days would overlap (extremely narrow plot), only
 * the last day's label is kept.
 */
export function playTimeChartVisibleXLabelIndices(
  input: XLabelLayoutInput,
): number[] {
  const n = input.pointCount;
  if (n <= 0) return [];
  if (n === 1) return [0];

  const minGap = input.minGap ?? PLAY_TIME_X_LABEL_MIN_GAP;
  const plotW = input.plotWidth;
  if (plotW <= 0) return [n - 1];

  const pos = (i: number) => input.plotLeft + (i / (n - 1)) * input.plotWidth;

  const overlaps = (a: number, b: number): boolean => {
    const wa = input.labelWidths[a] ?? 0;
    const wb = input.labelWidths[b] ?? 0;
    const need = (wa + wb) / 2 + minGap;
    return Math.abs(pos(a) - pos(b)) < need;
  };

  const candidateForStep = (step: number): number[] => {
    const set = new Set<number>([0, n - 1]);
    for (let i = step; i < n - 1; i += step) {
      set.add(i);
    }
    return [...set].sort((a, b) => a - b);
  };

  /** When the last label collides with its left neighbor, drop that neighbor (including index 0). */
  const resolveLastNeighborCollision = (indices: number[]): number[] => {
    const kept = [...indices].sort((a, b) => a - b);
    const lastPos = kept.indexOf(n - 1);
    if (lastPos <= 0) return kept;
    const prev = kept[lastPos - 1]!;
    if (overlaps(prev, n - 1)) {
      kept.splice(lastPos - 1, 1);
    }
    return kept;
  };

  const layoutOk = (indices: number[]): boolean => {
    for (let k = 0; k < indices.length - 1; k++) {
      if (overlaps(indices[k]!, indices[k + 1]!)) {
        return false;
      }
    }
    return true;
  };

  for (let step = 1; step <= n - 1; step++) {
    const indices = resolveLastNeighborCollision(candidateForStep(step));
    if (layoutOk(indices)) {
      return indices;
    }
  }
  return resolveLastNeighborCollision([0, n - 1]);
}

/** Half-width padding so the widest label stays inside the canvas when centered on an edge point. */
export function playTimeChartXLabelHorizontalPad(
  maxLabelWidth: number,
  edgeMargin: number = PLAY_TIME_X_LABEL_EDGE_MARGIN,
): number {
  return maxLabelWidth / 2 + edgeMargin;
}
