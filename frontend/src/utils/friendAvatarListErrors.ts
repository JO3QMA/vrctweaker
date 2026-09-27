import type { ComposerTranslation } from "vue-i18n";

const FRIEND_AVATAR_FETCH_FAILED = "friend avatar usage:";

/** Maps backend errors for friend avatar list fetch to safe i18n messages. */
export function friendAvatarListFetchErrorMessage(
  err: unknown,
  t: ComposerTranslation,
): string {
  if (err instanceof Error) {
    const msg = err.message.trim();
    if (msg.startsWith(FRIEND_AVATAR_FETCH_FAILED)) {
      return t("friendAvatars.fetchFailedGeneric");
    }
  }
  return t("friendAvatars.fetchFailedGeneric");
}
