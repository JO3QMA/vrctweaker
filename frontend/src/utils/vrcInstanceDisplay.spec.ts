import { describe, it, expect } from "vitest";
import {
  formatVrcInstanceCell,
  vrcInstanceWebLaunchUrl,
} from "./vrcInstanceDisplay";

const t = (key: string, params?: Record<string, string>) => {
  if (key === "encounterHistory.regionFallback" && params?.code) {
    return `[${params.code}]`;
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
    const cell = formatVrcInstanceCell(
      "wrld_db637cfb-64f8-4109-977b-6b755482f133:88577~region(jp)",
      t,
    );
    expect(cell).toEqual({
      text: "Public #88577 🇯🇵",
      href: "https://vrchat.com/home/launch?worldId=wrld_db637cfb-64f8-4109-977b-6b755482f133&instanceId=88577%7Eregion%28jp%29",
      title: "wrld_db637cfb-64f8-4109-977b-6b755482f133:88577~region(jp)",
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
});
