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

/**
 * Derives display privacy from VRChat instance segments (the part after the first `~`).
 *
 * Processing is left-to-right. Restrictive markers (`hidden`, `friends`, `private`, `group`,
 * `groupAccessType`) overwrite the previous privacy when matched. `canRequestInvite` does not
 * set privacy alone; it upgrades `invite` to `invitePlus` when it appears after `private(...)`
 * or when privacy is already `invite` (typical key order: `…~private(usr)~canRequestInvite`).
 */
function detectPrivacy(segments: string[]): InstancePrivacy {
  let privacy: InstancePrivacy = "public";
  let sawPrivate = false;
  for (const seg of segments) {
    const lower = seg.toLowerCase();
    if (lower === "canrequestinvite" || lower === "canrequestinvite()") {
      if (sawPrivate || privacy === "invite") {
        privacy = "invitePlus";
      }
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
      sawPrivate = true;
      privacy = "invite";
      continue;
    }
    if (lower.startsWith("group(")) {
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

function privacyLabel(
  privacy: InstancePrivacy,
  t: VrcInstanceTranslate,
): string {
  const key = `vrc.instanceType.${privacy}`;
  return t(key);
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
  const privacy = detectPrivacy(segments);
  const region = extractRegion(segments);
  const label = `${privacyLabel(privacy, t)} #${shortName}`;
  const suffix = regionSuffix(region, t);
  return suffix ? `${label} ${suffix}` : label;
}
