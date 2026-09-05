/** 遭遇ログの入室・退室列幅（`yyyy/MM/DD HH:mm:ss` + 余白）。 */
export const ENCOUNTER_LOG_TIME_COL_WIDTH = "22ch";

/** ISO 時刻文字列をロケール付き表示用文字列にする（ギャラリー・動画等）。 */
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

/** 遭遇ログ用: 端末ローカル時刻を `yyyy/MM/DD HH:mm:ss` で返す。不正・空は `null`。 */
export function formatEncounterLogTimestamp(iso: string): string | null {
  const trimmed = iso.trim();
  if (!trimmed) return null;
  const d = new Date(trimmed);
  if (Number.isNaN(d.getTime())) return null;
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}/${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}
