/** ISO / cache datetime strings for user profile fields (lastLogin, lastActivity, etc.). */
export function formatVrcUserCacheDateTime(
  raw: string | undefined | null,
  calendarLocale = "ja-JP",
): string | null {
  const trimmed = raw?.trim();
  if (!trimmed) return null;
  const d = new Date(trimmed);
  if (Number.isNaN(d.getTime())) return null;
  try {
    return new Intl.DateTimeFormat(calendarLocale, {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "2-digit",
    }).format(d);
  } catch {
    return trimmed;
  }
}
