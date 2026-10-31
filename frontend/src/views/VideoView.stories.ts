import type { Decorator, Meta, StoryObj } from "@storybook/vue3-vite";
import VideoView from "./VideoView.vue";
import { withWailsApp } from "../stories/wailsDecorator";
import type { VideoPlaybackDTO, YTDLPMaintainStatusDTO } from "../wails/app";
import {
  upToDateStatus,
  ytdlpWailsMocks,
} from "../stories/fixtures/ytdlpMaintainWails";

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
  ytdlpStatus: YTDLPMaintainStatusDTO = upToDateStatus(),
  historyRows: VideoPlaybackDTO[] = [],
  historyLoader?: () => Promise<VideoPlaybackDTO[]>,
): Decorator {
  return withWailsApp({
    ...ytdlpWailsMocks(ytdlpStatus),
    GetLogRetentionDays: () => Promise.resolve(30),
    VideoPlaybackHistory: () =>
      historyLoader ? historyLoader() : Promise.resolve(historyRows),
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
  decorators: [withVideoWails(upToDateStatus(), samplePlaybackHistory)],
};

export const HistoryEmpty: Story = {
  decorators: [withVideoWails(upToDateStatus(), [])],
};

export const HistoryLoading: Story = {
  decorators: [
    withVideoWails(upToDateStatus(), [], () => new Promise(() => {})),
  ],
};

export const BundledEffective: Story = {
  decorators: [
    withVideoWails({
      ...upToDateStatus(),
      maintainDesired: false,
      effectiveOfficial: false,
      latestVersion: "",
      latestTag: "",
      latestDownloadUrl: "",
    }),
  ],
};
