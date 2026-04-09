/** プレイ時間をロケール付きの H:M:S 表記で返す（時・分・秒は2桁ゼロ埋め）。 */
export type PlayDurationTranslate = (
  key: string,
  values?: Record<string, unknown>,
) => string;

export function formatPlayDurationHMS(
  totalSeconds: number,
  t: PlayDurationTranslate,
): string {
  const s = Math.max(0, Math.floor(totalSeconds));
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = s % 60;
  const pad = (n: number) => String(n).padStart(2, "0");
  const join = t("common.playDurationJoin");
  return [
    t("common.hours", { n: pad(h) }),
    t("common.minutes", { n: pad(m) }),
    t("common.seconds", { n: pad(sec) }),
  ].join(join);
}
