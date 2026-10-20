import { legalDocLocale } from "../constants/legalUrls";
import privacyJa from "virtual:legal-md/privacy-policy-ja.md";
import privacyEn from "virtual:legal-md/privacy-policy-en.md";
import termsJa from "virtual:legal-md/terms-ja.md";
import termsEn from "virtual:legal-md/terms-en.md";

export type LegalDocKind = "privacy" | "terms";

const PRIVACY: Record<"ja" | "en", string> = {
  ja: privacyJa,
  en: privacyEn,
};

const TERMS: Record<"ja" | "en", string> = {
  ja: termsJa,
  en: termsEn,
};

export function getLegalDocumentMarkdown(
  kind: LegalDocKind,
  uiLocale: string,
): string {
  const loc = legalDocLocale(uiLocale);
  return kind === "privacy" ? PRIVACY[loc] : TERMS[loc];
}
