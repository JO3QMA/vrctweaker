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
      GetOnboardingFolderDefaults: () =>
        Promise.resolve({
          vrchatPathWindows:
            "C:\\Program Files (x86)\\Steam\\steamapps\\common\\VRChat\\launch.exe",
          steamPathLinux: "/usr/bin/steam",
          outputLogPath: "C:\\Users\\You\\AppData\\LocalLow\\VRChat\\VRChat",
          pictureOutputFolder: "C:\\Users\\You\\Pictures\\VRChat",
        }),
      SetFirstLaunchGuideCompleted: () => Promise.resolve(),
      SaveOnboardingFolderSettings: () => Promise.resolve(),
      OpenDirectoryDialog: () => Promise.resolve(""),
      OpenFileDialog: () => Promise.resolve(""),
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
