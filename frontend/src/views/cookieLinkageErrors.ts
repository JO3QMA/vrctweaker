/** Map backend Cookie linkage error strings to video.cookieLinkage.errors.* keys */

export type CookieLinkageUserErrorKey =
  | "riskAck"
  | "unsupported"
  | "cookiesFileMissing"
  | "invalidBrowser"
  | "configRead"
  | "generic";

export function classifyCookieLinkageError(
  raw: unknown,
): CookieLinkageUserErrorKey {
  const s = String(raw ?? "").toLowerCase();
  if (
    s.includes("errorriskackrequired") ||
    s.includes("risk acknowledgment required") ||
    s.includes("cookie linkage risk acknowledgment required")
  ) {
    return "riskAck";
  }
  if (
    s.includes("errorunsupportedplatform") ||
    s.includes("windows only") ||
    s.includes("unsupported platform")
  ) {
    return "unsupported";
  }
  if (
    s.includes("errorcookiesfilemissing") ||
    s.includes("cookies file does not exist") ||
    s.includes("cookies file missing")
  ) {
    return "cookiesFileMissing";
  }
  if (s.includes("errorinvalidbrowser") || s.includes("unsupported browser")) {
    return "invalidBrowser";
  }
  if (s.includes("cookie linkage config read")) {
    return "configRead";
  }
  return "generic";
}

export function cookieLinkageErrorI18nKey(
  raw: unknown,
): `video.cookieLinkage.errors.${CookieLinkageUserErrorKey}` {
  return `video.cookieLinkage.errors.${classifyCookieLinkageError(raw)}`;
}
