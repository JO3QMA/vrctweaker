import type { SeedActivityStats } from "./seed-data";
import { E2E_WORLD_ID } from "./seed-data";

/** `VISUAL_FIXED_CLOCK_ISO`（2024-06-15）基準の直近14日。チャートの軸ラベル・データを固定する。 */
export const VISUAL_ACTIVITY_STATS: SeedActivityStats = {
  dailyPlaySeconds: [
    { date: "2024-06-02", seconds: 3600 },
    { date: "2024-06-03", seconds: 5400 },
    { date: "2024-06-04", seconds: 1800 },
    { date: "2024-06-05", seconds: 7200 },
    { date: "2024-06-06", seconds: 4500 },
    { date: "2024-06-07", seconds: 2700 },
    { date: "2024-06-08", seconds: 6300 },
    { date: "2024-06-09", seconds: 3200 },
    { date: "2024-06-10", seconds: 4100 },
    { date: "2024-06-11", seconds: 5200 },
    { date: "2024-06-12", seconds: 2900 },
    { date: "2024-06-13", seconds: 6100 },
    { date: "2024-06-14", seconds: 3800 },
    { date: "2024-06-15", seconds: 4700 },
  ],
  topWorlds: [
    {
      worldId: E2E_WORLD_ID,
      worldName: "E2E Gallery World",
      seconds: 12_600,
      sessions: 8,
    },
  ],
};
