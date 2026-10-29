/** Minimum gap between adjacent visible X-axis date labels (px). */
export const PLAY_TIME_X_LABEL_MIN_GAP = 4;

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
 * stride until visible labels do not overlap. Always keeps first and last; if the last
 * label would collide with an interior neighbor, drops that neighbor instead.
 */
export function playTimeChartVisibleXLabelIndices(
  input: XLabelLayoutInput,
): number[] {
  const n = input.pointCount;
  if (n <= 0) return [];
  if (n === 1) return [0];

  const minGap = input.minGap ?? PLAY_TIME_X_LABEL_MIN_GAP;
  const plotW = input.plotWidth;
  if (plotW <= 0) return [0, n - 1];

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

  const pruneForEndCollision = (indices: number[]): number[] => {
    const kept = [...indices].sort((a, b) => a - b);
    let changed = true;
    while (changed) {
      changed = false;
      for (let k = 0; k < kept.length - 1; k++) {
        const i = kept[k]!;
        const j = kept[k + 1]!;
        if (!overlaps(i, j)) continue;
        if (j === n - 1 && i !== 0) {
          kept.splice(k, 1);
          changed = true;
          break;
        }
        if (i === 0 && j !== n - 1) {
          kept.splice(k + 1, 1);
          changed = true;
          break;
        }
        if (j === n - 1) {
          kept.splice(k, 1);
          changed = true;
          break;
        }
        if (i === 0) {
          kept.splice(k + 1, 1);
          changed = true;
          break;
        }
        kept.splice(k, 1);
        changed = true;
        break;
      }
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
    const pruned = pruneForEndCollision(candidateForStep(step));
    if (layoutOk(pruned)) {
      return pruned;
    }
  }
  return pruneForEndCollision([0, n - 1]);
}

/** Half-width padding so the widest label stays inside the canvas when centered on an edge point. */
export function playTimeChartXLabelHorizontalPad(
  maxLabelWidth: number,
  edgeMargin: number = PLAY_TIME_X_LABEL_EDGE_MARGIN,
): number {
  return maxLabelWidth / 2 + edgeMargin;
}
