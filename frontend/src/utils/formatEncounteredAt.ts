/**
 * 遭遇ログの入室・退室列幅（px）。
 * `el-table-column` の `width` は数値 px のみ有効（`22ch` は `22` px と解釈される）。
 * `yyyy/MM/DD HH:mm:ss`（19 文字）+ セル余白向け。動画履歴の時刻列（140）に合わせる。
 */
export const ENCOUNTER_LOG_TIME_COL_WIDTH = 140;

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
