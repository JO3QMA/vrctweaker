import type { UserEncounterDTO } from "../wails/app";

/** Storybook: VrcUserCacheDetail の遭遇タブで EncountersByVRCUserID を返す。 */
export function wailsEncountersByUserDecorator(rows: UserEncounterDTO[]) {
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
              EncountersByVRCUserID: () => Promise.resolve(rows),
              EncountersByWorldID: () => Promise.resolve([]),
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
