/** Map backend Cookie linkage error strings to settings.cookieLinkage.errors.* keys */

export type CookieLinkageUserErrorKey =
  | "riskAck"
  | "unsupported"
  | "cookiesFileMissing"
  | "invalidBrowser"
  | "configRead"
  | "generic";

export function classifyCookieLinkageError(
  raw: string | null | undefined,
): CookieLinkageUserErrorKey {
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
    s.includes("cookies file") ||
    s.includes("errorcookiesfilemissing") ||
    s.includes("does not exist")
  ) {
    return "cookiesFileMissing";
  }
  if (s.includes("invalid browser") || s.includes("errorinvalidbrowser")) {
    return "invalidBrowser";
  }
  if (
    s.includes("cookie linkage config read") ||
    s.includes("path is a directory")
  ) {
    return "configRead";
  }
  return "generic";
}

export function cookieLinkageErrorI18nKey(
  raw: string | null | undefined,
): `settings.cookieLinkage.errors.${CookieLinkageUserErrorKey}` {
  return `settings.cookieLinkage.errors.${classifyCookieLinkageError(raw)}`;
}
