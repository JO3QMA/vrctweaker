export function truncateText(text: string, maxLength: number): string {
  if (maxLength <= 0) return "";
  if (text.length <= maxLength) return text;
  if (maxLength <= 3) return "...".slice(0, maxLength);
  return `${text.slice(0, maxLength - 3)}...`;
}

/** Fixed char count for 50% button width; not measured at runtime */
export const REJOIN_WORLD_NAME_MAX_LEN = 12;

export function truncateRejoinWorldName(name: string): string {
  return truncateText(name.trim(), REJOIN_WORLD_NAME_MAX_LEN);
}
