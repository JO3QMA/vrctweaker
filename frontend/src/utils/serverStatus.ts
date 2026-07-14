/** status.vrchat.com summary indicator → i18n key under dashboard.serverStatus */
export function serverStatusSummaryI18nKey(
  indicator: string | undefined | null,
): string {
  switch ((indicator ?? "").trim().toLowerCase()) {
    case "none":
      return "dashboard.serverStatus.summaryAllOperational";
    case "minor":
      return "dashboard.serverStatus.summaryDegraded";
    case "major":
      return "dashboard.serverStatus.summaryPartialOutage";
    case "critical":
      return "dashboard.serverStatus.summaryMajorOutage";
    case "maintenance":
      return "dashboard.serverStatus.summaryMaintenance";
    default:
      return "dashboard.serverStatus.summaryUnknown";
  }
}

/** statuspage component status → i18n key */
export function serverStatusComponentI18nKey(
  status: string | undefined | null,
): string {
  switch ((status ?? "").trim().toLowerCase()) {
    case "operational":
      return "dashboard.serverStatus.statusOperational";
    case "degraded_performance":
      return "dashboard.serverStatus.statusDegradedPerformance";
    case "partial_outage":
      return "dashboard.serverStatus.statusPartialOutage";
    case "major_outage":
      return "dashboard.serverStatus.statusMajorOutage";
    case "under_maintenance":
      return "dashboard.serverStatus.statusUnderMaintenance";
    default:
      return "dashboard.serverStatus.statusUnknown";
  }
}

export function serverStatusSummaryColorClass(
  indicator: string | undefined | null,
): string {
  switch ((indicator ?? "").trim().toLowerCase()) {
    case "none":
      return "server-status--operational";
    case "minor":
      return "server-status--degraded";
    case "major":
      return "server-status--partial";
    case "critical":
      return "server-status--major";
    case "maintenance":
      return "server-status--maintenance";
    default:
      return "server-status--unknown";
  }
}

export function serverStatusComponentColorClass(
  status: string | undefined | null,
): string {
  switch ((status ?? "").trim().toLowerCase()) {
    case "operational":
      return "server-status--operational";
    case "degraded_performance":
      return "server-status--degraded";
    case "partial_outage":
      return "server-status--partial";
    case "major_outage":
      return "server-status--major";
    case "under_maintenance":
      return "server-status--maintenance";
    default:
      return "server-status--unknown";
  }
}

export const SERVER_STATUS_POLL_MS = 5 * 60 * 1000;

export const SERVER_STATUS_PAGE_URL = "https://status.vrchat.com/";
