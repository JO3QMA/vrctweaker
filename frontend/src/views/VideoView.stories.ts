import type { Decorator, Meta, StoryObj } from "@storybook/vue3-vite";
import VideoView from "./VideoView.vue";
import { withWailsApp } from "../stories/wailsDecorator";
import type { VideoPlaybackDTO, YTDLPMaintainStatusDTO } from "../wails/app";

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

const samplePlaybackHistory: VideoPlaybackDTO[] = [
  {
    id: "vp-open",
    attemptedAt: "2026-03-18T14:00:00.000Z",
    url: "https://example.com/video/open",
    outcome: "",
    worldDisplayName: "Sample World",
  },
  {
    id: "vp-ok",
    attemptedAt: "2026-03-18T13:00:00.000Z",
    url: "https://example.com/video/ok",
    outcome: "success",
    resolvedUrl: "https://cdn.example.com/stream.mp4",
  },
  {
    id: "vp-fail",
    attemptedAt: "2026-03-18T12:00:00.000Z",
    url: "https://example.com/video/fail",
    outcome: "failure",
    failureReason:
      "[youtube] example-id: Requested format is not available. Use --list-formats for a list of available formats",
  },
];

function withVideoWails(
  overrides: Partial<YTDLPMaintainStatusDTO> = {},
  historyRows: VideoPlaybackDTO[] = [],
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
    GetLogRetentionDays: () => Promise.resolve(30),
    VideoPlaybackHistory: () => Promise.resolve(historyRows),
    GetYTDLPCookieLinkageStatus: () =>
      Promise.resolve({
        supported: true,
        enabled: false,
        sourceKind: "",
        riskAcknowledged: false,
        browser: "chrome",
      }),
    AcknowledgeYTDLPCookieLinkageRisk: () => Promise.resolve(),
    SetYTDLPCookieLinkageBrowser: () => Promise.resolve(),
    SetYTDLPCookieLinkageCookiesFile: () => Promise.resolve(),
    DisableYTDLPCookieLinkage: () => Promise.resolve(),
    OpenFileDialog: () => Promise.resolve(""),
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

export const HistoryWithRows: Story = {
  decorators: [withVideoWails({}, samplePlaybackHistory)],
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
