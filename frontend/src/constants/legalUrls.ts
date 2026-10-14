const LEGAL_DOCS_BASE =
  "https://github.com/JO3QMA/vrctweaker/blob/main/docs/legal";

/** Document locale: Japanese UI uses JA docs; all other supported locales use EN docs. */
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
