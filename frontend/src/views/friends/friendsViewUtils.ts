import type { UserCacheDTO } from "../../wails/app";

export function friendIsOffline(status: string): boolean {
  return !status || status.toLowerCase() === "offline";
}

export function friendThumbUrl(f: UserCacheDTO): string | undefined {
  return (
    f.currentAvatarThumbnailImageUrl ||
    f.profilePicOverrideThumbnail ||
    f.userIcon ||
    f.imageUrl
  );
}

/** プロフィールヘッダー（横幅いっぱいの背景）用。高精細寄りを優先し、無ければ friendThumbUrl と同系のサムネへ。 */
export function friendProfileBannerUrl(f: UserCacheDTO): string | undefined {
  return (
    f.profilePicOverride ||
    f.currentAvatarImageUrl ||
    f.imageUrl ||
    friendThumbUrl(f)
  );
}

export function jsonStringArray(raw: string | undefined): string[] {
  if (!raw?.trim()) return [];
  try {
    const v = JSON.parse(raw) as unknown;
    if (!Array.isArray(v)) return [];
    return v.filter((x): x is string => typeof x === "string");
  } catch {
    return [];
  }
}

export async function copyDisplayName(name: string): Promise<void> {
  const text = name || "";
  if (!text) return;
  try {
    await navigator.clipboard.writeText(text);
  } catch {
    const ta = document.createElement("textarea");
    ta.value = text;
    ta.setAttribute("readonly", "");
    ta.style.position = "fixed";
    ta.style.left = "-9999px";
    document.body.appendChild(ta);
    ta.select();
    try {
      document.execCommand("copy");
    } finally {
      document.body.removeChild(ta);
    }
  }
}
