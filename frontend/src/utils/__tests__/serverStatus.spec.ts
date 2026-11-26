import { afterAll, beforeAll, describe, expect, it, vi } from "vitest";
import en from "../../i18n/locales/en.json";
import ja from "../../i18n/locales/ja.json";
import ko from "../../i18n/locales/ko.json";
import zhCN from "../../i18n/locales/zh-CN.json";
import zhTW from "../../i18n/locales/zh-TW.json";
import {
  formatServerMaintenances,
  localizeMaintenanceTitle,
  localTimeZoneAbbreviation,
  serverStatusComponentColorClass,
  serverStatusComponentI18nKey,
  serverStatusSummaryColorClass,
  serverStatusSummaryI18nKey,
  type MaintenanceTranslator,
} from "../serverStatus";

describe("serverStatus utils", () => {
  it("maps summary indicator to i18n keys", () => {
    expect(serverStatusSummaryI18nKey("none")).toBe(
      "dashboard.serverStatus.summaryAllOperational",
    );
    expect(serverStatusSummaryI18nKey("maintenance")).toBe(
      "dashboard.serverStatus.summaryMaintenance",
    );
  });

  it("maps component status to i18n keys", () => {
    expect(serverStatusComponentI18nKey("under_maintenance")).toBe(
      "dashboard.serverStatus.statusUnderMaintenance",
    );
  });

  it("maps colors for indicator and component status", () => {
    expect(serverStatusSummaryColorClass("none")).toBe(
      "server-status--operational",
    );
    expect(serverStatusComponentColorClass("major_outage")).toBe(
      "server-status--major",
    );
  });

  it("normalizes case and whitespace before lookup", () => {
    expect(serverStatusSummaryI18nKey(" None ")).toBe(
      "dashboard.serverStatus.summaryAllOperational",
    );
    expect(serverStatusSummaryColorClass("MAJOR")).toBe(
      "server-status--partial",
    );
    expect(serverStatusComponentI18nKey(" Under_Maintenance ")).toBe(
      "dashboard.serverStatus.statusUnderMaintenance",
    );
    expect(serverStatusComponentColorClass("")).toBe("server-status--unknown");
  });

  it("ignores inherited object keys such as constructor", () => {
    expect(serverStatusSummaryI18nKey("constructor")).toBe(
      "dashboard.serverStatus.summaryUnknown",
    );
    expect(serverStatusSummaryColorClass("constructor")).toBe(
      "server-status--unknown",
    );
    expect(serverStatusComponentI18nKey("constructor")).toBe(
      "dashboard.serverStatus.statusUnknown",
    );
    expect(serverStatusComponentColorClass("constructor")).toBe(
      "server-status--unknown",
    );
  });
});

const maintenanceKeys = [
  "maintenanceScheduled",
  "maintenanceInProgress",
  "maintenanceDateTime",
  "maintenanceClock",
  "maintenanceRangeSameDay",
  "maintenanceRangeSpan",
  "maintenanceUntilClock",
  "maintenanceUntilStamp",
  "maintenanceOpenStart",
  "maintenanceWhenWithZone",
  "maintenanceTitleMaintenanceWindow",
  "maintenanceTitleScheduledMaintenance",
  "maintenanceTitleEmergencyMaintenance",
  "maintenanceWeekday0",
  "maintenanceWeekday1",
  "maintenanceWeekday2",
  "maintenanceWeekday3",
  "maintenanceWeekday4",
  "maintenanceWeekday5",
  "maintenanceWeekday6",
] as const;

function makeT(messages: {
  dashboard: { serverStatus: Record<string, string> };
}): MaintenanceTranslator {
  return (key, params) => {
    const parts = key.split(".");
    const name = parts[parts.length - 1] ?? "";
    const value = messages.dashboard.serverStatus[name];
    if (!value) throw new Error(`missing i18n key ${key}`);
    if (!params) return value;
    return value.replace(
      /\{(\w+)\}/g,
      (_, token: string) => params[token] ?? "",
    );
  };
}

function at(y: number, m: number, d: number, hh: number, mm: number): string {
  return new Date(y, m - 1, d, hh, mm, 0, 0).toISOString();
}

function timeZoneOnce(when: string, abbrev: string): void {
  expect(when.endsWith(` ${abbrev}`)).toBe(true);
  const matches = when.match(
    new RegExp(`\\b${abbrev.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}\\b`, "g"),
  );
  expect(matches?.length ?? 0).toBe(1);
}

