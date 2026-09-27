import { describe, it, expect } from "vitest";
import {
  formatVrcInstanceCell,
  vrcInstanceWebLaunchUrl,
} from "./vrcInstanceDisplay";

const t = (key: string, params?: Record<string, string>) => {
  if (key === "encounterHistory.regionFallback" && params?.code) {
    return `[${params.code}]`;
  }
  if (key === "encounterHistory.openInVrchat" && params?.text) {
    return `Open in VRChat: ${params.text}`;
  }
  const map: Record<string, string> = {
    "encounterHistory.instanceType.public": "Public",
    "encounterHistory.instanceType.friendsPlus": "Friends+",
    "encounterHistory.instanceType.friends": "Friends",
    "encounterHistory.instanceType.invite": "Invite",
    "encounterHistory.instanceType.invitePlus": "Invite+",
    "encounterHistory.instanceType.groupPublic": "Group Public",
    "encounterHistory.instanceType.groupPlus": "Group+",
    "encounterHistory.instanceType.groupMembers": "Group Members",
    "encounterHistory.instanceType.unknown": "Instance",
  };
  return map[key] ?? key;
};

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
  it("formats public instance with region flag", () => {
    const key = "wrld_db637cfb-64f8-4109-977b-6b755482f133:88577~region(jp)";
    const cell = formatVrcInstanceCell(key, t);
    expect(cell).toEqual({
      text: "Public #88577 🇯🇵",
      href: "https://vrchat.com/home/launch?worldId=wrld_db637cfb-64f8-4109-977b-6b755482f133&instanceId=88577%7Eregion%28jp%29",
      title: `Public #88577 🇯🇵 (${key})`,
    });
  });

  it("formats friends+ with unknown region fallback", () => {
    const cell = formatVrcInstanceCell(
      "wrld_abc:41550~hidden(usr_x)~region(aus)",
      t,
    );
    expect(cell?.text).toBe("Friends+ #41550 [AUS]");
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
    expect(cell?.text).toBe("Invite+ #64190 🇯🇵");
  });

  it("stays invite when canRequestInvite precedes private (non-standard order)", () => {
    const cell = formatVrcInstanceCell(
      "wrld_x:64190~canRequestInvite~private(usr_x)",
      t,
    );
    expect(cell?.text).toBe("Invite #64190");
  });

  it("uses group access type when present after group marker", () => {
    const cell = formatVrcInstanceCell(
      "wrld_x:100~group(grp_x)~groupAccessType(public)",
      t,
    );
    expect(cell?.text).toBe("Group Public #100");
  });

  it("keeps short name when rest starts with tilde", () => {
    const key = "wrld_x:~usr_segment";
    const cell = formatVrcInstanceCell(key, t);
    expect(cell?.text).toBe("Public #~usr_segment");
  });
});
