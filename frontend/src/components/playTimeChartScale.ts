export const PLAY_TIME_TEN_MINUTES_SEC = 10 * 60;
export const PLAY_TIME_ONE_HOUR_SEC = 60 * 60;
export const PLAY_TIME_AXIS_MAX_SEC = 24 * PLAY_TIME_ONE_HOUR_SEC;

function hourTickStep(maxHours: number): number {
  const candidates = [1, 2, 3, 4, 6, 8, 12];
  for (const step of candidates) {
    const count = Math.floor(maxHours / step) + 1;
    if (count >= 4 && count <= 9) {
      return step;
    }
  }
  return 1;
}

/** Ceil axis max to the next whole tick so labels use uniform hour steps (no orphan max tick). */
function snapPlayTimeChartMaxYToTicks(maxY: number): number {
  if (maxY <= PLAY_TIME_ONE_HOUR_SEC) {
    return maxY;
  }
  if (maxY < 4 * PLAY_TIME_ONE_HOUR_SEC) {
    const hours = maxY / PLAY_TIME_ONE_HOUR_SEC;
    return Math.ceil(hours) * PLAY_TIME_ONE_HOUR_SEC;
  }
  const maxHours = maxY / PLAY_TIME_ONE_HOUR_SEC;
  const stepH = hourTickStep(maxHours);
  const snappedHours = Math.ceil(maxHours / stepH) * stepH;
  return Math.min(
    snappedHours * PLAY_TIME_ONE_HOUR_SEC,
    PLAY_TIME_AXIS_MAX_SEC,
  );
}

/** Smallest axis max on the play-time grid that fits dataMaxSeconds. */
export function playTimeChartMaxY(dataMaxSeconds: number): number {
  const d = Math.max(0, Math.floor(dataMaxSeconds));
  if (d === 0) {
    return PLAY_TIME_TEN_MINUTES_SEC;
  }
  if (d <= PLAY_TIME_ONE_HOUR_SEC) {
    const steps = Math.ceil(d / PLAY_TIME_TEN_MINUTES_SEC);
    return Math.min(PLAY_TIME_ONE_HOUR_SEC, steps * PLAY_TIME_TEN_MINUTES_SEC);
  }
  const hours = Math.ceil(d / PLAY_TIME_ONE_HOUR_SEC);
  const raw = Math.min(hours * PLAY_TIME_ONE_HOUR_SEC, PLAY_TIME_AXIS_MAX_SEC);
  return snapPlayTimeChartMaxYToTicks(raw);
}

/** Y-axis tick values (seconds); each value lies on the 10-minute / whole-hour grid. */
export function playTimeChartYAxisTicks(maxY: number): number[] {
  if (maxY <= 0) {
    return [0];
  }
  if (maxY <= PLAY_TIME_ONE_HOUR_SEC) {
    const ticks: number[] = [];
    for (let s = 0; s <= maxY; s += PLAY_TIME_TEN_MINUTES_SEC) {
      ticks.push(s);
    }
    return ticks;
  }
  if (maxY < 4 * PLAY_TIME_ONE_HOUR_SEC) {
    const ticks: number[] = [];
    for (let s = 0; s <= maxY; s += PLAY_TIME_ONE_HOUR_SEC) {
      ticks.push(s);
    }
    return ticks;
  }
  const maxHours = maxY / PLAY_TIME_ONE_HOUR_SEC;
  const stepH = hourTickStep(maxHours);
  const ticks: number[] = [];
  for (let h = 0; h * PLAY_TIME_ONE_HOUR_SEC <= maxY; h += stepH) {
    ticks.push(h * PLAY_TIME_ONE_HOUR_SEC);
  }
  return ticks;
}

export type PlayTimeAxisLabelUnits = {
  hour: string;
  minute: string;
};

/** Fixed unit for every Y-axis tick; derived from axis max, not individual tick values. */
export type PlayTimeAxisTickUnit = "hour" | "minute";

/** Hours only when axis max is strictly above one hour (60-minute cap stays in minutes). */
export function playTimeChartAxisTickUnit(maxY: number): PlayTimeAxisTickUnit {
  return maxY > PLAY_TIME_ONE_HOUR_SEC ? "hour" : "minute";
}

export function formatPlayTimeAxisTickLabel(
  sec: number,
  units: PlayTimeAxisLabelUnits,
  axisUnit: PlayTimeAxisTickUnit,
): string {
  if (axisUnit === "hour") {
    // Caller passes Y-axis tick seconds from playTimeChartYAxisTicks (whole-hour multiples only).
    if (sec % PLAY_TIME_ONE_HOUR_SEC !== 0) {
      return `${Math.floor(sec / 60)}${units.minute}`;
    }
    return `${sec / PLAY_TIME_ONE_HOUR_SEC}${units.hour}`;
  }
  return `${Math.floor(sec / 60)}${units.minute}`;
}
