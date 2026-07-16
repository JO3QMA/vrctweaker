export function formatError(e: unknown, fallback: string): string {
  if (typeof e === "string") {
    const msg = e.trim();
    if (msg) return msg;
  }
  if (e instanceof Error && e.message) return e.message;
  if (e && typeof e === "object" && "message" in e) {
    const m = (e as { message: unknown }).message;
    if (typeof m === "string" && m) return m;
  }
  return fallback;
}
