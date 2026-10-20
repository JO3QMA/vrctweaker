import { describe, expect, it } from "vitest";
import en from "../../i18n/locales/en.json";
import ja from "../../i18n/locales/ja.json";
import ko from "../../i18n/locales/ko.json";
import zhCN from "../../i18n/locales/zh-CN.json";
import zhTW from "../../i18n/locales/zh-TW.json";
import {
  formatServerMaintenances,
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

describe("formatServerMaintenances", () => {
  const t = makeT(ja);
  const now = new Date(2026, 11, 23, 15, 0, 0);

  it("defines maintenance copy in every locale", () => {
    for (const messages of [ja, en, ko, zhCN, zhTW]) {
      for (const key of maintenanceKeys) {
        expect(messages.dashboard.serverStatus[key]?.length).toBeGreaterThan(0);
      }
      expect(messages.dashboard.serverStatus.maintenanceRangeSameDay).toContain(
        "–",
      );
    }
    expect(ja.dashboard.serverStatus.maintenanceScheduled).toBe("(メンテ予告)");
    expect(ja.dashboard.serverStatus.maintenanceInProgress).toBe("(メンテ中)");
  });

  it("formats a scheduled notice as a same-day window", () => {
    const [row] = formatServerMaintenances(
      [
        {
          name: "API",
          status: "scheduled",
          scheduledFor: at(2026, 12, 23, 12, 34),
          scheduledUntil: at(2026, 12, 23, 23, 45),
          updates: ["Edge cache restart"],
        },
      ],
      t,
      now,
    );
    expect(row?.phase).toBe("scheduled");
    expect(row?.text).toBe("(メンテ予告) API 12/23 12:34–23:45");
    expect(row?.tooltip).toBe("API\n\nEdge cache restart");
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
    );
    expect(row?.phase).toBe("in_progress");
    expect(row?.text).toBe("(メンテ中) API 〜23:45");
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
    );
    expect(row?.text).toBe("(メンテ中) API 〜23:45");
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
    );
    expect(row?.text).toBe("(メンテ中) API 〜12/24 02:00");
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
    );
    expect(row?.text).toBe("(メンテ予告) API 12/23 23:00–12/24 02:00");
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
    );
    expect(row?.text).toBe("(メンテ予告) API 12/23 12:34〜");
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
    );
    expect(row?.text).toBe("(メンテ中) API 12/23 09:05〜");
  });

  it("puts in-progress windows first and hides finished ones", () => {
    const rows = formatServerMaintenances(
      [
        {
          name: "Later",
          status: "scheduled",
          scheduledFor: at(2026, 12, 25, 1, 0),
          scheduledUntil: at(2026, 12, 25, 2, 0),
        },
        {
          name: "Done",
          status: "completed",
          scheduledFor: at(2026, 12, 20, 1, 0),
          scheduledUntil: at(2026, 12, 20, 2, 0),
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
    );
    expect(rows.map((row) => row.name)).toEqual(["Live", "Soon", "Later"]);
  });

  it("formats English scheduled windows with a 12-hour clock", () => {
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
    );
    expect(row?.text).toBe("(Scheduled) API 12/23, 12:34 PM–11:45 PM");
  });

  it("marks an English window with no end as open-ended", () => {
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
    );
    expect(scheduled[0]?.text).toBe("(Scheduled) API 12/23, 12:34 PM〜");
    expect(live[0]?.text).toBe("(In progress) API 12/23, 9:05 AM〜");
  });
});
