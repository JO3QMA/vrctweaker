import type { Decorator, Meta, StoryObj } from "@storybook/vue3-vite";
import VideoView from "./VideoView.vue";
import { withWailsApp } from "../stories/wailsDecorator";
import type { YTDLPMaintainStatusDTO } from "../wails/app";

const status: YTDLPMaintainStatusDTO = {
  supported: true,
  unsupportedReason: "",
  maintainDesired: true,
  riskAcknowledged: true,
  effectiveOfficial: true,
  cachePresent: true,
  cacheVersion: "2026.07.04",
  toolsPath:
    "C:\\Users\\example\\AppData\\LocalLow\\VRChat\\VRChat\\Tools\\yt-dlp.exe",
  cachePath:
    "C:\\Users\\example\\AppData\\Local\\vrchat-tweaker\\ytdlp\\yt-dlp.exe",
  pendingError: "",
  latestVersion: "2026.07.04",
  latestTag: "2026.07.04",
  latestDownloadUrl: "https://example.com/yt-dlp.exe",
  latestError: "",
};

function withVideoWails(
  overrides: Partial<YTDLPMaintainStatusDTO> = {},
): Decorator {
  const st = { ...status, ...overrides };
  return withWailsApp({
    GetYTDLPMaintainStatus: () => Promise.resolve(st),
    AcknowledgeYTDLPToolsReplaceRisk: () => Promise.resolve(),
    SetYTDLPToolsReplaceMaintain: () => Promise.resolve(),
    CheckYTDLPLatestRelease: () => Promise.resolve(st),
    UpdateOfficialYTDLPCache: () => Promise.resolve(st),
    OpenYTDLPCacheFolder: () => Promise.resolve(),
    OpenYTDLPToolsFolder: () => Promise.resolve(),
    RuntimeIsWindows: () => Promise.resolve(true),
  });
}

const meta = {
  title: "Views/VideoView",
  component: VideoView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof VideoView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  decorators: [withVideoWails()],
};

export const BundledEffective: Story = {
  decorators: [
    withVideoWails({
      maintainDesired: false,
      effectiveOfficial: false,
      latestVersion: "",
      latestTag: "",
      latestDownloadUrl: "",
    }),
  ],
};

export const GitHubRateLimit: Story = {
  decorators: [
    withVideoWails({
      latestError:
        'github api: 403 Forbidden: {"message":"API rate limit exceeded"}',
    }),
  ],
};
