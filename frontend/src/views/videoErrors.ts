/** Map backend / network error strings to stable i18n message keys under video.errors.* */

export type VideoUserErrorKey =
  | "githubRateLimit"
  | "githubUnauthorized"
  | "githubForbidden"
  | "network"
  | "placeOfficial"
  | "riskAck"
  | "unsupported"
  | "generic";

export function classifyVideoError(
  raw: string | null | undefined,
): VideoUserErrorKey {
  const s = (raw ?? "").trim().toLowerCase();
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
    s.includes("rate limit") ||
    s.includes("api rate limit") ||
    s.includes("secondary rate")
  ) {
    return "githubRateLimit";
  }
  if (s.includes("401") || /\bunauthorized\b/.test(s)) {
    return "githubUnauthorized";
  }
  if (s.includes("403") || s.includes("forbidden")) {
    return "githubForbidden";
  }
  if (
    s.includes("developer mode") ||
    s.includes("symlink") ||
    s.includes("elevated") ||
    s.includes("administrator")
  ) {
    return "placeOfficial";
  }
  if (
    s.includes("timeout") ||
    s.includes("timed out") ||
    s.includes("connection") ||
    s.includes("network") ||
    s.includes("no such host") ||
    /\beof\b/.test(s) ||
    s.includes("connection reset") ||
    s.includes("download request failed") ||
    s.includes("download failed")
  ) {
    return "network";
  }
  return "generic";
}

export function videoErrorI18nKey(
  raw: string | null | undefined,
): `video.errors.${VideoUserErrorKey}` {
  return `video.errors.${classifyVideoError(raw)}`;
}
