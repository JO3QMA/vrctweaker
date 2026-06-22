/** Formats a Date as YYYY-MM-DD in the local calendar (not UTC). */
export function localDateISO(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

/** Adds calendar days in local time. */
export function addLocalDays(d: Date, days: number): Date {
  const out = new Date(d);
  out.setDate(out.getDate() + days);
  return out;
}

/** Iterates local calendar dates from start through end (inclusive), as YYYY-MM-DD. */
export function* eachLocalDateISO(
  startISO: string,
  endISO: string,
): Generator<string> {
  const start = new Date(`${startISO}T00:00:00`);
  const end = new Date(`${endISO}T00:00:00`);
  for (let cur = new Date(start); cur <= end; cur.setDate(cur.getDate() + 1)) {
    yield localDateISO(cur);
  }
}
