import type { UserCacheDTO } from "../../wails/app";

/**
 * FriendsView が onMounted で呼ぶ Wails を Storybook 用に差し替える。
 */
export function wailsFriendsLoggedInDecorator(friends: UserCacheDTO[]) {
  return (story: () => unknown) => {
    let prevGo: typeof window.go;
    return {
      components: { story },
      template: "<story />",
      created() {
        prevGo = window.go;
        const appStub = {
          Friends: () => Promise.resolve(friends),
          IsLoggedIn: () => Promise.resolve(true),
          RefreshFriends: () => Promise.resolve(),
          SetFavorite: () => Promise.resolve(),
        };
        window.go = {
          main: {
            App: appStub as unknown as NonNullable<
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
