import type { ActivityStatsDTO, UserEncounterDTO } from "../wails/app";

/**
 * ActivityView が onMounted で呼ぶ Wails を Storybook 用に差し替える。
 */
export function activityViewWailsDecorator(
  encounters: UserEncounterDTO[],
  stats: ActivityStatsDTO,
) {
  return (story: () => unknown) => {
    let prevGo: typeof window.go;
    return {
      components: { story },
      template: "<story />",
      created() {
        prevGo = window.go;
        window.go = {
          main: {
            App: {
              Encounters: () => Promise.resolve(encounters),
              GetActivityStats: (_from: string, _to: string) =>
                Promise.resolve(stats),
              ResolveUserProfileNavigation: (id: string) =>
                Promise.resolve({
                  user: {
                    vrcUserId: id,
                    displayName: "mock",
                    status: "",
                    isFavorite: false,
                    lastUpdated: "",
                  },
                  openInFriendsView: false,
                }),
            } as unknown as NonNullable<
              NonNullable<typeof window.go>["main"]
            >["App"],
          },
        };
      },
      beforeUnmount() {
        window.go = prevGo;
      },
    };
  };
}

export const sampleActivityEncounters: UserEncounterDTO[] = [
  {
    id: "story-1",
    vrcUserId: "usr_story",
    displayName: "ストーリー太郎",
    instanceId: "inst_story",
    worldId: "wrld_story",
    worldDisplayName: "Sample World",
    joinedAt: new Date(Date.now() - 3_600_000).toISOString(),
    leftAt: new Date(Date.now() - 1_800_000).toISOString(),
  },
];

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
