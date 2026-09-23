export type BioLinkSite =
  | "twitter"
  | "booth"
  | "misskey"
  | "github"
  | "youtube"
  | "discord"
  | "twitch"
  | "vrchat"
  | "pixiv"
  | "instagram"
  | "tiktok"
  | "bluesky"
  | "link";

const HOST_SUFFIX_MATCHERS: Array<{ site: BioLinkSite; hosts: string[] }> = [
  { site: "twitter", hosts: ["twitter.com", "x.com", "mobile.twitter.com"] },
  { site: "booth", hosts: ["booth.pm"] },
  {
    site: "misskey",
    hosts: [
      "misskey.io",
      "misskey.design",
      "misskey-hub.net",
      "mksn.social",
      "mk.shrimpia.network",
    ],
  },
  { site: "github", hosts: ["github.com", "gist.github.com"] },
  { site: "youtube", hosts: ["youtube.com", "youtu.be", "m.youtube.com"] },
  { site: "discord", hosts: ["discord.com", "discord.gg", "discordapp.com"] },
  { site: "twitch", hosts: ["twitch.tv", "www.twitch.tv"] },
  { site: "vrchat", hosts: ["vrchat.com", "www.vrchat.com"] },
  { site: "pixiv", hosts: ["pixiv.net", "www.pixiv.net"] },
  { site: "instagram", hosts: ["instagram.com", "www.instagram.com"] },
  { site: "tiktok", hosts: ["tiktok.com", "www.tiktok.com"] },
  { site: "bluesky", hosts: ["bsky.app", "bsky.social"] },
];

function normalizeHostname(hostname: string): string {
  return hostname.trim().toLowerCase().replace(/\.$/, "");
}

function hostEqualsOrEndsWith(host: string, pattern: string): boolean {
  return host === pattern || host.endsWith(`.${pattern}`);
}

function hostMatchesAny(host: string, patterns: string[]): boolean {
  return patterns.some((pattern) => hostEqualsOrEndsWith(host, pattern));
}

/** Parse a bio link URL; returns null when the value cannot be parsed. */
export function parseBioLinkUrl(raw: string): URL | null {
  const trimmed = raw.trim();
  if (!trimmed) return null;
  try {
    const withProtocol = /^https?:\/\//i.test(trimmed)
      ? trimmed
      : `https://${trimmed}`;
    return new URL(withProtocol);
  } catch {
    return null;
  }
}

function looksLikeMisskeyInstance(host: string): boolean {
  if (host.includes("misskey")) return true;
  return host.endsWith(".mk");
}

/** Detect a known social / VRChat site from a bio link URL. */
export function detectBioLinkSite(url: string): BioLinkSite {
  const parsed = parseBioLinkUrl(url);
  if (!parsed) return "link";

  const host = normalizeHostname(parsed.hostname);
  const pathname = parsed.pathname;

  for (const { site, hosts } of HOST_SUFFIX_MATCHERS) {
    if (site === "misskey") continue;
    if (hostMatchesAny(host, hosts)) return site;
  }

  if (looksLikeMisskeyInstance(host)) return "misskey";

  if (
    hostMatchesAny(host, ["vrchat.com", "www.vrchat.com"]) &&
    pathname.includes("/home/user/")
  ) {
    return "vrchat";
  }

  return "link";
}

/** Short hostname label for display (without scheme). */
export function bioLinkHostname(url: string): string {
  const parsed = parseBioLinkUrl(url);
  return parsed?.hostname.replace(/^www\./i, "") ?? url.trim();
}
