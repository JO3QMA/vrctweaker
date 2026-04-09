import type { LaunchProfileDTO } from "../wails/app";

const sampleProfiles: LaunchProfileDTO[] = [
  {
    id: "story-profile-1",
    name: "Storybook 既定",
    arguments: "",
    isDefault: true,
  },
];

/**
 * DashboardView の onMounted が呼ぶ Wails を Storybook 用に差し替える。
 */
export function dashboardViewWailsDecorator() {
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
              LaunchProfiles: () => Promise.resolve([...sampleProfiles]),
              LaunchVRChat: () => Promise.resolve(),
              SetStatus: () => Promise.resolve(),
              SetStatusDescription: () => Promise.resolve(),
              SetStatusAndDescription: () => Promise.resolve(),
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
