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
    expect(formatVrcInstanceLabel(key, t)).toBe("Public #88577 [JP]");
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
    const region = t("vrc.regionFallback", { code: "AUS" });
    expect(text).toBe(`Friends+ #41550 ${region}`);
  });

  it("returns null for non-parseable instance id", () => {
    expect(formatVrcInstanceLabel("inst_1", t)).toBeNull();
  });

  it("uses invite+ when private and canRequestInvite follow VRChat segment order", () => {
    const text = formatVrcInstanceLabel(
      "wrld_x:64190~private(usr_x)~canRequestInvite~region(jp)",
      t,
    );
    expect(text).toBe("Invite+ #64190 [JP]");
  });

  it("uses invite+ when canRequestInvite and private appear in any order", () => {
    const text = formatVrcInstanceLabel(
      "wrld_x:64190~canRequestInvite~private(usr_x)",
      t,
    );
    expect(text).toBe("Invite+ #64190");
  });

  it("uses invite+ when canRequestInvite includes user id segment", () => {
    const text = formatVrcInstanceLabel(
      "wrld_x:100~canRequestInvite(usr_x)~private(usr_x)",
      t,
    );
    expect(text).toBe("Invite+ #100");
  });

  it("uses group access type when present after group marker", () => {
    const text = formatVrcInstanceLabel(
      "wrld_x:100~group(grp_x)~groupAccessType(public)",
      t,
    );
    expect(text).toBe("Group Public #100");
  });

  it("keeps full rest when numeric short name is empty (leading tilde)", () => {
    const key = "wrld_x:~usr_segment";
    expect(formatVrcInstanceLabel(key, t)).toBeNull();
  });

  it("labels group shorthand grp segment as group members", () => {
    expect(formatVrcInstanceLabel("wrld_1:1~grp", t)).toBe("Group Members #1");
  });

  it("returns null when privacy segments are not recognized", () => {
    expect(formatVrcInstanceLabel("wrld_x:100~unknown_marker", t)).toBeNull();
  });

  it("ignores nonce segments on real instance keys", () => {
    const text = formatVrcInstanceLabel(
      "wrld_x:12345~private(usr_x)~region(jp)~nonce(abc)",
      t,
    );
    expect(text).toBe("Invite #12345 [JP]");
  });

  it("trims whitespace around privacy and region segments", () => {
    const text = formatVrcInstanceLabel(
      "wrld_x:100~ private(usr_x) ~ region(jp) ",
      t,
    );
    expect(text).toBe("Invite #100 [JP]");
  });

  it("returns null for malformed privacy segment prefixes", () => {
    expect(formatVrcInstanceLabel("wrld_x:100~hidden(", t)).toBeNull();
  });
});
