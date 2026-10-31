/** Compare yt-dlp product versions (YYYY.MM.DD). Returns true when latest is newer than current. */
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
