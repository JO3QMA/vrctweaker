import type { PlayDurationUnits } from "../utils/formatPlayDuration";

export const PLAY_TIME_TEN_MINUTES_SEC = 10 * 60;
export const PLAY_TIME_ONE_HOUR_SEC = 60 * 60;
export const PLAY_TIME_AXIS_MAX_SEC = 24 * PLAY_TIME_ONE_HOUR_SEC;

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
  return Math.min(hours * PLAY_TIME_ONE_HOUR_SEC, PLAY_TIME_AXIS_MAX_SEC);
}

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
  const maxHours = maxY / PLAY_TIME_ONE_HOUR_SEC;
  const stepH = hourTickStep(maxHours);
  const ticks: number[] = [];
  for (let h = 0; h * PLAY_TIME_ONE_HOUR_SEC <= maxY; h += stepH) {
    ticks.push(h * PLAY_TIME_ONE_HOUR_SEC);
  }
  if (ticks[ticks.length - 1] !== maxY) {
    ticks.push(maxY);
  }
  return ticks;
}

export function formatPlayTimeAxisTickLabel(
  sec: number,
  units: PlayDurationUnits,
): string {
  if (sec >= PLAY_TIME_ONE_HOUR_SEC && sec % PLAY_TIME_ONE_HOUR_SEC === 0) {
    return `${sec / PLAY_TIME_ONE_HOUR_SEC}${units.hour}`;
  }
  if (sec % PLAY_TIME_TEN_MINUTES_SEC === 0) {
    return `${sec / 60}${units.minute}`;
  }
  return `${sec}${units.second}`;
}
