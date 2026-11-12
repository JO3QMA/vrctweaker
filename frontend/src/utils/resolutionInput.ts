/** Normalizes GUI resolution dimensions to integers (el-input-number may emit floats). */
export function normalizeResolutionDimension(
  value: number | null | undefined,
): number {
  if (value === null || value === undefined || Number.isNaN(value)) {
    return 0;
  }
  return Math.trunc(value);
}
