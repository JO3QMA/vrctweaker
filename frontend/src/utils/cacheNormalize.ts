/**
 * キャッシュ関連の数値を保存・blur 時に min 以上の有限値へ揃える。
 */
export function clampCacheNumeric(n: number | undefined, min: number): number {
  if (typeof n !== "number" || !Number.isFinite(n)) {
    return min;
  }
  return Math.max(min, n);
}
