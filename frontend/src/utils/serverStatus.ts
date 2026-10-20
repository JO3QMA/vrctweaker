type ServerStatusStyle = {
  i18nKey: string;
  colorClass: string;
};

/** status.vrchat.com summary indicator → i18n key and CSS class. */
const SUMMARY_STATUS: Record<string, ServerStatusStyle> = {
  none: {
    i18nKey: "dashboard.serverStatus.summaryAllOperational",
    colorClass: "server-status--operational",
  },
  minor: {
    i18nKey: "dashboard.serverStatus.summaryDegraded",
    colorClass: "server-status--degraded",
  },
  major: {
    i18nKey: "dashboard.serverStatus.summaryPartialOutage",
    colorClass: "server-status--partial",
  },
  critical: {
    i18nKey: "dashboard.serverStatus.summaryMajorOutage",
    colorClass: "server-status--major",
  },
  maintenance: {
    i18nKey: "dashboard.serverStatus.summaryMaintenance",
    colorClass: "server-status--maintenance",
  },
};

/** statuspage component status → i18n key and CSS class. */
const COMPONENT_STATUS: Record<string, ServerStatusStyle> = {
  operational: {
    i18nKey: "dashboard.serverStatus.statusOperational",
    colorClass: "server-status--operational",
  },
  degraded_performance: {
    i18nKey: "dashboard.serverStatus.statusDegradedPerformance",
    colorClass: "server-status--degraded",
  },
  partial_outage: {
    i18nKey: "dashboard.serverStatus.statusPartialOutage",
    colorClass: "server-status--partial",
  },
  major_outage: {
    i18nKey: "dashboard.serverStatus.statusMajorOutage",
    colorClass: "server-status--major",
  },
  under_maintenance: {
    i18nKey: "dashboard.serverStatus.statusUnderMaintenance",
    colorClass: "server-status--maintenance",
  },
};

const SUMMARY_STATUS_UNKNOWN: ServerStatusStyle = {
  i18nKey: "dashboard.serverStatus.summaryUnknown",
  colorClass: "server-status--unknown",
};

const COMPONENT_STATUS_UNKNOWN: ServerStatusStyle = {
  i18nKey: "dashboard.serverStatus.statusUnknown",
  colorClass: "server-status--unknown",
};

function serverStatusStyle(
  table: Record<string, ServerStatusStyle>,
  fallback: ServerStatusStyle,
  raw: string | undefined | null,
): ServerStatusStyle {
  const key = (raw ?? "").trim().toLowerCase();
  return table[key] ?? fallback;
}

/** status.vrchat.com summary indicator → i18n key under dashboard.serverStatus */
export function serverStatusSummaryI18nKey(
  indicator: string | undefined | null,
): string {
  return serverStatusStyle(SUMMARY_STATUS, SUMMARY_STATUS_UNKNOWN, indicator)
    .i18nKey;
}

/** statuspage component status → i18n key */
export function serverStatusComponentI18nKey(
  status: string | undefined | null,
): string {
  return serverStatusStyle(COMPONENT_STATUS, COMPONENT_STATUS_UNKNOWN, status)
    .i18nKey;
}

export function serverStatusSummaryColorClass(
  indicator: string | undefined | null,
): string {
  return serverStatusStyle(SUMMARY_STATUS, SUMMARY_STATUS_UNKNOWN, indicator)
    .colorClass;
}

export function serverStatusComponentColorClass(
  status: string | undefined | null,
): string {
  return serverStatusStyle(COMPONENT_STATUS, COMPONENT_STATUS_UNKNOWN, status)
    .colorClass;
}

export const SERVER_STATUS_POLL_MS = 5 * 60 * 1000;

export const SERVER_STATUS_PAGE_URL = "https://status.vrchat.com/";

/** statuspage scheduled maintenance passed through GetServerStatus. */
export type ServerMaintenanceInput = {
  id?: string;
  name?: string;
  status?: string;
  scheduledFor?: string;
  scheduledUntil?: string;
  updates?: string[] | null;
};

export type MaintenanceDisplayPhase = "scheduled" | "in_progress";

export type MaintenanceDisplayRow = {
  key: string;
  phase: MaintenanceDisplayPhase;
  label: string;
  /** Text after the label, including the leading space. */
  rest: string;
  name: string;
  when: string;
  text: string;
  tooltip: string;
};

export type MaintenanceTranslator = (
  key: string,
  params?: Record<string, string>,
) => string;

const MAINTENANCE_I18N = "dashboard.serverStatus";

function maintenanceStatus(status: string | undefined | null): string {
  return (status ?? "").trim().toLowerCase();
}

function maintenancePhase(status: string): MaintenanceDisplayPhase | "hidden" {
  switch (status) {
    case "in_progress":
    case "verifying":
      return "in_progress";
    case "completed":
    case "resolved":
      return "hidden";
    default:
      return "scheduled";
  }
}

function parseMaintenanceInstant(raw: string | undefined | null): Date | null {
  const trimmed = (raw ?? "").trim();
  if (!trimmed) return null;
  const date = new Date(trimmed);
  if (Number.isNaN(date.getTime())) return null;
  return date;
}

function isSameLocalDay(a: Date, b: Date): boolean {
  return (
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  );
}

