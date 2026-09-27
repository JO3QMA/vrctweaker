import type { ComposerTranslation } from "vue-i18n";

/** Maps backend errors for friend avatar list fetch to safe i18n messages. */
export function friendAvatarListFetchErrorMessage(
  _err: unknown,
  t: ComposerTranslation,
): string {
  if (import.meta.env.DEV && _err) {
    console.error("[FriendAvatarList] fetch failed:", _err);
  }
  return t("friendAvatars.fetchFailedGeneric");
}
