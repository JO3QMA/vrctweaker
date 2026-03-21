/**
 * Drop thumbnail URL entries that are not both in the current screenshot list
 * and in the retain set (typically virtualizer-visible grid IDs plus selection).
 */
export function pruneThumbnailUrlMap(
  map: Readonly<Record<string, string>>,
  listIds: ReadonlySet<string>,
  retainedIds: ReadonlySet<string>,
): Record<string, string> {
  const out: Record<string, string> = {};
  for (const [id, url] of Object.entries(map)) {
    if (listIds.has(id) && retainedIds.has(id)) {
      out[id] = url;
    }
  }
  return out;
}