describe("localizeMaintenanceTitle", () => {
  const t = makeT(ja);

  it("maps known English maintenance titles", () => {
    expect(localizeMaintenanceTitle("Maintenance Window", t)).toEqual({
      displayName: "定期メンテナンス",
      originalName: "Maintenance Window",
    });
    expect(
      localizeMaintenanceTitle("  scheduled   maintenance  ", t).displayName,
    ).toBe("予定メンテナンス");
  });

  it("returns the original title when no mapping exists", () => {
    expect(localizeMaintenanceTitle("Partner API rollout", t)).toEqual({
      displayName: "Partner API rollout",
      originalName: "Partner API rollout",
    });
  });
});

describe("formatServerMaintenances", () => {
  const t = makeT(ja);
  const now = new Date(2026, 11, 23, 15, 0, 0);
  const jaOpts = { locale: "ja", timeZoneAbbrev: "JST" };
  const prevTz = process.env.TZ;

  beforeAll(() => {
    vi.stubEnv("TZ", "Asia/Tokyo");
  });

  afterAll(() => {
    if (prevTz === undefined) {
      vi.unstubAllEnvs();
    } else {
      vi.stubEnv("TZ", prevTz);
    }
  });

  it("defines maintenance copy in every locale", () => {
    for (const messages of [ja, en, ko, zhCN, zhTW]) {
      for (const key of maintenanceKeys) {
        expect(messages.dashboard.serverStatus[key]?.length).toBeGreaterThan(0);
      }
      expect(messages.dashboard.serverStatus.maintenanceRangeSameDay).toContain(
        messages === ja ? "〜" : "–",
      );
    }
    expect(ja.dashboard.serverStatus.maintenanceScheduled).toBe("(メンテ予告)");
    expect(ja.dashboard.serverStatus.maintenanceInProgress).toBe("(メンテ中)");
  });

  it("formats a scheduled notice as a same-day window with one time zone", () => {
    const [row] = formatServerMaintenances(
      [
        {
          name: "Maintenance Window",
          status: "scheduled",
          scheduledFor: at(2026, 12, 23, 12, 34),
          scheduledUntil: at(2026, 12, 23, 23, 45),
          updates: ["Edge cache restart"],
        },
      ],
      t,
      now,
      jaOpts,
    );
    expect(row?.phase).toBe("scheduled");
    expect(row?.name).toBe("定期メンテナンス");
    expect(row?.originalName).toBe("Maintenance Window");
    expect(row?.text).toBe(
      "(メンテ予告) 定期メンテナンス 12/23(水) 12:34〜23:45 JST",
    );
    timeZoneOnce(row?.when ?? "", "JST");
    expect(row?.tooltip).toBe("Maintenance Window\n\nEdge cache restart");
  });

  it("formats in-progress maintenance with today's end clock only", () => {
    const [row] = formatServerMaintenances(
      [
        {
          name: "API",
          status: "in_progress",
          scheduledFor: at(2026, 12, 23, 12, 34),
          scheduledUntil: at(2026, 12, 23, 23, 45),
        },
      ],
      t,
      now,
      jaOpts,
    );
    expect(row?.phase).toBe("in_progress");
    expect(row?.text).toBe("(メンテ中) API 〜23:45 JST");
    timeZoneOnce(row?.when ?? "", "JST");
  });

  it("formats verifying the same way as in progress", () => {
    const [row] = formatServerMaintenances(
      [
        {
          name: "API",
          status: "verifying",
          scheduledFor: at(2026, 12, 22, 23, 0),
          scheduledUntil: at(2026, 12, 23, 23, 45),
        },
      ],
      t,
      now,
      jaOpts,
    );
    expect(row?.text).toBe("(メンテ中) API 〜23:45 JST");
  });

  it("includes the end date when an in-progress window ends on another day", () => {
    const [row] = formatServerMaintenances(
      [
        {
          name: "API",
          status: "in_progress",
          scheduledFor: at(2026, 12, 23, 23, 0),
          scheduledUntil: at(2026, 12, 24, 2, 0),
        },
      ],
      t,
      now,
      jaOpts,
    );
    expect(row?.text).toBe("(メンテ中) API 〜12/24(木) 02:00 JST");
    timeZoneOnce(row?.when ?? "", "JST");
  });

  it("formats a scheduled window that crosses midnight with both dates", () => {
    const [row] = formatServerMaintenances(
      [
        {
          name: "API",
          status: "scheduled",
          scheduledFor: at(2026, 12, 23, 23, 0),
          scheduledUntil: at(2026, 12, 24, 2, 0),
        },
      ],
      t,
      now,
      jaOpts,
    );
    expect(row?.text).toBe(
      "(メンテ予告) API 12/23(水) 23:00〜12/24(木) 02:00 JST",
    );
    timeZoneOnce(row?.when ?? "", "JST");
  });

  it("shows only the start when a scheduled window has no end", () => {
    const [row] = formatServerMaintenances(
      [
        {
          name: "API",
          status: "scheduled",
          scheduledFor: at(2026, 12, 23, 12, 34),
        },
      ],
      t,
      now,
      jaOpts,
    );
    expect(row?.text).toBe("(メンテ予告) API 12/23(水) 12:34〜 JST");
  });

  it("shows only the start when an in-progress window has no end", () => {
    const [row] = formatServerMaintenances(
      [
        {
          name: "API",
          status: "in_progress",
          scheduledFor: at(2026, 12, 23, 9, 5),
        },
      ],
      t,
      now,
      jaOpts,
    );
    expect(row?.text).toBe("(メンテ中) API 12/23(水) 09:05〜 JST");
  });

  it("puts in-progress windows first and omits scheduled windows that already ended", () => {
    const rows = formatServerMaintenances(
      [
        {
          name: "Later",
          status: "scheduled",
          scheduledFor: at(2026, 12, 25, 1, 0),
          scheduledUntil: at(2026, 12, 25, 2, 0),
        },
        {
          name: "Ended",
          status: "scheduled",
          scheduledFor: at(2026, 12, 23, 1, 0),
          scheduledUntil: at(2026, 12, 23, 2, 0),
        },
        {
          name: "Soon",
          status: "scheduled",
          scheduledFor: at(2026, 12, 24, 1, 0),
          scheduledUntil: at(2026, 12, 24, 2, 0),
        },
        {
          name: "Live",
          status: "in_progress",
          scheduledFor: at(2026, 12, 23, 12, 0),
          scheduledUntil: at(2026, 12, 23, 18, 0),
        },
      ],
      t,
      now,
      jaOpts,
    );
    expect(rows.map((row) => row.name)).toEqual(["Live", "Soon", "Later"]);
  });

  it("compresses newlines and repeated whitespace in maintenance names", () => {
    const [row] = formatServerMaintenances(
      [
        {
          name: "API\n  edge   cache",
          status: "in_progress",
        },
      ],
      t,
      now,
      jaOpts,
    );
    expect(row?.name).toBe("API edge cache");
    expect(row?.text).toBe("(メンテ中) API edge cache");
  });

  it("formats English scheduled windows with a 12-hour clock and zone once", () => {
    const abbrev = "GMT+9";
    const [row] = formatServerMaintenances(
      [
        {
          name: "API",
          status: "scheduled",
          scheduledFor: at(2026, 12, 23, 12, 34),
          scheduledUntil: at(2026, 12, 23, 23, 45),
        },
      ],
      makeT(en),
      now,
      { locale: "en-US", timeZoneAbbrev: abbrev },
    );
    expect(row?.text).toBe(
      "(Scheduled) API 12/23 (Wed), 12:34 PM–11:45 PM GMT+9",
    );
    timeZoneOnce(row?.when ?? "", abbrev);
  });

  it("marks an English window with no end as open-ended", () => {
    const abbrev = "GMT+9";
    const scheduled = formatServerMaintenances(
      [
        {
          name: "API",
          status: "scheduled",
          scheduledFor: at(2026, 12, 23, 12, 34),
        },
      ],
      makeT(en),
      now,
      { locale: "en-US", timeZoneAbbrev: abbrev },
    );
    const live = formatServerMaintenances(
      [
        {
          name: "API",
          status: "in_progress",
          scheduledFor: at(2026, 12, 23, 9, 5),
        },
      ],
      makeT(en),
      now,
      { locale: "en-US", timeZoneAbbrev: abbrev },
    );
    expect(scheduled[0]?.text).toBe(
      "(Scheduled) API 12/23 (Wed), 12:34 PM〜 GMT+9",
    );
    expect(live[0]?.text).toBe(
      "(In progress) API 12/23 (Wed), 9:05 AM〜 GMT+9",
    );
  });

  it("derives a local time zone abbreviation via Intl", () => {
    const sample = new Date(2026, 11, 1, 10, 0, 0);
    expect(localTimeZoneAbbreviation(sample, "ja")).toBe("JST");
  });
});
