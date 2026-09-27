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

const VRCHAT_WORLD_PREFIX = "wrld_";
const VRCHAT_LAUNCH_BASE_URL = "https://vrchat.com/home/launch";

const REGION_CODE: Record<string, string> = {
  jp: "[JP]",
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

function instanceShortName(rest: string): string {
  const tilde = rest.indexOf("~");
  const short = tilde >= 0 ? rest.slice(0, tilde) : rest;
  return short || rest;
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

function regionSuffix(
  region: string,
  t: (key: string, params?: Record<string, string>) => string,
): string {
  if (!region) return "";
  const regionKey = region.toLowerCase();
  if (Object.prototype.hasOwnProperty.call(REGION_CODE, regionKey)) {
    return REGION_CODE[regionKey];
  }
  return t("encounterHistory.regionFallback", {
    code: region.toUpperCase(),
  });
}

function privacyLabel(
  privacy: InstancePrivacy,
  t: (key: string) => string,
): string {
  const key = `encounterHistory.instanceType.${privacy}`;
  return t(key);
}

/** VRChat website launch URL for a stored VRChat instance key, or null if not parseable. */
export function vrcInstanceWebLaunchUrl(instanceKey: string): string | null {
  const parsed = parseWorldAndRest(instanceKey);
  if (!parsed) return null;
  return buildLaunchUrl(parsed);
}

/**
 * Human-readable instance label + link metadata for encounter history cells.
 * Returns null when the key cannot be parsed into a launch URL (legacy inst_*,
 * wrld_* without `:`, empty rest, etc.). UI should fall back to showing the raw stored id.
 */
export function formatVrcInstanceCell(
  instanceKey: string,
  t: (key: string, params?: Record<string, string>) => string,
): VrcInstanceCell | null {
  const parsed = parseWorldAndRest(instanceKey);
  if (!parsed) return null;
  const href = buildLaunchUrl(parsed);

  const trimmedKey = instanceKey.trim();
  const shortName = instanceShortName(parsed.rest);
  const segmentPart = parsed.rest.includes("~")
    ? parsed.rest.slice(parsed.rest.indexOf("~") + 1)
    : "";
  const segments = segmentPart ? segmentPart.split("~") : [];
  const privacy = detectPrivacy(segments);
  const region = extractRegion(segments);
  const label = `${privacyLabel(privacy, t)} #${shortName}`;
  const suffix = regionSuffix(region, t);
  const text = suffix ? `${label} ${suffix}` : label;

  return {
    text,
    href,
    title: `${text} (${trimmedKey})`,
  };
}
