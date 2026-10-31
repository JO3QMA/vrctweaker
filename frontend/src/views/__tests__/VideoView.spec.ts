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
    expect(card.text()).toContain("yt-dlp(実験的機能)");
    expect(card.text()).toContain("yt-dlpの置き換え");
    expect(card.text()).toContain("Cookieを利用する");
    expect(card.find('[data-testid="ytdlp-replace-section"]').exists()).toBe(
      true,
    );
    expect(card.find('[data-testid="video-cookie-linkage"]').exists()).toBe(
      true,
    );
    // History card + one experimental-features card (cookie inside, no nested card)
    expect(wrapper.findAll(".el-card").length).toBe(2);
    expect(wrapper.get("#ytdlp-replace-heading").text()).toBe(
      "yt-dlpの置き換え",
    );
    expect(wrapper.get("#ytdlp-cookie-heading").text()).toBe(
      "Cookieを利用する",
    );
    // No switch-side feature name label (hint may still mention 置き換え)
    expect(wrapper.find(".switch-label").exists()).toBe(false);
  });

  it("loads without paths in the replace section", async () => {
    const wrapper = mountView();
    await flushPromises();
    expect(wrapper.text()).toContain("VRChat同梱版");
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

  it("shows playback history rows with outcome labels", async () => {
    mockVideoPlaybackHistory.mockResolvedValue([
      {
        id: "1",
        attemptedAt: "2026-03-18T12:00:00.000Z",
        url: "https://example.com/open",
        outcome: "",
        failureReason: "",
        worldDisplayName: "Example World",
      },
      {
        id: "2",
        attemptedAt: "2026-03-18T11:00:00.000Z",
        url: "https://example.com/ok",
        outcome: "success",
        failureReason: "",
        worldDisplayName: "",
      },
      {
        id: "3",
        attemptedAt: "2026-03-18T10:00:00.000Z",
        url: "https://example.com/fail",
        outcome: "failure",
        failureReason: "[youtube] id: format missing",
        worldDisplayName: "",
      },
    ]);
    const wrapper = mountView();
    await flushPromises();
    expect(wrapper.find('[data-testid="video-history-table"]').exists()).toBe(
      true,
    );
    expect(wrapper.text()).toContain("https://example.com/open");
    expect(wrapper.text()).toContain("未解決");
    expect(wrapper.text()).toContain("成功");
    expect(wrapper.text()).toContain("失敗");
    expect(wrapper.text()).toContain("format missing");
    expect(wrapper.text()).toContain("Example World");
  });

  it("shows in-card error when history fetch fails", async () => {
    mockVideoPlaybackHistory.mockRejectedValue(new Error("db down"));
    const wrapper = mountView();
    await flushPromises();
    const err = wrapper.get('[data-testid="video-history-fetch-error"]');
    expect(err.text()).toContain("再生履歴を取得できませんでした");
    expect(wrapper.find('[data-testid="video-history-table"]').exists()).toBe(
      false,
    );
    expect(wrapper.text()).not.toContain("db down");
  });

  it("copies attempt URL from history row", async () => {
    mockVideoPlaybackHistory.mockResolvedValue([
      {
        id: "copy-1",
        attemptedAt: "2026-03-18T12:00:00.000Z",
        url: "https://example.com/copy-me",
        outcome: "success",
        failureReason: "",
      },
    ]);
    const wrapper = mountView();
    await flushPromises();
    const copyBtn = wrapper.find('[data-testid="video-history-copy-url"]');
    expect(copyBtn.attributes("aria-label")).toBe("動画URLをコピー");
    await copyBtn.trigger("click");
    await flushPromises();
    expect(mockCopyDisplayName).toHaveBeenCalledWith(
      "https://example.com/copy-me",
    );
    expect(
      wrapper.get('[data-testid="video-history-copy-ok"]').text(),
    ).toContain("URLをコピーしました");
  });

  it("clears copy flash after a short delay", async () => {
    vi.useFakeTimers();
    mockVideoPlaybackHistory.mockResolvedValue([
      {
        id: "copy-flash",
        attemptedAt: "2026-03-18T12:00:00.000Z",
        url: "https://example.com/flash",
        outcome: "success",
        failureReason: "",
      },
    ]);
    const wrapper = mountView();
    await flushPromises();
    await wrapper
      .find('[data-testid="video-history-copy-url"]')
      .trigger("click");
    await flushPromises();
    expect(wrapper.find('[data-testid="video-history-copy-ok"]').exists()).toBe(
      true,
    );
    await vi.advanceTimersByTimeAsync(2000);
    expect(wrapper.find('[data-testid="video-history-copy-ok"]').exists()).toBe(
      false,
    );
    vi.useRealTimers();
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
