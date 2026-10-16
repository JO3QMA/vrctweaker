/**
 * Validates untrusted URLs (e.g. VRChat profile fields) for safe external navigation.
 * Only http: and https: with a host are allowed.
 */

export type SafeHttpUrl =
  { ok: true; href: string } | { ok: false; raw: string };

export function parseSafeHttpUrl(raw: string): SafeHttpUrl {
  const trimmed = raw.trim();
  if (!trimmed) {
    return { ok: false, raw: trimmed };
  }
  try {
    const u = new URL(trimmed);
    if (u.protocol !== "https:" && u.protocol !== "http:") {
      return { ok: false, raw: trimmed };
    }
    if (!u.hostname) {
      return { ok: false, raw: trimmed };
    }
    return { ok: true, href: u.href };
  } catch {
    return { ok: false, raw: trimmed };
  }
}

export function isSafeHttpUrl(raw: string): boolean {
  return parseSafeHttpUrl(raw).ok;
}
