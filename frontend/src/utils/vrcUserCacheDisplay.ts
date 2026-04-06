import type { UserCacheDTO } from "../wails/app";

/** Stored in users_cache.location when VRChat hides instance details (Pipeline). */
export const PIPELINE_LOCATION_UNKNOWN = "pipeline:location_unknown";

/** Label for the location field in profile UI (empty = omit row). */
export function friendLocationLabel(loc: string | undefined): string {
  if (!loc?.trim()) return "";
  if (loc === PIPELINE_LOCATION_UNKNOWN) return "不明";
  return loc;
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

/**
 * ユーザーキャッシュ詳細のコンパクトヘッダーを出すか。
 * 未スクロールでは出さず、表示名行がスクロール領域上端付近まで来たときだけ出す。
 */
export function friendDetailStickyHeaderVisible(opts: {
  scrollTop: number;
  anchorTopViewport: number;
  bodyTopViewport: number;
  edgeSlopPx?: number;
}): boolean {
  const slop = opts.edgeSlopPx ?? 8;
  return (
    opts.scrollTop > 0 && opts.anchorTopViewport <= opts.bodyTopViewport + slop
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
