import { describe, it, expect } from "vitest";
import { createI18n } from "vue-i18n";
import en from "../i18n/locales/en.json";
import { formatVrcInstanceLabel } from "./vrcInstanceDisplay";

const i18n = createI18n({
  legacy: false,
  locale: "en",
  messages: { en },
});

const t = i18n.global.t as (
  key: string,
  params?: Record<string, string>,
) => string;

describe("formatVrcInstanceLabel", () => {
  it("formats public instance with region code", () => {
    const key = "wrld_db637cfb-64f8-4109-977b-6b755482f133:88577~region(jp)";
    const publicLabel = t("vrc.instanceType.public");
    expect(formatVrcInstanceLabel(key, t)).toBe(`${publicLabel} #88577 [JP]`);
  });

  it("maps legacy region alias use to US", () => {
    const text = formatVrcInstanceLabel("wrld_x:100~region(use)", t);
    expect(text).toContain("[US]");
    expect(text).not.toContain("[USE]");
  });

  it("formats friends+ with unknown region fallback", () => {
    const text = formatVrcInstanceLabel(
      "wrld_abc:41550~hidden(usr_x)~region(aus)",
      t,
    );
    const friendsPlus = t("vrc.instanceType.friendsPlus");
    const region = t("vrc.regionFallback", { code: "AUS" });
    expect(text).toBe(`${friendsPlus} #41550 ${region}`);
  });

  it("returns null for non-parseable instance id", () => {
    expect(formatVrcInstanceLabel("inst_1", t)).toBeNull();
  });

  it("uses invite+ when private and canRequestInvite follow VRChat segment order", () => {
    const text = formatVrcInstanceLabel(
      "wrld_x:64190~private(usr_x)~canRequestInvite~region(jp)",
      t,
    );
    const invitePlus = t("vrc.instanceType.invitePlus");
    expect(text).toBe(`${invitePlus} #64190 [JP]`);
  });

  it("stays invite when canRequestInvite precedes private (non-standard order)", () => {
    const text = formatVrcInstanceLabel(
      "wrld_x:64190~canRequestInvite~private(usr_x)",
      t,
    );
    const invite = t("vrc.instanceType.invite");
    expect(text).toBe(`${invite} #64190`);
  });

  it("uses group access type when present after group marker", () => {
    const text = formatVrcInstanceLabel(
      "wrld_x:100~group(grp_x)~groupAccessType(public)",
      t,
    );
    const groupPublic = t("vrc.instanceType.groupPublic");
    expect(text).toBe(`${groupPublic} #100`);
  });

  it("keeps full rest when numeric short name is empty (leading tilde)", () => {
    const key = "wrld_x:~usr_segment";
    const publicLabel = t("vrc.instanceType.public");
    expect(formatVrcInstanceLabel(key, t)).toBe(`${publicLabel} #~usr_segment`);
  });
});
