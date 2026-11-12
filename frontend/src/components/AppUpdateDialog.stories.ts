import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { onMounted } from "vue";
import AppUpdateDialog from "./AppUpdateDialog.vue";
import { useAppUpdate } from "../composables/useAppUpdate";
import { withWailsApp } from "../stories/wailsDecorator";

const appUpdateWailsStub = {
  ClaimStartupAppUpdatePrompt: () =>
    Promise.resolve({
      currentVersion: "1.0.0",
      updateAvailable: false,
      shouldPrompt: false,
      latestVersion: "",
      latestTag: "",
      releaseNotes: "",
      downloadUrl: "",
      zipSha256: "",
      releasePageUrl: "",
      autoApplySupported: true,
    }),
  DismissAppUpdate: () => Promise.resolve(),
  ApplyAppUpdate: () => Promise.resolve(),
  OpenAppReleasePage: () => Promise.resolve(),
};

function resetAppUpdateSingleton() {
  const { dialogVisible, status } = useAppUpdate();
  dialogVisible.value = false;
  status.value = null;
}

const meta: Meta<typeof AppUpdateDialog> = {
  title: "Shell/AppUpdateDialog",
  component: AppUpdateDialog,
  decorators: [
    withWailsApp(appUpdateWailsStub, {
      beforeUnmount: resetAppUpdateSingleton,
    }),
  ],
  parameters: {
    viewport: { defaultViewport: "desktop1280" },
  },
};

export default meta;

type Story = StoryObj<typeof meta>;

export const UpdateAvailable: Story = {
  render: () => ({
    components: { AppUpdateDialog },
    setup() {
      onMounted(() => {
        const { dialogVisible, status } = useAppUpdate();
        status.value = {
          currentVersion: "1.0.0",
          updateAvailable: true,
          shouldPrompt: true,
          latestVersion: "2.0.0",
          latestTag: "v2.0.0",
          releaseNotes: "Storybook 用のリリースノートです。",
          downloadUrl: "https://example.com/z.zip",
          zipSha256: "a".repeat(64),
          releasePageUrl: "https://example.com/r",
          autoApplySupported: true,
        };
        dialogVisible.value = true;
      });
    },
    template: "<AppUpdateDialog />",
  }),
};
