const ISO_DATE = /^(\d{4})-(\d{2})-(\d{2})$/;

const FRONTMATTER_RE = /^---\r?\n([\s\S]*?)\r?\n---\r?\n?/;

const LAST_UPDATED_LINE_RE =
  /^(?:最終更新|Last updated):\s*(\d{4}-\d{2}-\d{2})\s*\r?\n/im;

/** Document title duplicated by the page `<h1>`; strip before render. */
const LEADING_ATX_H1_RE = /^#\s+.+\r?\n(?:\r?\n)?/;

function parseFrontmatterLastUpdated(yaml: string): string | null {
  const m = yaml.match(
    /^(?:lastUpdated|last-updated):\s*(\d{4}-\d{2}-\d{2})\s*$/m,
  );
  return m?.[1] ?? null;
}

function isValidIsoDate(iso: string): boolean {
  const m = iso.match(ISO_DATE);
  if (!m) {
    return false;
  }
  const y = Number(m[1]);
  const mo = Number(m[2]);
  const d = Number(m[3]);
  const dt = new Date(Date.UTC(y, mo - 1, d));
  return (
    dt.getUTCFullYear() === y &&
    dt.getUTCMonth() === mo - 1 &&
    dt.getUTCDate() === d
  );
}

export type ParsedLegalDocument = {
  lastUpdated: string | null;
  bodyMarkdown: string;
};

/**
 * Reads `lastUpdated` from YAML frontmatter or a “最終更新 / Last updated” line, then
 * removes that metadata from the markdown passed to the renderer.
 */
export function parseLegalDocument(markdown: string): ParsedLegalDocument {
  let text = markdown;
  let lastUpdated: string | null = null;

  const fm = text.match(FRONTMATTER_RE);
  if (fm) {
    const fromYaml = parseFrontmatterLastUpdated(fm[1]);
    if (fromYaml && isValidIsoDate(fromYaml)) {
      lastUpdated = fromYaml;
    }
    text = text.slice(fm[0].length);
  }

  const lineMatch = text.match(LAST_UPDATED_LINE_RE);
  if (lineMatch) {
    const fromLine = lineMatch[1];
    if (isValidIsoDate(fromLine)) {
      lastUpdated = lastUpdated ?? fromLine;
    }
    text = text.replace(LAST_UPDATED_LINE_RE, "");
  }

  text = text.replace(LEADING_ATX_H1_RE, "");

  return {
    lastUpdated,
    bodyMarkdown: text.trimStart(),
  };
}

/** Display date for legal doc header (ja vs other UI locales). */
export function formatLegalLastUpdated(
  isoDate: string,
  uiLocale: string,
): string {
  const m = isoDate.match(ISO_DATE);
  if (!m) {
    return isoDate;
  }
  const y = Number(m[1]);
  const mo = Number(m[2]);
  const d = Number(m[3]);
  if (uiLocale === "ja") {
    return `${y}年${mo}月${d}日`;
  }
  const date = new Date(Date.UTC(y, mo - 1, d));
  return date.toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
    timeZone: "UTC",
  });
}
