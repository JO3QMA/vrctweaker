export type ColorVarName = `--color-${string}`;

export type SimpleColorToken = {
  name: string;
  varName: ColorVarName;
  legacyAlias?: string;
};

/** Brand color tokens (ADR 0019). */
export const BRAND_COLOR_TOKENS = [
  { name: "brand", varName: "--color-brand" },
  { name: "brand-hover", varName: "--color-brand-hover" },
] as const satisfies ReadonlyArray<SimpleColorToken>;

/** Neutral color tokens (ADR 0019). */
export const NEUTRAL_COLOR_TOKENS = [
  { name: "bg-base", varName: "--color-bg-base", legacyAlias: "--bg-primary" },
  {
    name: "bg-elevated",
    varName: "--color-bg-elevated",
    legacyAlias: "--bg-secondary",
  },
  {
    name: "bg-muted",
    varName: "--color-bg-muted",
    legacyAlias: "--bg-tertiary",
  },
  {
    name: "text-primary",
    varName: "--color-text-primary",
    legacyAlias: "--text-primary",
  },
  {
    name: "text-secondary",
    varName: "--color-text-secondary",
    legacyAlias: "--text-secondary",
  },
  { name: "text-muted", varName: "--color-text-muted" },
  { name: "border", varName: "--color-border", legacyAlias: "--border" },
] as const satisfies ReadonlyArray<SimpleColorToken>;

/** Semantic color catalog (ADR 0019). */
export const SEMANTIC_COLOR_TOKENS = [
  { name: "danger", varName: "--color-danger", legacyAlias: "--danger" },
  { name: "success", varName: "--color-success", legacyAlias: "--success" },
  { name: "warning", varName: "--color-warning" },
  { name: "info", varName: "--color-info" },
] as const satisfies ReadonlyArray<SimpleColorToken>;

export type ServerStatusColorName =
  "operational" | "degraded" | "partial" | "major" | "maintenance" | "unknown";

export type ServerStatusColorToken = {
  name: ServerStatusColorName;
  varName: `--color-status-${string}`;
};

/** Server status domain colors (ADR 0019). */
export const SERVER_STATUS_COLOR_TOKENS = [
  { name: "operational", varName: "--color-status-operational" },
  { name: "degraded", varName: "--color-status-degraded" },
  { name: "partial", varName: "--color-status-partial" },
  { name: "major", varName: "--color-status-major" },
  { name: "maintenance", varName: "--color-status-maintenance" },
  { name: "unknown", varName: "--color-status-unknown" },
] as const satisfies ReadonlyArray<ServerStatusColorToken>;

export type PresenceColorName = "join-me" | "active" | "ask-me" | "busy";

export type PresenceColorToken = {
  name: PresenceColorName;
  bgVar: `--color-presence-${string}`;
  borderVar: `--color-presence-${string}-border`;
};

/** Presence domain colors (ADR 0019). */
export const PRESENCE_COLOR_TOKENS = [
  {
    name: "join-me",
    bgVar: "--color-presence-join-me",
    borderVar: "--color-presence-join-me-border",
  },
  {
    name: "active",
    bgVar: "--color-presence-active",
    borderVar: "--color-presence-active-border",
  },
  {
    name: "ask-me",
    bgVar: "--color-presence-ask-me",
    borderVar: "--color-presence-ask-me-border",
  },
  {
    name: "busy",
    bgVar: "--color-presence-busy",
    borderVar: "--color-presence-busy-border",
  },
] as const satisfies ReadonlyArray<PresenceColorToken>;

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
] as const satisfies ReadonlyArray<{
  legacy: string;
  target: ColorVarName;
}>;
