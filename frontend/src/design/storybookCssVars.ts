const VAR_REF_RE = /^var\(\s*(--[\w-]+)\s*\)$/;

const DEFAULT_MAX_VAR_DEPTH = 12;

/**
 * Read a `:root` custom property (Storybook loads `style.css` in preview).
 * Returns `""` when `document` is unavailable (SSR / non-DOM tests) or the chain
 * cannot be resolved within `maxDepth` hops (aligned with `resolveCssVarInMap`).
 */
export function readRootCssCustomProperty(
  varName: `--${string}`,
  maxDepth = DEFAULT_MAX_VAR_DEPTH,
): string {
  if (typeof document === "undefined") {
    return "";
  }
  const style = getComputedStyle(document.documentElement);
  let key = varName;
  for (let i = 0; i < maxDepth; i++) {
    const raw = style.getPropertyValue(key).trim();
    const ref = VAR_REF_RE.exec(raw);
    if (!ref) {
      return raw;
    }
    key = ref[1] as `--${string}`;
  }
  return "";
}
