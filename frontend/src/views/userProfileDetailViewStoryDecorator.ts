import { useRouter } from "vue-router";
import type { UserEncounterDTO, UserProfileNavigationDTO } from "../wails/app";

const defaultEncounters: UserEncounterDTO[] = [
  {
    id: "up-story-1",
    vrcUserId: "usr_story",
    displayName: "Story User",
    instanceId: "inst_1",
    worldId: "wrld_1",
    worldDisplayName: "Test World",
    joinedAt: "2026-02-01T10:00:00+09:00",
    leftAt: "2026-02-01T11:00:00+09:00",
  },
];

export function userProfileDetailRouterDecorator(
  query: Record<string, string>,
) {
  return (story: () => unknown) => ({
    components: { story },
    template: "<story />",
    async mounted() {
      const router = useRouter();
      await router.push({ name: "user-profile", query });
    },
  });
}

export interface UserProfileDetailWailsConfig {
  resolveUserProfileNavigation?: (
    vrcUserID: string,
  ) => Promise<UserProfileNavigationDTO>;
  encountersByUser?: UserEncounterDTO[];
  encountersByWorld?: UserEncounterDTO[];
}

/** UserProfileDetailView / VrcUserCacheDetail 用の Wails スタブ。 */
export function userProfileDetailWailsDecorator(
  config: UserProfileDetailWailsConfig = {},
) {
  const rowsUser = config.encountersByUser ?? defaultEncounters;
  const rowsWorld = config.encountersByWorld ?? [];
  const resolve =
    config.resolveUserProfileNavigation ??
    ((id: string): Promise<UserProfileNavigationDTO> =>
      Promise.resolve({
        user: {
          vrcUserId: id,
          displayName: "Story User",
          status: "active",
          isFavorite: false,
          lastUpdated: "",
        },
        openInFriendsView: false,
      }));

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
              ResolveUserProfileNavigation: (id: string) => resolve(id),
              SetFavorite: () => Promise.resolve(),
              EncountersByVRCUserID: () => Promise.resolve(rowsUser),
              EncountersByWorldID: () => Promise.resolve(rowsWorld),
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
