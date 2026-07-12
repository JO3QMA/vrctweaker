/** Map backend / network error strings to stable i18n message keys under video.errors.* */

export type VideoUserErrorKey =
  | "githubRateLimit"
  | "githubForbidden"
  | "network"
  | "placeOfficial"
  | "riskAck"
  | "unsupported"
  | "generic";

export function classifyVideoError(raw: string): VideoUserErrorKey {
  const s = raw.trim().toLowerCase();
  if (!s) return "generic";

  if (
    s.includes("risk acknowledgment") ||
    s.includes("risk_ack") ||
    s.includes("errorriskack")
  ) {
    return "riskAck";
  }
  if (
    s.includes("unsupported") ||
    s.includes("windows only") ||
    s.includes("errorunsupported")
  ) {
    return "unsupported";
  }
  if (
    s.includes("403") ||
    s.includes("rate limit") ||
    s.includes("api rate limit") ||
    s.includes("secondary rate")
  ) {
    return "githubRateLimit";
  }
  if (s.includes("401") || s.includes("forbidden")) {
    return "githubForbidden";
  }
  if (
    s.includes("開発者モード") ||
    s.includes("developer mode") ||
    s.includes("配置に失敗") ||
    s.includes("symlink") ||
    s.includes("elevated")
  ) {
    return "placeOfficial";
  }
  if (
    s.includes("timeout") ||
    s.includes("timed out") ||
    s.includes("connection") ||
    s.includes("network") ||
    s.includes("no such host") ||
    s.includes("eof") ||
    s.includes("download request failed") ||
    s.includes("download failed")
  ) {
    return "network";
  }
  return "generic";
}

export function videoErrorI18nKey(
  raw: string,
): `video.errors.${VideoUserErrorKey}` {
  return `video.errors.${classifyVideoError(raw)}`;
}
