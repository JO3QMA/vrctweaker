export type ColorVarName = `--color-${string}`;

export type SimpleColorToken = {
  name: string;
  varName: ColorVarName;
  legacyAlias?: string;
};

export type ColorLegacyAlias = {
  legacy: string;
  target: ColorVarName;
};

function legacyAliasesFrom(
  tokens: ReadonlyArray<SimpleColorToken>,
): ColorLegacyAlias[] {
  return tokens.flatMap((token) =>
    token.legacyAlias
      ? [{ legacy: token.legacyAlias, target: token.varName }]
      : [],
  );
}

/** Brand color tokens (ADR 0019). */
export const BRAND_COLOR_TOKENS = [
  { name: "brand", varName: "--color-brand", legacyAlias: "--accent" },
  {
    name: "brand-hover",
    varName: "--color-brand-hover",
    legacyAlias: "--accent-hover",
  },
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
  { name: "text-inverse", varName: "--color-text-inverse" },
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

/** Legacy aliases derived from token `legacyAlias` fields (single source of truth). */
export const COLOR_LEGACY_ALIASES = legacyAliasesFrom([
  ...BRAND_COLOR_TOKENS,
  ...NEUTRAL_COLOR_TOKENS,
  ...SEMANTIC_COLOR_TOKENS,
]);
