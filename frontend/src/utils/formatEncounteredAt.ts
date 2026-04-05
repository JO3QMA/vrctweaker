/** ISO 時刻文字列を日本語ロケールの表示用文字列にする（遭遇ログ等）。 */
export function formatEncounteredAt(iso: string): string {
  try {
    return new Date(iso).toLocaleString("ja-JP");
  } catch {
    return iso;
  }
}
