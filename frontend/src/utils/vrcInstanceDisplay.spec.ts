import { describe, it, expect } from "vitest";
import { createI18n } from "vue-i18n";
import en from "../i18n/locales/en.json";
import {
  formatVrcInstanceCell,
  vrcInstanceWebLaunchUrl,
} from "./vrcInstanceDisplay";

const i18n = createI18n({
  legacy: false,
  locale: "en",
  messages: { en },
});

const t = i18n.global.t as (
  key: string,
  params?: Record<string, string>,
) => string;

describe("vrcInstanceWebLaunchUrl", () => {
  it("builds vrchat.com launch URL from a full instance key", () => {
    const key =
      "wrld_e055f1a3-6fcb-4d19-9945-f0a1c92cc19b:64190~private(usr_x)~region(jp)";
    expect(vrcInstanceWebLaunchUrl(key)).toBe(
      "https://vrchat.com/home/launch?worldId=wrld_e055f1a3-6fcb-4d19-9945-f0a1c92cc19b&instanceId=64190%7Eprivate%28usr_x%29%7Eregion%28jp%29",
    );
  });

  it("returns null for legacy opaque ids", () => {
    expect(vrcInstanceWebLaunchUrl("inst_e2e_001")).toBeNull();
  });
});

describe("formatVrcInstanceCell", () => {
  it("formats public instance with region code", () => {
    const key = "wrld_db637cfb-64f8-4109-977b-6b755482f133:88577~region(jp)";
    const cell = formatVrcInstanceCell(key, t);
    const publicLabel = t("encounterHistory.instanceType.public");
    expect(cell).toEqual({
      text: `${publicLabel} #88577 [JP]`,
      href: "https://vrchat.com/home/launch?worldId=wrld_db637cfb-64f8-4109-977b-6b755482f133&instanceId=88577%7Eregion%28jp%29",
      title: `${publicLabel} #88577 [JP] (${key})`,
    });
  });

  it("maps legacy region alias use to US", () => {
    const cell = formatVrcInstanceCell("wrld_x:100~region(use)", t);
    expect(cell?.text).toContain("[US]");
    expect(cell?.text).not.toContain("[USE]");
  });

  it("formats friends+ with unknown region fallback", () => {
    const cell = formatVrcInstanceCell(
      "wrld_abc:41550~hidden(usr_x)~region(aus)",
      t,
    );
    const friendsPlus = t("encounterHistory.instanceType.friendsPlus");
    const region = t("encounterHistory.regionFallback", { code: "AUS" });
    expect(cell?.text).toBe(`${friendsPlus} #41550 ${region}`);
    expect(cell?.href).toContain("worldId=wrld_abc");
  });

  it("returns null for non-parseable instance id", () => {
    expect(formatVrcInstanceCell("inst_1", t)).toBeNull();
  });

  it("uses invite+ when private and canRequestInvite follow VRChat segment order", () => {
    const cell = formatVrcInstanceCell(
      "wrld_x:64190~private(usr_x)~canRequestInvite~region(jp)",
      t,
    );
    const invitePlus = t("encounterHistory.instanceType.invitePlus");
    expect(cell?.text).toBe(`${invitePlus} #64190 [JP]`);
  });

  it("stays invite when canRequestInvite precedes private (non-standard order)", () => {
    const cell = formatVrcInstanceCell(
      "wrld_x:64190~canRequestInvite~private(usr_x)",
      t,
    );
    const invite = t("encounterHistory.instanceType.invite");
    expect(cell?.text).toBe(`${invite} #64190`);
  });

  it("uses group access type when present after group marker", () => {
    const cell = formatVrcInstanceCell(
      "wrld_x:100~group(grp_x)~groupAccessType(public)",
      t,
    );
    const groupPublic = t("encounterHistory.instanceType.groupPublic");
    expect(cell?.text).toBe(`${groupPublic} #100`);
  });

  it("keeps full rest when numeric short name is empty (leading tilde)", () => {
    const key = "wrld_x:~usr_segment";
    const cell = formatVrcInstanceCell(key, t);
    const publicLabel = t("encounterHistory.instanceType.public");
    expect(cell?.text).toBe(`${publicLabel} #~usr_segment`);
  });
});
