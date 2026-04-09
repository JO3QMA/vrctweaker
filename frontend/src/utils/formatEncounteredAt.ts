/** ISO 時刻文字列を表示用文字列にする。locale を省略すると実行環境の既定ロケールを使う。 */
export function formatEncounteredAt(iso: string, locale?: string): string {
  try {
    const d = new Date(iso);
    return locale !== undefined ? d.toLocaleString(locale) : d.toLocaleString();
  } catch {
    return iso;
  }
}
