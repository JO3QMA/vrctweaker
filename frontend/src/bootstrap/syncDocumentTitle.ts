/** Sets `document.title` from `meta.titleKey` and `appTitle` when `titleKey` is a non-empty string. */
export function syncDocumentTitle(
  t: (key: string) => string,
  meta: Record<string, unknown>,
): void {
  const titleKey = meta.titleKey;
  if (typeof titleKey === "string" && titleKey.length > 0) {
    document.title = `${t(titleKey)} - ${t("appTitle")}`;
  }
}
