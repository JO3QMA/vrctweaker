/** プレイ時間を「HH時間MM分SS秒」形式で返す（時・分・秒は2桁ゼロ埋め）。 */
export function formatPlayDurationHMS(totalSeconds: number): string {
  const s = Math.max(0, Math.floor(totalSeconds));
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = s % 60;
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${pad(h)}時間${pad(m)}分${pad(sec)}秒`;
}
