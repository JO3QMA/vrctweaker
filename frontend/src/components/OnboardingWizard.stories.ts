import type { Meta, StoryObj } from "@storybook/vue3-vite";
import OnboardingWizard from "./OnboardingWizard.vue";
import { withWailsApp } from "../stories/wailsDecorator";
import { resetSessionUnlockForStorybook } from "@/composables/useSessionUnlock";

const meta: Meta<typeof OnboardingWizard> = {
  title: "Components/OnboardingWizard",
  component: OnboardingWizard,
  decorators: [
    withWailsApp({
      IsLoggedIn: () => Promise.resolve(false),
      GetPathSettings: () =>
        Promise.resolve({
          vrchatPathWindows: "",
          steamPathLinux: "",
          outputLogPath: "",
        }),
      GetOnboardingFolderDefaults: () =>
        Promise.resolve({
          vrchatPathWindows:
            "C:\\Program Files (x86)\\Steam\\steamapps\\common\\VRChat\\launch.exe",
          steamPathLinux: "/usr/bin/steam",
          outputLogPath: "C:\\Users\\You\\AppData\\LocalLow\\VRChat\\VRChat",
          pictureOutputFolder: "C:\\Users\\You\\Pictures\\VRChat",
        }),
      GetSelfProfile: () =>
        Promise.resolve({ displayName: "Story User", vrcUserId: "" }),
      Login: () =>
        Promise.resolve({
          ok: false,
          error: "Storybook ではログインできません",
        }),
      SetFirstLaunchGuideCompleted: () => Promise.resolve(),
      SaveOnboardingFolderSettings: () => Promise.resolve(),
      OpenDirectoryDialog: (_title: string, _initial: string) =>
        Promise.resolve("C:\\Users\\You\\Pictures\\VRChat"),
      OpenFileDialog: (_title: string, _initial: string, _filter: string) =>
        Promise.resolve(
          "C:\\Program Files (x86)\\Steam\\steamapps\\common\\VRChat\\launch.exe",
        ),
    }),
    () => ({
      template: '<div style="height: 100vh"><story /></div>',
    }),
  ],
  parameters: {
    layout: "fullscreen",
  },
};

export default meta;
type Story = StoryObj<typeof OnboardingWizard>;

export const Default: Story = {
  render: () => ({
    components: { OnboardingWizard },
    setup() {
      resetSessionUnlockForStorybook();
      return {};
    },
    template: '<OnboardingWizard @finished="() => undefined" />',
  }),
};
