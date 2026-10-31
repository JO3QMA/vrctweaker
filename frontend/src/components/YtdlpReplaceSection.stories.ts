import type { Decorator, Meta, StoryObj } from "@storybook/vue3-vite";
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

function wailsForStatus(status: YTDLPMaintainStatusDTO): Decorator {
  return withWailsApp({
    GetYTDLPMaintainStatus: () => Promise.resolve(status),
    RefreshYTDLPLatestRelease: () => Promise.resolve(status),
    AcknowledgeYTDLPToolsReplaceRisk: () => Promise.resolve(),
    SetYTDLPToolsReplaceMaintain: () => Promise.resolve(),
    CheckAndUpdateOfficialYTDLPCache: () =>
      Promise.resolve({
        status,
        outcome: "unchanged",
        previousVersion: status.cacheVersion,
        newVersion: status.latestVersion,
        detailError: "",
      }),
    OpenYTDLPCacheFolder: () => Promise.resolve(),
    OpenYTDLPToolsFolder: () => Promise.resolve(),
    RuntimeIsWindows: () => Promise.resolve(true),
  });
}

function storyWithInjectedState(
  status: YTDLPMaintainStatusDTO,
  setup: (section: InstanceType<typeof YtdlpReplaceSection>) => void,
): Story {
  return {
    render: () => ({
      components: { YtdlpReplaceSection },
      template: '<YtdlpReplaceSection ref="section" />',
      mounted() {
        const section = this.$refs.section as InstanceType<
          typeof YtdlpReplaceSection
        >;
        section.applyStatusForTest(status);
        section.setLoadingForTest(false);
        setup(section);
      },
    }),
    decorators: [wailsForStatus(status)],
  };
}

const paddedDecorator: Decorator = (story) => ({
  components: { story },
  template:
    '<div style="max-width: var(--story-max-width, 1280px); padding: 16px;"><story /></div>',
});

const narrowDecorator: Decorator = (story) => ({
  components: { story },
  template: '<div style="max-width: 600px; padding: 16px;"><story /></div>',
});

const meta = {
  title: "Components/YtdlpReplaceSection",
  component: YtdlpReplaceSection,
  tags: ["autodocs"],
  parameters: {
    layout: "padded",
  },
  decorators: [paddedDecorator],
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

export const ResultUpToDate: Story = storyWithInjectedState(
  { ...baseStatus },
  (section) => {
    section.setResultForTest("unchanged", { next: "2026.07.04" });
  },
);

export const ResultError: Story = storyWithInjectedState(
  { ...baseStatus },
  (section) => {
    section.setResultForTest("error", {
      error: 'github api: 403 Forbidden: {"message":"API rate limit exceeded"}',
    });
  },
);

export const InProgressChecking: Story = storyWithInjectedState(
  { ...baseStatus },
  (section) => {
    section.setActionPhaseForTest("checking");
  },
);

export const ResultUpdated: Story = storyWithInjectedState(
  { ...baseStatus, cacheVersion: "2026.10.26" },
  (section) => {
    section.setResultForTest("updated", {
      previous: "2026.07.04",
      next: "2026.10.26",
    });
  },
);

export const Narrow600UpToDate: Story = {
  decorators: [narrowDecorator, wailsForStatus({ ...baseStatus })],
};

export const Narrow600UpdateAvailable: Story = {
  decorators: [
    narrowDecorator,
    wailsForStatus({
      ...baseStatus,
      cacheVersion: "2026.07.04",
      latestVersion: "2026.10.26",
    }),
  ],
};

export const Narrow600NotDownloaded: Story = {
  decorators: [
    narrowDecorator,
    wailsForStatus({
      ...baseStatus,
      cachePresent: false,
      cacheVersion: "",
      latestVersion: "2026.10.26",
    }),
  ],
};

export const Narrow600Error: Story = {
  ...storyWithInjectedState({ ...baseStatus }, (section) => {
    section.setResultForTest("error", {
      error: 'github api: 403 Forbidden: {"message":"API rate limit exceeded"}',
    });
  }),
  decorators: [narrowDecorator, wailsForStatus({ ...baseStatus })],
};
