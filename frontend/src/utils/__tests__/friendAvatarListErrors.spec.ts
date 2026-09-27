import { describe, it, expect, vi, afterEach } from "vitest";
import { friendAvatarListFetchErrorMessage } from "../friendAvatarListErrors";

describe("friendAvatarListFetchErrorMessage", () => {
  const t = (key: string) => key;

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("returns generic i18n key", () => {
    expect(friendAvatarListFetchErrorMessage(new Error("secret path"), t)).toBe(
      "friendAvatars.fetchFailedGeneric",
    );
  });

  it("logs in DEV without changing user message", () => {
    vi.stubEnv("DEV", true);
    const spy = vi.spyOn(console, "error").mockImplementation(() => {});
    const err = new Error("backend detail");
    expect(friendAvatarListFetchErrorMessage(err, t)).toBe(
      "friendAvatars.fetchFailedGeneric",
    );
    expect(spy).toHaveBeenCalledWith("[FriendAvatarList] fetch failed:", err);
  });
});
