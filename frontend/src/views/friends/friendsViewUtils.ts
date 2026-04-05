export {
  copyDisplayName,
  friendDetailStickyHeaderVisible,
  friendProfileBannerUrl,
  friendThumbUrl,
  jsonStringArray,
} from "../../utils/vrcUserCacheDisplay";

export function friendIsOffline(status: string): boolean {
  return !status || status.toLowerCase() === "offline";
}
