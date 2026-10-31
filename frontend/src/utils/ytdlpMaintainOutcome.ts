/**
 * yt-dlp maintain action outcomes from CheckAndUpdateOfficialYTDLPCache.
 * Keep in sync with internal/usecase/ytdlp_outcome.go.
 */
export const YTDLP_MAINTAIN_OUTCOME = {
  firstDownload: "firstDownload",
  updated: "updated",
  unchanged: "unchanged",
  error: "error",
} as const;

export type YtdlpMaintainOutcome =
  (typeof YTDLP_MAINTAIN_OUTCOME)[keyof typeof YTDLP_MAINTAIN_OUTCOME];

const KNOWN_OUTCOMES: readonly YtdlpMaintainOutcome[] = [
  YTDLP_MAINTAIN_OUTCOME.firstDownload,
  YTDLP_MAINTAIN_OUTCOME.updated,
  YTDLP_MAINTAIN_OUTCOME.unchanged,
  YTDLP_MAINTAIN_OUTCOME.error,
];

export function isYtdlpMaintainOutcome(
  value: string,
): value is YtdlpMaintainOutcome {
  return (KNOWN_OUTCOMES as readonly string[]).includes(value);
}
