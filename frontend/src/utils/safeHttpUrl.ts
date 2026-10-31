/**
 * Validates untrusted URLs (e.g. VRChat profile fields) for safe external navigation.
 * Only http: and https: with a host are allowed.
 */

declare const safeHttpHrefBrand: unique symbol;

/** An href produced by {@link parseSafeHttpUrl}. */
export type SafeHttpHref = string & { readonly [safeHttpHrefBrand]: true };

export type SafeHttpUrl =
  { ok: true; href: SafeHttpHref } | { ok: false; raw: string };

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
    return { ok: true, href: u.href as SafeHttpHref };
  } catch {
    return { ok: false, raw: trimmed };
  }
}

export function isSafeHttpUrl(raw: string): boolean {
  return parseSafeHttpUrl(raw).ok;
}
