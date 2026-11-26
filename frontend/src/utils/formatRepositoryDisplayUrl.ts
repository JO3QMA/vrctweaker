/** Visible repository label: full host+path without `http://` or `https://`. */
export function formatRepositoryDisplayUrl(url: string): string {
  return url.trim().replace(/^https?:\/\//i, "");
}
