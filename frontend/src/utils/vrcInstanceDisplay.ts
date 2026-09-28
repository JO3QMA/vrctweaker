export type VrcInstanceTranslate = (
  key: string,
  params?: Record<string, string>,
) => string;

export type VrcInstanceCell = {
  text: string;
  href: string;
  title: string;
};

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
/** VRChat client launch URL; centralized here until a shared urls module exists. */
const VRCHAT_LAUNCH_BASE_URL = "https://vrchat.com/home/launch";

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

function buildLaunchUrl(parsed: { worldId: string; rest: string }): string {
  const params = new URLSearchParams({
    worldId: parsed.worldId,
    instanceId: parsed.rest,
  });
  return `${VRCHAT_LAUNCH_BASE_URL}?${params.toString()}`;
}

/** Instance number before the first `~` in the post-colon rest segment. */
function instanceShortName(rest: string): string {
  const base = rest.trim();
  const tilde = base.indexOf("~");
  const short = tilde >= 0 ? base.slice(0, tilde).trim() : base;
  return short || "";
}

function isCanRequestInviteSegment(segment: string): boolean {
  return /^canrequestinvite(\([^)]*\))?$/.test(segment.trim());
}

function isKnownPrivacySegment(segment: string): boolean {
  const lower = segment.toLowerCase().trim();
  if (!lower) return true;
  if (isCanRequestInviteSegment(lower)) return true;
  if (lower === "grp") return true;
  if (/^region\([^)]*\)$/.test(lower)) return true;
  if (/^nonce\([^)]*\)$/.test(lower)) return true;
  if (
    /^hidden\([^)]*\)$/.test(lower) ||
    /^friends\([^)]*\)$/.test(lower) ||
    /^private\([^)]*\)$/.test(lower) ||
    /^group\([^)]*\)$/.test(lower)
  ) {
    return true;
  }
  return /^groupaccesstype\([^)]*\)$/.test(lower);
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
    const lower = seg.toLowerCase().trim();
    if (isCanRequestInviteSegment(lower)) {
      hasCanRequestInvite = true;
      continue;
    }
    if (/^nonce\([^)]*\)$/.test(lower)) {
      continue;
    }
    if (/^hidden\([^)]*\)$/.test(lower)) {
      privacy = "friendsPlus";
      continue;
    }
    if (/^friends\([^)]*\)$/.test(lower)) {
      privacy = "friends";
      continue;
    }
    if (/^private\([^)]*\)$/.test(lower)) {
      hasPrivate = true;
      privacy = "invite";
      continue;
    }
    if (lower === "grp" || /^group\([^)]*\)$/.test(lower)) {
      privacy = "groupMembers";
      continue;
    }
    const groupAccess = lower.match(/^groupaccesstype\(([^)]*)\)$/);
    if (groupAccess) {
      const access = groupAccess[1].trim();
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
    const m = seg
      .trim()
      .toLowerCase()
      .match(/^region\(([^)]*)\)$/);
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
  if (!shortName) return null;
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

/**
 * VRChat website launch URL for a stored VRChat instance key, or null if not parseable.
 *
 * @internal exported for unit tests only
 */
export function vrcInstanceWebLaunchUrl(instanceKey: string): string | null {
  const parsed = parseWorldAndRest(instanceKey);
  if (!parsed) return null;
  return buildLaunchUrl(parsed);
}

/**
 * Human-readable instance label + link metadata for encounter history cells.
 * Returns null when the key cannot be labeled (same rules as `formatVrcInstanceLabel`).
 */
export function formatVrcInstanceCell(
  instanceKey: string,
  t: VrcInstanceTranslate,
): VrcInstanceCell | null {
  const text = formatVrcInstanceLabel(instanceKey, t);
  if (!text) return null;
  const href = vrcInstanceWebLaunchUrl(instanceKey);
  if (!href) return null;
  const trimmedKey = instanceKey.trim();
  return {
    text,
    href,
    title: `${text} (${trimmedKey})`,
  };
}
