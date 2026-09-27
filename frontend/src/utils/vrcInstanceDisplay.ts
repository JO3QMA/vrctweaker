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
  | "groupMembers"
  | "unknown";

const REGION_FLAG: Record<string, string> = {
  jp: "🇯🇵",
  use: "🇺🇸",
  us: "🇺🇸",
  eu: "🇪🇺",
};

function parseWorldAndRest(
  instanceKey: string,
): { worldId: string; rest: string } | null {
  const key = instanceKey.trim();
  if (!key.startsWith("wrld_")) return null;
  const colon = key.indexOf(":");
  if (colon <= 0) return null;
  const worldId = key.slice(0, colon);
  const rest = key.slice(colon + 1).trim();
  if (!rest) return null;
  return { worldId, rest };
}

function instanceShortName(rest: string): string {
  const tilde = rest.indexOf("~");
  return tilde >= 0 ? rest.slice(0, tilde) : rest;
}

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
  const flag = REGION_FLAG[region.toLowerCase()];
  if (flag) return flag;
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
  const params = new URLSearchParams({
    worldId: parsed.worldId,
    instanceId: parsed.rest,
  });
  return `https://vrchat.com/home/launch?${params.toString()}`;
}

/** Human-readable instance label + link metadata for encounter history cells. */
export function formatVrcInstanceCell(
  instanceKey: string,
  t: (key: string, params?: Record<string, string>) => string,
): VrcInstanceCell | null {
  const parsed = parseWorldAndRest(instanceKey);
  if (!parsed) return null;
  const href = vrcInstanceWebLaunchUrl(instanceKey);
  if (!href) return null;

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
    title: instanceKey.trim(),
  };
}
