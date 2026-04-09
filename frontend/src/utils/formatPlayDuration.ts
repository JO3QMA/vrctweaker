export interface PlayDurationUnits {
  hour: string;
  minute: string;
  second: string;
}

const jaUnits: PlayDurationUnits = {
  hour: "時間",
  minute: "分",
  second: "秒",
};

/** プレイ時間を「HH+時+MM+分+SS+秒」形式で返す（数値は2桁ゼロ埋め）。 */
export function formatPlayDurationHMS(
  totalSeconds: number,
  units: PlayDurationUnits = jaUnits,
): string {
  const s = Math.max(0, Math.floor(totalSeconds));
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = s % 60;
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${pad(h)}${units.hour}${pad(m)}${units.minute}${pad(sec)}${units.second}`;
}
