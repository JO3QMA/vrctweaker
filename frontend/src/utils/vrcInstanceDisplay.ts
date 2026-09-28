export type VrcInstanceTranslate = (
  key: string,
  params?: Record<string, string>,
) => string;

type InstancePrivacy =
  | "public"
  | "friendsPlus"
  | "friends"
  | "invite"
  | "invitePlus"
  | "groupPublic"
  | "groupPlus"
  | "groupMembers";

/**
 * VRChat access-type words stay English in all UI locales (product choice).
 * Region suffixes still use `vrc.regionFallback` / `REGION_CODE` via `t`.
 */
const INSTANCE_TYPE_LABEL: Record<InstancePrivacy, string> = {
  public: "Public",
  friendsPlus: "Friends+",
  friends: "Friends",
  invite: "Invite",
  invitePlus: "Invite+",
  groupPublic: "Group Public",
  groupPlus: "Group+",
  groupMembers: "Group Members",
};

const VRCHAT_WORLD_PREFIX = "wrld_";

const REGION_CODE: Record<string, string> = {
  jp: "[JP]",
  /** Legacy VRChat region alias (same as `us`). */
  use: "[US]",
  us: "[US]",
  eu: "[EU]",
};

function parseWorldAndRest(
  instanceKey: string,
): { worldId: string; rest: string } | null {
  const key = instanceKey.trim();
  if (!key.startsWith(VRCHAT_WORLD_PREFIX)) return null;
  const colon = key.indexOf(":");
  if (colon <= 0) return null;
  const worldId = key.slice(0, colon);
  const rest = key.slice(colon + 1).trim();
  if (!rest) return null;
  return { worldId, rest };
}

/**
 * Instance number before the first `~` in the post-colon rest segment.
 * When the key starts with `~` (no numeric prefix), the slice is empty and we
 * fall back to the full rest so the UI still shows a distinguishable fragment.
 */
function instanceShortName(rest: string): string {
  const base = rest.trim();
  const tilde = base.indexOf("~");
  const short = tilde >= 0 ? base.slice(0, tilde).trim() : base;
  return short || base;
}

function isCanRequestInviteSegment(segment: string): boolean {
  return /^canrequestinvite(\([^)]*\))?$/i.test(segment.trim());
}

function isKnownPrivacySegment(segment: string): boolean {
  const lower = segment.toLowerCase().trim();
  if (!lower) return true;
  if (isCanRequestInviteSegment(lower)) return true;
  if (lower === "grp") return true;
  if (/^region\([^)]*\)$/i.test(lower)) return true;
  if (
    lower.startsWith("hidden(") ||
    lower.startsWith("friends(") ||
    lower.startsWith("private(") ||
    lower.startsWith("group(")
  ) {
    return true;
  }
  return /^groupaccesstype\([^)]*\)/i.test(lower);
}

function hasUnknownPrivacySegment(segments: string[]): boolean {
  return segments.some((seg) => !isKnownPrivacySegment(seg));
}

/**
 * Derives display privacy from VRChat instance segments (the part after the first `~`).
 *
 * Restrictive markers are applied left-to-right. `canRequestInvite` (with or without `(...)`)
 * is collected during the pass; after all segments, invite becomes invite+ when both private
 * and can-request-invite are present (segment order independent).
 */
function detectPrivacy(segments: string[]): InstancePrivacy {
  let privacy: InstancePrivacy = "public";
  let hasCanRequestInvite = false;
  let hasPrivate = false;
  for (const seg of segments) {
    const lower = seg.toLowerCase();
    if (isCanRequestInviteSegment(lower)) {
      hasCanRequestInvite = true;
      continue;
    }
    if (lower.startsWith("hidden(")) {
      privacy = "friendsPlus";
      continue;
    }
    if (lower.startsWith("friends(")) {
      privacy = "friends";
      continue;
    }
    if (lower.startsWith("private(")) {
      hasPrivate = true;
      privacy = "invite";
      continue;
    }
    if (lower === "grp" || lower.startsWith("group(")) {
      privacy = "groupMembers";
      continue;
    }
    const groupAccess = lower.match(/^groupaccesstype\(([^)]*)\)/);
    if (groupAccess) {
      const access = groupAccess[1];
      if (access === "public") privacy = "groupPublic";
      else if (access === "plus") privacy = "groupPlus";
      else privacy = "groupMembers";
    }
  }
  if (hasCanRequestInvite && hasPrivate) {
    privacy = "invitePlus";
  }
  return privacy;
}

function extractRegion(segments: string[]): string {
  for (const seg of segments) {
    const m = seg.match(/^region\(([^)]*)\)$/i);
    if (m) return m[1].trim();
  }
  return "";
}

function regionSuffix(region: string, t: VrcInstanceTranslate): string {
  if (!region) return "";
  const regionKey = region.toLowerCase();
  if (Object.prototype.hasOwnProperty.call(REGION_CODE, regionKey)) {
    return REGION_CODE[regionKey];
  }
  return t("vrc.regionFallback", {
    code: region.toUpperCase(),
  });
}

function privacyLabel(privacy: InstancePrivacy): string {
  return INSTANCE_TYPE_LABEL[privacy];
}

/**
 * Human-readable label for a stored VRChat instance key (e.g. public #88577 [JP]).
 * Returns null when the key cannot be parsed (legacy inst_*, malformed wrld_* , etc.).
 */
export function formatVrcInstanceLabel(
  instanceKey: string,
  t: VrcInstanceTranslate,
): string | null {
  const parsed = parseWorldAndRest(instanceKey);
  if (!parsed) return null;

  const shortName = instanceShortName(parsed.rest);
  const segmentPart = parsed.rest.includes("~")
    ? parsed.rest.slice(parsed.rest.indexOf("~") + 1)
    : "";
  const segments = segmentPart ? segmentPart.split("~") : [];
  if (hasUnknownPrivacySegment(segments)) return null;
  const privacy = detectPrivacy(segments);
  const region = extractRegion(segments);
  const label = `${privacyLabel(privacy)} #${shortName}`;
  const suffix = regionSuffix(region, t);
  return suffix ? `${label} ${suffix}` : label;
}
