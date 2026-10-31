import type { Decorator, Meta, StoryObj } from "@storybook/vue3-vite";
import YtdlpReplaceSection from "./YtdlpReplaceSection.vue";
import type { YTDLPMaintainStatusDTO } from "../wails/app";
import {
  backgroundCheckFailedStatus,
  mockCheckAndUpdateResult,
  notDownloadedStatus,
  upToDateStatus,
  updateAvailableStatus,
  wailsForYtdlpStatus,
} from "../stories/fixtures/ytdlpMaintainWails";
import { YTDLP_MAINTAIN_OUTCOME } from "../utils/ytdlpMaintainOutcome";
import { withWailsApp } from "../stories/wailsDecorator";

const rateLimitDetailError =
  'github api: 403 Forbidden: {"message":"API rate limit exceeded"}';

function storyWithInjectedState(
  status: YTDLPMaintainStatusDTO,
  setup: (section: InstanceType<typeof YtdlpReplaceSection>) => void,
): StoryObj {
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
    decorators: [wailsForYtdlpStatus(status)],
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
    docs: {
      description: {
        component:
          "Row 1: maintain toggle. Row 2: version summary. Row 3: primary update/fetch, then bordered folder buttons (secondary, #3d4f66 border).",
      },
    },
  },
  decorators: [paddedDecorator],
} satisfies Meta<typeof YtdlpReplaceSection>;

export default meta;
type Story = StoryObj<typeof meta>;

export const NotDownloaded: Story = {
  decorators: [wailsForYtdlpStatus(notDownloadedStatus())],
};

export const UpToDate: Story = {
  decorators: [wailsForYtdlpStatus(upToDateStatus())],
};

export const UpdateAvailable: Story = {
  decorators: [wailsForYtdlpStatus(updateAvailableStatus())],
};

export const BackgroundCheckFailed: Story = {
  decorators: [wailsForYtdlpStatus(backgroundCheckFailedStatus())],
};

export const ResultUpToDate: Story = storyWithInjectedState(
  upToDateStatus(),
  (section) => {
    section.setResultForTest(YTDLP_MAINTAIN_OUTCOME.unchanged, {
      next: "2026.07.04",
    });
  },
);

const injectedError = storyWithInjectedState(upToDateStatus(), (section) => {
  section.setResultForTest(YTDLP_MAINTAIN_OUTCOME.error, {
    error: rateLimitDetailError,
  });
});

export const ResultError: Story = injectedError;

export const InProgressChecking: Story = storyWithInjectedState(
  upToDateStatus(),
  (section) => {
    section.setActionPhaseForTest("checking");
  },
);

export const ResultUpdated: Story = storyWithInjectedState(
  upToDateStatus({ cacheVersion: "2026.10.26" }),
  (section) => {
    section.setResultForTest(YTDLP_MAINTAIN_OUTCOME.updated, {
      previous: "2026.07.04",
      next: "2026.10.26",
    });
  },
);

export const Narrow600UpToDate: Story = {
  decorators: [narrowDecorator, wailsForYtdlpStatus(upToDateStatus())],
};

export const Narrow600UpdateAvailable: Story = {
  decorators: [narrowDecorator, wailsForYtdlpStatus(updateAvailableStatus())],
};

export const Narrow600NotDownloaded: Story = {
  decorators: [narrowDecorator, wailsForYtdlpStatus(notDownloadedStatus())],
};

export const Narrow600BackgroundCheckFailed: Story = {
  decorators: [
    narrowDecorator,
    wailsForYtdlpStatus(backgroundCheckFailedStatus()),
  ],
};

export const Narrow600Error: Story = {
  ...injectedError,
  decorators: [narrowDecorator, wailsForYtdlpStatus(upToDateStatus())],
};

/** Button press after rate limit — uses live mock for CheckAndUpdate */
export const RateLimitOnButtonPress: Story = {
  decorators: [
    wailsForYtdlpStatus(upToDateStatus()),
    withWailsApp({
      CheckAndUpdateOfficialYTDLPCache: () =>
        Promise.resolve(
          mockCheckAndUpdateResult(
            upToDateStatus(),
            YTDLP_MAINTAIN_OUTCOME.error,
            {
              detailError: rateLimitDetailError,
            },
          ),
        ),
    }),
  ],
};
