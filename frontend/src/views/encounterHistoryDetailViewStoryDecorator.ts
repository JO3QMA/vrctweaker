import { useRouter } from "vue-router";
import type { UserEncounterDTO } from "../wails/app";

const sampleByUser: UserEncounterDTO[] = [
  {
    id: "eh-user-1",
    vrcUserId: "usr_enc_story",
    displayName: "Encounter User",
    instanceId: "inst_u1",
    worldId: "wrld_u1",
    worldDisplayName: "User Story World",
    joinedAt: "2026-03-01T12:00:00+09:00",
    leftAt: "2026-03-01T13:00:00+09:00",
  },
];

const sampleByWorld: UserEncounterDTO[] = [
  {
    id: "eh-world-1",
    vrcUserId: "usr_w1",
    displayName: "Visitor One",
    instanceId: "inst_w1",
    worldId: "wrld_enc_story",
    worldDisplayName: "World Story",
    joinedAt: "2026-03-02T09:00:00+09:00",
    leftAt: "2026-03-02T10:30:00+09:00",
  },
];

export function encounterHistoryDetailRouterDecorator(
  query: Record<string, string>,
) {
  return (story: () => unknown) => ({
    components: { story },
    template: "<story />",
    async mounted() {
      const router = useRouter();
      await router.push({ name: "encounter-history", query });
    },
  });
}

/** EncounterHistoryDetailView 用: ユーザー別・ワールド別で遭遇 API をスタブ */
export function encounterHistoryDetailWailsDecorator(mode: "user" | "world") {
  const byUser = mode === "user" ? sampleByUser : [];
  const byWorld = mode === "world" ? sampleByWorld : [];

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
              EncountersByVRCUserID: () => Promise.resolve(byUser),
              EncountersByWorldID: () => Promise.resolve(byWorld),
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
