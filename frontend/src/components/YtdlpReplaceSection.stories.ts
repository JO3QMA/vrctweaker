import type { Meta, StoryObj } from "@storybook/vue3-vite";
import YtdlpReplaceSection from "./YtdlpReplaceSection.vue";
import { withWailsApp } from "../stories/wailsDecorator";
import type { YTDLPMaintainStatusDTO } from "../wails/app";

const baseStatus: YTDLPMaintainStatusDTO = {
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
  latestCheckedAt: "2026-10-31T00:14:00.000Z",
};

function wailsForStatus(
  status: YTDLPMaintainStatusDTO,
  checkOutcome?: {
    outcome: string;
    previousVersion?: string;
    newVersion?: string;
    detailError?: string;
  },
) {
  return withWailsApp({
    GetYTDLPMaintainStatus: () => Promise.resolve(status),
    RefreshYTDLPLatestRelease: () => Promise.resolve(status),
    AcknowledgeYTDLPToolsReplaceRisk: () => Promise.resolve(),
    SetYTDLPToolsReplaceMaintain: () => Promise.resolve(),
    CheckAndUpdateOfficialYTDLPCache: () =>
      Promise.resolve({
        status,
        outcome: checkOutcome?.outcome ?? "unchanged",
        previousVersion: checkOutcome?.previousVersion ?? status.cacheVersion,
        newVersion: checkOutcome?.newVersion ?? status.latestVersion,
        detailError: checkOutcome?.detailError ?? "",
      }),
    OpenYTDLPCacheFolder: () => Promise.resolve(),
    OpenYTDLPToolsFolder: () => Promise.resolve(),
    RuntimeIsWindows: () => Promise.resolve(true),
  });
}

const meta = {
  title: "Components/YtdlpReplaceSection",
  component: YtdlpReplaceSection,
  tags: ["autodocs"],
  parameters: {
    layout: "padded",
  },
  decorators: [
    (story) => ({
      components: { story },
      template:
        '<div style="max-width: var(--story-max-width, 1280px); padding: 16px;"><story /></div>',
    }),
  ],
} satisfies Meta<typeof YtdlpReplaceSection>;

export default meta;
type Story = StoryObj<typeof meta>;

export const NotDownloaded: Story = {
  decorators: [
    wailsForStatus({
      ...baseStatus,
      maintainDesired: false,
      effectiveOfficial: false,
      cachePresent: false,
      cacheVersion: "",
      latestVersion: "2026.10.26",
      latestCheckedAt: "2026-10-31T09:14:00.000Z",
    }),
  ],
  parameters: { "--story-max-width": "1280px" },
};

export const UpToDate: Story = {
  decorators: [wailsForStatus({ ...baseStatus })],
};

export const UpdateAvailable: Story = {
  decorators: [
    wailsForStatus({
      ...baseStatus,
      cacheVersion: "2026.07.04",
      latestVersion: "2026.10.26",
    }),
  ],
};

export const InProgressChecking: Story = {
  render: () => ({
    components: { YtdlpReplaceSection },
    template: '<YtdlpReplaceSection ref="section" />',
    mounted() {
      const section = this.$refs.section as InstanceType<
        typeof YtdlpReplaceSection
      >;
      section.applyStatusForTest({ ...baseStatus });
      section.setLoadingForTest(false);
      section.setActionPhaseForTest("checking");
    },
  }),
  decorators: [wailsForStatus({ ...baseStatus })],
};

export const ResultUpdated: Story = {
  render: () => ({
    components: { YtdlpReplaceSection },
    template: '<YtdlpReplaceSection ref="section" />',
    mounted() {
      const section = this.$refs.section as InstanceType<
        typeof YtdlpReplaceSection
      >;
      section.applyStatusForTest({
        ...baseStatus,
        cacheVersion: "2026.10.26",
      });
      section.setLoadingForTest(false);
      section.setResultForTest("updated", {
        previous: "2026.07.04",
        next: "2026.10.26",
      });
    },
  }),
  decorators: [wailsForStatus({ ...baseStatus })],
};

export const ResultError: Story = {
  render: () => ({
    components: { YtdlpReplaceSection },
    template: '<YtdlpReplaceSection ref="section" />',
    mounted() {
      const section = this.$refs.section as InstanceType<
        typeof YtdlpReplaceSection
      >;
      section.applyStatusForTest({ ...baseStatus });
      section.setLoadingForTest(false);
      section.setResultForTest("error", {
        error:
          'github api: 403 Forbidden: {"message":"API rate limit exceeded"}',
      });
    },
  }),
  decorators: [wailsForStatus({ ...baseStatus })],
};

export const Narrow600UpToDate: Story = {
  decorators: [
    (story) => ({
      components: { story },
      template: '<div style="max-width: 600px; padding: 16px;"><story /></div>',
    }),
    wailsForStatus({ ...baseStatus }),
  ],
};

export const Narrow600UpdateAvailable: Story = {
  decorators: [
    (story) => ({
      components: { story },
      template: '<div style="max-width: 600px; padding: 16px;"><story /></div>',
    }),
    wailsForStatus({
      ...baseStatus,
      cacheVersion: "2026.07.04",
      latestVersion: "2026.10.26",
    }),
  ],
};
