import type { Decorator, Meta, StoryObj } from "@storybook/vue3-vite";
import CookieLinkageSection from "./CookieLinkageSection.vue";
import { withWailsApp } from "../stories/wailsDecorator";
import type {
  CookieLinkageStatusDTO,
  YTDLPMaintainStatusDTO,
} from "../wails/app";

const cookieOff: CookieLinkageStatusDTO = {
  supported: true,
  enabled: false,
  sourceKind: "",
  riskAcknowledged: false,
  browser: "chrome",
  cookiesFilePath: "",
  configPath: "",
};

const maintainOfficial: YTDLPMaintainStatusDTO = {
  supported: true,
  unsupportedReason: "",
  maintainDesired: true,
  riskAcknowledged: true,
  effectiveOfficial: true,
  cachePresent: true,
  cacheVersion: "2026.07.04",
  toolsPath: "",
  cachePath: "",
  pendingError: "",
  latestVersion: "",
  latestTag: "",
  latestDownloadUrl: "",
  latestError: "",
};

function withCookieWails(
  cookie: CookieLinkageStatusDTO = cookieOff,
  maintain: YTDLPMaintainStatusDTO = maintainOfficial,
  overrides: Record<string, unknown> = {},
): Decorator {
  return withWailsApp({
    GetYTDLPCookieLinkageStatus: () => Promise.resolve(cookie),
    GetYTDLPMaintainStatus: () => Promise.resolve(maintain),
    AcknowledgeYTDLPCookieLinkageRisk: () => Promise.resolve(),
    SetYTDLPCookieLinkageBrowser: () => Promise.resolve(),
    SetYTDLPCookieLinkageCookiesFile: () => Promise.resolve(),
    DisableYTDLPCookieLinkage: () => Promise.resolve(),
    OpenFileDialog: () => Promise.resolve(""),
    ...overrides,
  });
}

const meta = {
  title: "Components/CookieLinkageSection",
  component: CookieLinkageSection,
  tags: ["autodocs"],
  parameters: {
    layout: "padded",
  },
} satisfies Meta<typeof CookieLinkageSection>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  decorators: [withCookieWails()],
};

export const OfficialHint: Story = {
  decorators: [
    withCookieWails(cookieOff, {
      ...maintainOfficial,
      maintainDesired: false,
      effectiveOfficial: false,
    }),
  ],
};

export const BrowserEnabled: Story = {
  decorators: [
    withCookieWails({
      ...cookieOff,
      enabled: true,
      sourceKind: "browser",
      browser: "chrome",
      riskAcknowledged: true,
    }),
  ],
};

export const UnsupportedPlatform: Story = {
  decorators: [
    withCookieWails({
      ...cookieOff,
      supported: false,
      browser: "",
    }),
  ],
};

export const UnsupportedSourceKind: Story = {
  decorators: [
    withCookieWails({
      ...cookieOff,
      enabled: true,
      sourceKind: "unsupported",
      riskAcknowledged: true,
    }),
  ],
};

/** cookieActionError is set from GetStatus failure (not maintain pendingError). */
export const ConfigReadError: Story = {
  decorators: [
    withCookieWails(cookieOff, maintainOfficial, {
      GetYTDLPCookieLinkageStatus: () =>
        Promise.reject(
          new Error("cookie linkage config read: permission denied"),
        ),
    }),
  ],
};
