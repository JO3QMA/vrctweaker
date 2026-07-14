import { describe, expect, it } from "vitest";
import {
  serverStatusComponentColorClass,
  serverStatusComponentI18nKey,
  serverStatusSummaryColorClass,
  serverStatusSummaryI18nKey,
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
});
