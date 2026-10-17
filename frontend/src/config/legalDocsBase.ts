/** Default matches JO3QMA/vrctweaker `main`; override via `VITE_LEGAL_DOCS_BASE` in `frontend/.env`. */
const DEFAULT_LEGAL_DOCS_BASE =
  "https://github.com/JO3QMA/vrctweaker/blob/main/docs/legal";

function trimTrailingSlash(url: string): string {
  return url.endsWith("/") ? url.slice(0, -1) : url;
}

/** Base URL for privacy policy / terms markdown on GitHub (no trailing slash). */
export function legalDocsBaseUrl(): string {
  const fromEnv = import.meta.env.VITE_LEGAL_DOCS_BASE;
  if (typeof fromEnv === "string" && fromEnv.trim() !== "") {
    return trimTrailingSlash(fromEnv.trim());
  }
  return DEFAULT_LEGAL_DOCS_BASE;
}
