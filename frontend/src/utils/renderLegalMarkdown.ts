import DOMPurify from "dompurify";
import type { Config } from "dompurify";
import { marked } from "marked";
import { parseSafeHttpUrl } from "./safeHttpUrl";

const EXTERNAL_LINK_CLASS = "legal-md-external-link";

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

marked.use({
  gfm: true,
  breaks: false,
  renderer: {
    link({ href, text }) {
      const label = escapeHtml(text);
      if (!href) {
        return label;
      }
      const safe = parseSafeHttpUrl(href);
      if (!safe.ok) {
        return label;
      }
      const encoded = encodeURIComponent(safe.href);
      return `<span class="${EXTERNAL_LINK_CLASS}" role="link" tabindex="0" data-external-href="${encoded}">${label}</span>`;
    },
  },
});

const PURIFY_CONFIG: Config = {
  ALLOWED_TAGS: [
    "h1",
    "h2",
    "h3",
    "h4",
    "h5",
    "h6",
    "p",
    "ul",
    "ol",
    "li",
    "strong",
    "em",
    "del",
    "code",
    "pre",
    "blockquote",
    "table",
    "thead",
    "tbody",
    "tr",
    "th",
    "td",
    "hr",
    "br",
    "span",
  ],
  ALLOWED_ATTR: ["class", "role", "tabindex", "data-external-href"],
};

/**
 * Parses legal markdown to sanitized HTML. Raw HTML in the source is disabled; links are
 * emitted as inert spans opened via {@link openExternalHttpUrl} on click.
 */
export function renderLegalMarkdown(markdown: string): string {
  const rawHtml = marked.parse(markdown, { async: false }) as string;
  return String(DOMPurify.sanitize(rawHtml, PURIFY_CONFIG));
}

export { EXTERNAL_LINK_CLASS };
