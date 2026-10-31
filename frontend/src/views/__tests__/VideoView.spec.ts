import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import ElementPlus from "element-plus";
import { createI18n } from "vue-i18n";
import VideoView from "../VideoView.vue";
import { App } from "../../wails/app";
import ja from "../../i18n/locales/ja.json";

const { mockVideoPlaybackHistory, mockCopyDisplayName } = vi.hoisted(() => ({
  mockVideoPlaybackHistory: vi.fn().mockResolvedValue([]),
  mockCopyDisplayName: vi.fn().mockResolvedValue(undefined),
}));

const runtimeHooks = vi.hoisted(() => ({
  videoPlaybackChangedHandler: null as (() => void) | null,
}));

vi.mock("../../utils/vrcUserCacheDisplay", () => ({
  copyDisplayName: mockCopyDisplayName,
}));

vi.mock("../../wails/app", () => ({
  App: {
    getYTDLPMaintainStatus: vi.fn(),
    refreshYTDLPLatestRelease: vi.fn(),
    acknowledgeYTDLPToolsReplaceRisk: vi.fn(),
    setYTDLPToolsReplaceMaintain: vi.fn(),
    checkAndUpdateOfficialYTDLPCache: vi.fn(),
    openYTDLPCacheFolder: vi.fn(),
    openYTDLPToolsFolder: vi.fn(),
    getYTDLPCookieLinkageStatus: vi.fn(),
    acknowledgeYTDLPCookieLinkageRisk: vi.fn(),
    setYTDLPCookieLinkageBrowser: vi.fn(),
    setYTDLPCookieLinkageCookiesFile: vi.fn(),
    disableYTDLPCookieLinkage: vi.fn(),
    openFileDialog: vi.fn(),
    videoPlaybackHistory: mockVideoPlaybackHistory,
    getLogRetentionDays: vi.fn(),
  },
}));

vi.mock("../../wails/runtime", () => ({
  getRuntime: () => ({
    EventsOn: (event: string, handler: () => void) => {
      if (event === "activity:video-playback-changed") {
        runtimeHooks.videoPlaybackChangedHandler = handler;
      }
      return () => {};
    },
  }),
}));

const baseStatus = {
  supported: true,
  unsupportedReason: "",
  maintainDesired: false,
  riskAcknowledged: true,
  effectiveOfficial: false,
  cachePresent: true,
  cacheVersion: "2025.04.01",
  toolsPath: "C:\\Tools\\yt-dlp.exe",
  cachePath: "C:\\cache\\yt-dlp.exe",
  pendingError: "",
  latestVersion: "2025.04.01",
  latestTag: "2025.04.01",
  latestDownloadUrl: "https://example.com/yt-dlp.exe",
  latestError: "",
  latestCheckedAt: "",
};

describe("VideoView", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    runtimeHooks.videoPlaybackChangedHandler = null;
    vi.mocked(App.getYTDLPMaintainStatus).mockResolvedValue({ ...baseStatus });
    vi.mocked(App.refreshYTDLPLatestRelease).mockResolvedValue({
      ...baseStatus,
    });
    vi.mocked(App.checkAndUpdateOfficialYTDLPCache).mockResolvedValue({
      status: baseStatus,
      outcome: "unchanged",
      previousVersion: "2025.04.01",
      newVersion: "2025.04.01",
      detailError: "",
    });
    vi.mocked(App.getYTDLPCookieLinkageStatus).mockResolvedValue({
      supported: false,
      enabled: false,
      sourceKind: "",
      riskAcknowledged: false,
    });
    vi.mocked(App.videoPlaybackHistory).mockResolvedValue([]);
    vi.mocked(App.getLogRetentionDays).mockResolvedValue(30);
    mockVideoPlaybackHistory.mockResolvedValue([]);
  });

  function mountView() {
    const i18n = createI18n({
      legacy: false,
      locale: "ja",
      messages: { ja },
    });
    return mount(VideoView, {
      global: { plugins: [ElementPlus, i18n] },
    });
  }

  it("groups replace and cookie under one experimental features card", async () => {
    vi.mocked(App.getYTDLPCookieLinkageStatus).mockResolvedValue({
      supported: true,
      enabled: false,
      sourceKind: "",
      riskAcknowledged: false,
      browser: "chrome",
    });
    const wrapper = mountView();
    await flushPromises();
    const card = wrapper.get('[data-testid="ytdlp-experimental-features"]');
    expect(card.text()).toContain("yt-dlp (実験的機能)");
    expect(card.text()).toContain("yt-dlp の置換");
    expect(card.find('[data-testid="ytdlp-replace-section"]').exists()).toBe(
      true,
    );
  });

  it("loads without paths in the replace section", async () => {
    const wrapper = mountView();
    await flushPromises();
    expect(wrapper.text()).toContain("VRChat 同梱版");
    expect(wrapper.text()).not.toContain("C:\\Tools\\yt-dlp.exe");
    expect(wrapper.find('[data-testid="ytdlp-maintain-switch"]').exists()).toBe(
      true,
    );
  });

  it("shows cookie linkage section when supported", async () => {
    vi.mocked(App.getYTDLPCookieLinkageStatus).mockResolvedValue({
      supported: true,
      enabled: false,
      sourceKind: "",
      riskAcknowledged: false,
      browser: "chrome",
    });
    const wrapper = mountView();
    await flushPromises();
    expect(wrapper.find('[data-testid="video-cookie-linkage"]').exists()).toBe(
      true,
    );
  });

  it("places playback history card above maintain section", async () => {
    const wrapper = mountView();
    await flushPromises();
    const history = wrapper.get('[data-testid="video-playback-history"]');
    const maintainSwitch = wrapper.get('[data-testid="ytdlp-maintain-switch"]');
    expect(
      history.element.compareDocumentPosition(maintainSwitch.element) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
  });

  it("debounces history refresh on activity:video-playback-changed", async () => {
    vi.useFakeTimers();
    await mountView();
    await flushPromises();
    mockVideoPlaybackHistory.mockClear();

    runtimeHooks.videoPlaybackChangedHandler?.();
    runtimeHooks.videoPlaybackChangedHandler?.();
    expect(mockVideoPlaybackHistory).not.toHaveBeenCalled();

    await vi.advanceTimersByTimeAsync(400);
    expect(mockVideoPlaybackHistory).toHaveBeenCalledTimes(1);
    vi.useRealTimers();
  });
});