function dateParams(date: Date): Record<string, string> {
  const hour = date.getHours();
  const hour12 = hour % 12 === 0 ? 12 : hour % 12;
  return {
    month: String(date.getMonth() + 1),
    day: String(date.getDate()),
    hour: String(hour).padStart(2, "0"),
    minute: String(date.getMinutes()).padStart(2, "0"),
    hour12: String(hour12),
    ampm: hour < 12 ? "AM" : "PM",
  };
}

function formatStamp(date: Date, t: MaintenanceTranslator): string {
  return t(`${MAINTENANCE_I18N}.maintenanceDateTime`, dateParams(date));
}

function formatClock(date: Date, t: MaintenanceTranslator): string {
  return t(`${MAINTENANCE_I18N}.maintenanceClock`, dateParams(date));
}

function formatWhen(
  phase: MaintenanceDisplayPhase,
  start: Date | null,
  end: Date | null,
  now: Date,
  t: MaintenanceTranslator,
): string {
  if (phase === "in_progress") {
    if (end) {
      if (isSameLocalDay(end, now)) {
        return t(`${MAINTENANCE_I18N}.maintenanceUntilClock`, {
          endClock: formatClock(end, t),
        });
      }
      return t(`${MAINTENANCE_I18N}.maintenanceUntilStamp`, {
        end: formatStamp(end, t),
      });
    }
    return start ? formatOpenStart(start, t) : "";
  }
  if (start && end) {
    if (isSameLocalDay(start, end)) {
      return t(`${MAINTENANCE_I18N}.maintenanceRangeSameDay`, {
        start: formatStamp(start, t),
        endClock: formatClock(end, t),
      });
    }
    return t(`${MAINTENANCE_I18N}.maintenanceRangeSpan`, {
      start: formatStamp(start, t),
      end: formatStamp(end, t),
    });
  }
  if (start) return formatOpenStart(start, t);
  if (end) return formatStamp(end, t);
  return "";
}

function formatOpenStart(start: Date, t: MaintenanceTranslator): string {
  return t(`${MAINTENANCE_I18N}.maintenanceOpenStart`, {
    start: formatStamp(start, t),
  });
}

function visibleName(name: string): string {
  return name
    .replace(/[\r\n]+/g, " ")
    .replace(/\s+/g, " ")
    .trim();
}

function maintenanceTooltip(
  name: string,
  updates: string[] | null | undefined,
): string {
  const bodies = (updates ?? [])
    .map((body) => body.replace(/\r\n/g, "\n").trim())
    .filter(Boolean);
  if (bodies.length === 0) return name;
  return [name, ...bodies].join("\n\n");
}

function maintenanceSortKey(item: {
  phase: MaintenanceDisplayPhase;
  start: Date | null;
}): [number, number] {
  const rank = item.phase === "in_progress" ? 0 : 1;
  const time = item.start ? item.start.getTime() : Number.POSITIVE_INFINITY;
  return [rank, time];
}

/**
 * Visible maintenance lines: in-progress first, then soonest start.
 * Completed windows and scheduled windows that already ended are omitted.
 * In-progress lines show only the end time (with the date when it is not today).
 */
export function formatServerMaintenances(
  items: readonly ServerMaintenanceInput[] | null | undefined,
  t: MaintenanceTranslator,
  now: Date = new Date(),
): MaintenanceDisplayRow[] {
  const prepared: Array<{
    key: string;
    phase: MaintenanceDisplayPhase;
    name: string;
    start: Date | null;
    end: Date | null;
    tooltip: string;
    order: number;
  }> = [];

  (items ?? []).forEach((item, index) => {
    const status = maintenanceStatus(item.status);
    const phase = maintenancePhase(status);
    if (phase === "hidden") return;
    const name = visibleName(item.name ?? "");
    if (!name) return;
    const start = parseMaintenanceInstant(item.scheduledFor);
    const end = parseMaintenanceInstant(item.scheduledUntil);
    if (phase === "scheduled" && end && end.getTime() <= now.getTime()) {
      return;
    }
    const id = (item.id ?? "").trim();
    prepared.push({
      key: id || `${index}:${status}:${item.scheduledFor ?? ""}:${name}`,
      phase,
      name,
      start,
      end,
      tooltip: maintenanceTooltip(name, item.updates),
      order: index,
    });
  });

  prepared.sort((a, b) => {
    const [ra, ta] = maintenanceSortKey(a);
    const [rb, tb] = maintenanceSortKey(b);
    if (ra !== rb) return ra - rb;
    if (ta !== tb) return ta - tb;
    return a.order - b.order;
  });

  return prepared.map((row) => {
    const label = t(
      row.phase === "in_progress"
        ? `${MAINTENANCE_I18N}.maintenanceInProgress`
        : `${MAINTENANCE_I18N}.maintenanceScheduled`,
    );
    const when = formatWhen(row.phase, row.start, row.end, now, t);
    const text = when ? `${label} ${row.name} ${when}` : `${label} ${row.name}`;
    return {
      key: row.key,
      phase: row.phase,
      label,
      rest: text.slice(label.length),
      name: row.name,
      when,
      text,
      tooltip: row.tooltip,
    };
  });
}
