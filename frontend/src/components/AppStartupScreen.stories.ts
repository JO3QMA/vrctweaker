import type { Meta, StoryObj } from "@storybook/vue3-vite";
import AppStartupScreen from "./AppStartupScreen.vue";
import type { StartupInitStatus } from "../composables/useStartupInit";

const failureStatus: StartupInitStatus = {
  ready: false,
  loading: false,
  phase: "",
  summaryKey: "app.startup.failed.databaseOpen",
  detail:
    "apply schema: friend avatar observations unique index: UNIQUE constraint failed",
  dataDir: "C:\\Users\\Example\\AppData\\Roaming\\vrchat-tweaker",
  canOpenDataFolder: true,
};

const meta: Meta<typeof AppStartupScreen> = {
  title: "Shell/AppStartupScreen",
  component: AppStartupScreen,
  parameters: {
    layout: "fullscreen",
  },
  args: {
    retrying: false,
  },
};

export default meta;
type Story = StoryObj<typeof AppStartupScreen>;

export const Loading: Story = {
  args: {
    status: {
      ready: false,
      loading: true,
      phase: "opening",
      summaryKey: "",
      detail: "",
      dataDir: "",
      canOpenDataFolder: false,
    },
  },
};

export const Migrating: Story = {
  args: {
    status: {
      ready: false,
      loading: true,
      phase: "dedupe",
      summaryKey: "",
      detail: "",
      dataDir: "",
      canOpenDataFolder: false,
    },
  },
};

export const MigratingWide: Story = {
  name: "Migrating (1280px)",
  parameters: {
    viewport: { defaultViewport: "desktop1280" },
  },
  args: {
    status: {
      ready: false,
      loading: true,
      phase: "dedupe",
      summaryKey: "",
      detail: "",
      dataDir: "",
      canOpenDataFolder: false,
    },
  },
};

export const MigratingNarrow: Story = {
  name: "Migrating (600px)",
  parameters: {
    viewport: { defaultViewport: "mobile600" },
  },
  args: {
    status: {
      ready: false,
      loading: true,
      phase: "dedupe",
      summaryKey: "",
      detail: "",
      dataDir: "",
      canOpenDataFolder: false,
    },
  },
};

export const DatabaseFailureWide: Story = {
  name: "Failure (1280px)",
  parameters: {
    viewport: { defaultViewport: "desktop1280" },
  },
  args: {
    status: failureStatus,
  },
};

export const DatabaseFailureNarrow: Story = {
  name: "Failure (600px)",
  parameters: {
    viewport: { defaultViewport: "mobile600" },
  },
  args: {
    status: failureStatus,
  },
};

export const FailureNoDataFolder: Story = {
  args: {
    status: {
      ...failureStatus,
      summaryKey: "app.startup.failed.dataDirResolve",
      canOpenDataFolder: false,
      dataDir: "",
    },
  },
};
