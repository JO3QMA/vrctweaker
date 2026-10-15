const LEGAL_DOCS_BASE =
  "https://github.com/JO3QMA/vrctweaker/blob/main/docs/legal";

/**
 * Maps UI locale to legal document language.
 * Only `ja` selects Japanese docs; any other value (including `en`, `ko`, `zh-*`, empty, or
 * unexpected strings) falls back to English — we do not ship localized legal text beyond JA/EN.
 *
 * Exported for unit tests only. UI code should call {@link privacyPolicyUrl} / {@link termsOfUseUrl}
 * instead; this helper is not a stable public API for other modules.
 */
export function legalDocLocale(uiLocale: string): "ja" | "en" {
  return uiLocale === "ja" ? "ja" : "en";
}

export function privacyPolicyUrl(uiLocale: string): string {
  const loc = legalDocLocale(uiLocale);
  return `${LEGAL_DOCS_BASE}/privacy-policy-${loc}.md`;
}

export function termsOfUseUrl(uiLocale: string): string {
  const loc = legalDocLocale(uiLocale);
  return `${LEGAL_DOCS_BASE}/terms-${loc}.md`;
}
