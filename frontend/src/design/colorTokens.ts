/** Brand color tokens (ADR 0019). */
export const BRAND_COLOR_TOKENS = [
  { name: "brand", varName: "--color-brand" },
  { name: "brand-hover", varName: "--color-brand-hover" },
] as const;

/** Neutral color tokens (ADR 0019). */
export const NEUTRAL_COLOR_TOKENS = [
  { role: "bg-base", varName: "--color-bg-base", legacyAlias: "--bg-primary" },
  {
    role: "bg-elevated",
    varName: "--color-bg-elevated",
    legacyAlias: "--bg-secondary",
  },
  {
    role: "bg-muted",
    varName: "--color-bg-muted",
    legacyAlias: "--bg-tertiary",
  },
  {
    role: "text-primary",
    varName: "--color-text-primary",
    legacyAlias: "--text-primary",
  },
  {
    role: "text-secondary",
    varName: "--color-text-secondary",
    legacyAlias: "--text-secondary",
  },
  { role: "text-muted", varName: "--color-text-muted" },
  { role: "border", varName: "--color-border", legacyAlias: "--border" },
] as const;

/** Semantic color catalog (ADR 0019). */
export const SEMANTIC_COLOR_TOKENS = [
  { name: "danger", varName: "--color-danger", legacyAlias: "--danger" },
  { name: "success", varName: "--color-success", legacyAlias: "--success" },
  { name: "warning", varName: "--color-warning" },
  { name: "info", varName: "--color-info" },
] as const;

export type ServerStatusColorKey =
  "operational" | "degraded" | "partial" | "major" | "maintenance" | "unknown";

/** Server status domain colors (ADR 0019). */
export const SERVER_STATUS_COLOR_TOKENS = [
  { key: "operational", varName: "--color-status-operational" },
  { key: "degraded", varName: "--color-status-degraded" },
  { key: "partial", varName: "--color-status-partial" },
  { key: "major", varName: "--color-status-major" },
  { key: "maintenance", varName: "--color-status-maintenance" },
  { key: "unknown", varName: "--color-status-unknown" },
] as const satisfies ReadonlyArray<{
  key: ServerStatusColorKey;
  varName: `--color-status-${string}`;
}>;

export type PresenceColorKey = "join-me" | "active" | "ask-me" | "busy";

/** Presence domain colors (ADR 0019). */
export const PRESENCE_COLOR_TOKENS = [
  {
    key: "join-me",
    bgVar: "--color-presence-join-me",
    borderVar: "--color-presence-join-me-border",
  },
  {
    key: "active",
    bgVar: "--color-presence-active",
    borderVar: "--color-presence-active-border",
  },
  {
    key: "ask-me",
    bgVar: "--color-presence-ask-me",
    borderVar: "--color-presence-ask-me-border",
  },
  {
    key: "busy",
    bgVar: "--color-presence-busy",
    borderVar: "--color-presence-busy-border",
  },
] as const satisfies ReadonlyArray<{
  key: PresenceColorKey;
  bgVar: `--color-presence-${string}`;
  borderVar: `--color-presence-${string}-border`;
}>;

/** Legacy aliases that delegate to App color tokens (migration). */
export const COLOR_LEGACY_ALIASES = [
  { legacy: "--accent", target: "--color-brand" },
  { legacy: "--accent-hover", target: "--color-brand-hover" },
  { legacy: "--bg-primary", target: "--color-bg-base" },
  { legacy: "--bg-secondary", target: "--color-bg-elevated" },
  { legacy: "--bg-tertiary", target: "--color-bg-muted" },
  { legacy: "--text-primary", target: "--color-text-primary" },
  { legacy: "--text-secondary", target: "--color-text-secondary" },
  { legacy: "--border", target: "--color-border" },
  { legacy: "--danger", target: "--color-danger" },
  { legacy: "--success", target: "--color-success" },
] as const;
