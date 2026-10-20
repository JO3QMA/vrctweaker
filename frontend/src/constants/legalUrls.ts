import { legalDocsBaseUrl } from "../config/legalDocsBase";

/**
 * Maps UI locale to legal document language.
 * Only `ja` selects Japanese docs; any other value (including `en`, `ko`, `zh-*`, empty, or
 * unexpected strings) falls back to English — we do not ship localized legal text beyond JA/EN.
 */
export function legalDocLocale(uiLocale: string): "ja" | "en" {
  return uiLocale === "ja" ? "ja" : "en";
}

export function privacyPolicyUrl(uiLocale: string): string {
  const loc = legalDocLocale(uiLocale);
  return `${legalDocsBaseUrl()}/privacy-policy-${loc}.md`;
}

export function termsOfUseUrl(uiLocale: string): string {
  const loc = legalDocLocale(uiLocale);
  return `${legalDocsBaseUrl()}/terms-${loc}.md`;
}
