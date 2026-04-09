/** ISO 時刻文字列をロケール付き表示用文字列にする（遭遇ログ等）。 */
export function formatEncounteredAt(
  iso: string,
  calendarLocale = "ja-JP",
): string {
  try {
    return new Date(iso).toLocaleString(calendarLocale);
  } catch {
    return iso;
  }
}
