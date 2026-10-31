/**
 * Compare yt-dlp release tags (typically `YYYY.MM.DD`, optional `v` prefix).
 *
 * Comparison uses the first three dot-separated numeric segments only.
 * Extra suffix segments (e.g. `2026.02.02.1`) and non-numeric tail parts are
 * ignored for ordering — `2026.02.02.1` compares equal to `2026.02.02`.
 * Non-numeric segments parse as 0 (`2026.a.01` → year 2026, month 0, day 1).
 */
export function isYtdlpVersionOlder(current: string, latest: string): boolean {
  const norm = (v: string) => v.trim().replace(/^v/i, "");
  const a = norm(current);
  const b = norm(latest);
  if (!b) return false;
  if (!a) return true;
  const pa = a.split(".").map((n) => parseInt(n, 10) || 0);
  const pb = b.split(".").map((n) => parseInt(n, 10) || 0);
  for (let i = 0; i < 3; i++) {
    const x = pa[i] ?? 0;
    const y = pb[i] ?? 0;
    if (x < y) return true;
    if (x > y) return false;
  }
  return false;
}
