/** Parsed `--name: value` entries from a `:root { ... }` block (comments stripped). */
export type RootCssVarMap = Record<string, string>;

const ROOT_BLOCK_RE = /:root\s*\{([\s\S]*?)\}/;

/** Extract custom properties declared on `:root` from a stylesheet fragment. */
export function parseStyleCssRootDeclarations(css: string): RootCssVarMap {
  const cssWithoutComments = css.replace(/\/\*[\s\S]*?\*\//g, "");
  const match = ROOT_BLOCK_RE.exec(cssWithoutComments);
  if (!match) {
    throw new Error(":root block not found in stylesheet");
  }
  const withoutComments = match[1];
  const result: RootCssVarMap = {};
  for (const part of withoutComments.split(";")) {
    const trimmed = part.trim();
    if (!trimmed) continue;
    const colon = trimmed.indexOf(":");
    if (colon < 0) continue;
    const name = trimmed.slice(0, colon).trim();
    const value = trimmed.slice(colon + 1).trim();
    if (name.startsWith("--")) {
      result[name] = value;
    }
  }
  return result;
}

const VAR_REF_RE = /^var\(\s*(--[\w-]+)\s*\)$/;

/** Resolve a single `var(--x)` hop in a parsed map (no computed/calc). */
export function resolveCssVarInMap(
  vars: RootCssVarMap,
  varName: `--${string}`,
  maxDepth = 12,
): string {
  let key = varName;
  let depth = 0;
  while (depth < maxDepth) {
    const value = vars[key];
    if (value === undefined) {
      return "";
    }
    const ref = VAR_REF_RE.exec(value);
    if (!ref) {
      return value;
    }
    key = ref[1] as `--${string}`;
    depth += 1;
  }
  return "";
}

/** Parse a pixel length such as `16px`. */
export function parseCssPxLength(value: string): number | null {
  const m = /^(\d+(?:\.\d+)?)px$/.exec(value.trim());
  return m ? Number(m[1]) : null;
}

/** Lists numeric scale steps declared as `--prefix{N}` on `:root` (excludes pattern names). */
export function rootNumericScalePx(
  vars: RootCssVarMap,
  prefix: "--space-" | "--font-size-" | "--icon-size-",
): number[] {
  const escaped = prefix.replace(/-/g, "\\-");
  const re = new RegExp(`^${escaped}(\\d+)$`);
  return Object.keys(vars)
    .filter((key) => re.test(key))
    .map((key) => Number(key.slice(prefix.length)))
    .sort((a, b) => a - b);
}
