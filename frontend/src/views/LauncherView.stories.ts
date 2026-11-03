import type { Meta, StoryObj } from "@storybook/vue3-vite";
import LauncherView from "./LauncherView.vue";
import { sampleLaunchProfiles } from "../stories/fixtures/launcher";
import { withWailsApp } from "../stories/wailsDecorator";
import type { LaunchArgsParsedDTO } from "../wails/app";

const emptyLaunchArgs = (): LaunchArgsParsedDTO => ({
  noVr: false,
  screenMode: "",
  screenWidth: 0,
  screenHeight: 0,
  fps: 0,
  skipRegistry: false,
  processPriority: -999,
  mainThreadPriority: -999,
  monitor: 0,
  profile: -1,
  enableDebugGui: false,
  enableSDKLogLevels: false,
  enableUdonDebugLogging: false,
  midi: "",
  watchWorlds: false,
  watchAvatars: false,
  ignoreTrackers: "",
  videoDecoding: "",
  disableAMDStutterWorkaround: false,
  osc: "",
  affinity: "",
  enforceWorldServerChecks: false,
  customArmRatio: 0,
  disableShoulderTracking: false,
  enableIkDebugLogging: false,
  calibrationRange: 0,
  freezeTrackingOnDisconnect: false,
  custom: "",
});

const meta = {
  title: "Views/LauncherView",
  component: LauncherView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
  decorators: [
    withWailsApp({
      LaunchProfiles: () => Promise.resolve([...sampleLaunchProfiles]),
      ParseLaunchArgsForGUI: () => Promise.resolve(emptyLaunchArgs()),
      MergeLaunchArgsForGUI: () => Promise.resolve(""),
      SaveLaunchProfile: () => Promise.resolve(undefined),
      DeleteLaunchProfile: () => Promise.resolve(undefined),
      LaunchVRChatWithArgs: () => Promise.resolve(undefined),
      GetLogicalProcessorCount: () => Promise.resolve(16),
    }),
  ],
} satisfies Meta<typeof LauncherView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const ArgsCollapsePanel: Story = {
  name: "Args collapse (1280px)",
  parameters: {
    viewport: { defaultViewport: "desktop1280" },
  },
};

export const ArgsCollapsePanelNarrow: Story = {
  name: "Args collapse (600px)",
  parameters: {
    viewport: { defaultViewport: "mobile600" },
  },
};
