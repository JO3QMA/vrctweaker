const VAR_REF_RE = /^var\(\s*(--[\w-]+)\s*\)$/;

/**
 * Read a `:root` custom property (Storybook loads `style.css` in preview).
 * Returns `""` when `document` is unavailable (SSR / non-DOM tests).
 * Resolves one `var(--other)` hop so alias tokens show a concrete value in catalogs.
 */
export function readRootCssCustomProperty(varName: `--${string}`): string {
  if (typeof document === "undefined") {
    return "";
  }
  const raw = getComputedStyle(document.documentElement)
    .getPropertyValue(varName)
    .trim();
  const ref = VAR_REF_RE.exec(raw);
  if (!ref) {
    return raw;
  }
  return getComputedStyle(document.documentElement)
    .getPropertyValue(ref[1])
    .trim();
}
