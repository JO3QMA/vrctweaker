import type { ActivityStatsDTO } from "../../wails/app";

export function sampleActivityStats(): ActivityStatsDTO {
  const dailyPlaySeconds: ActivityStatsDTO["dailyPlaySeconds"] = [];
  const d = new Date();
  for (let i = 13; i >= 0; i--) {
    const x = new Date(d);
    x.setDate(x.getDate() - i);
    dailyPlaySeconds.push({
      date: x.toISOString().slice(0, 10),
      seconds: 1800 * ((i % 5) + 1),
    });
  }
  return { dailyPlaySeconds, topWorlds: [] };
}
