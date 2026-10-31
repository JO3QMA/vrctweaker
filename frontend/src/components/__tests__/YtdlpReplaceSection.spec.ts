import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import ElementPlus from "element-plus";
import { createI18n } from "vue-i18n";
import YtdlpReplaceSection from "../YtdlpReplaceSection.vue";
import { App } from "../../wails/app";
import ja from "../../i18n/locales/ja.json";

vi.mock("../../wails/app", () => ({
  App: {
    getYTDLPMaintainStatus: vi.fn(),
    refreshYTDLPLatestRelease: vi.fn(),
    acknowledgeYTDLPToolsReplaceRisk: vi.fn(),
    setYTDLPToolsReplaceMaintain: vi.fn(),
    checkAndUpdateOfficialYTDLPCache: vi.fn(),
    openYTDLPCacheFolder: vi.fn(),
    openYTDLPToolsFolder: vi.fn(),
  },
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
  latestCheckedAt: "2026-10-31T00:14:00.000Z",
};

describe("YtdlpReplaceSection", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(App.getYTDLPMaintainStatus).mockResolvedValue({ ...baseStatus });
    vi.mocked(App.refreshYTDLPLatestRelease).mockResolvedValue({
      ...baseStatus,
    });
    vi.mocked(App.acknowledgeYTDLPToolsReplaceRisk).mockResolvedValue(
      undefined,
    );
    vi.mocked(App.setYTDLPToolsReplaceMaintain).mockResolvedValue(undefined);
    vi.mocked(App.checkAndUpdateOfficialYTDLPCache).mockResolvedValue({
      status: { ...baseStatus, cacheVersion: "2025.05.01" },
      outcome: "updated",
      previousVersion: "2025.04.01",
      newVersion: "2025.05.01",
      detailError: "",
    });
    vi.mocked(App.openYTDLPCacheFolder).mockResolvedValue(undefined);
    vi.mocked(App.openYTDLPToolsFolder).mockResolvedValue(undefined);
  });

  function mountSection() {
    const i18n = createI18n({
      legacy: false,
      locale: "ja",
      messages: { ja },
    });
    return mount(YtdlpReplaceSection, {
      global: { plugins: [ElementPlus, i18n] },
    });
  }

  it("shows single primary update button and tertiary folder buttons", async () => {
    const wrapper = mountSection();
    await flushPromises();
    expect(wrapper.find('[data-testid="ytdlp-check-latest"]').exists()).toBe(
      false,
    );
    expect(wrapper.get('[data-testid="ytdlp-check-update"]').text()).toContain(
      "yt-dlp を更新",
    );
    expect(wrapper.find('[data-testid="ytdlp-folder-row"]').exists()).toBe(
      true,
    );
  });

  it("shows update-available hint when GitHub version is newer", async () => {
    vi.mocked(App.getYTDLPMaintainStatus).mockResolvedValue({
      ...baseStatus,
      cacheVersion: "2025.04.01",
      latestVersion: "2025.05.01",
    });
    vi.mocked(App.refreshYTDLPLatestRelease).mockResolvedValue({
      ...baseStatus,
      latestVersion: "2025.05.01",
    });
    const wrapper = mountSection();
    await flushPromises();
    expect(
      wrapper.get('[data-testid="ytdlp-update-available-hint"]').text(),
    ).toContain("2025.05.01");
  });

  it("calls check-and-update on primary button", async () => {
    const wrapper = mountSection();
    await flushPromises();
    await wrapper.get('[data-testid="ytdlp-check-update"]').trigger("click");
    await flushPromises();
    expect(App.checkAndUpdateOfficialYTDLPCache).toHaveBeenCalled();
    expect(wrapper.get('[data-testid="ytdlp-result-line"]').text()).toContain(
      "2025.04.01",
    );
    expect(wrapper.get('[data-testid="ytdlp-result-line"]').text()).toContain(
      "2025.05.01",
    );
  });

  it("shows secondary result text when already up to date", async () => {
    vi.mocked(App.checkAndUpdateOfficialYTDLPCache).mockResolvedValue({
      status: { ...baseStatus },
      outcome: "unchanged",
      previousVersion: "2025.04.01",
      newVersion: "2025.04.01",
      detailError: "",
    });
    const wrapper = mountSection();
    await flushPromises();
    await wrapper.get('[data-testid="ytdlp-check-update"]').trigger("click");
    await flushPromises();
    const line = wrapper.get('[data-testid="ytdlp-result-line"]');
    expect(line.text()).toContain("最新版です");
    expect(line.classes()).toContain("flash-secondary");
  });

  it("shows friendly error on result line and raw error in details", async () => {
    vi.mocked(App.checkAndUpdateOfficialYTDLPCache).mockResolvedValue({
      status: { ...baseStatus },
      outcome: "error",
      previousVersion: "2025.04.01",
      newVersion: "",
      detailError:
        'github api: 403 Forbidden: {"message":"API rate limit exceeded"}',
    });
    const wrapper = mountSection();
    await flushPromises();
    await wrapper.get('[data-testid="ytdlp-check-update"]').trigger("click");
    await flushPromises();
    expect(wrapper.get('[data-testid="ytdlp-result-line"]').text()).toContain(
      "GitHub の通信制限",
    );
    await wrapper.get('[data-testid="ytdlp-details-toggle"]').trigger("click");
    await flushPromises();
    expect(wrapper.get('[data-testid="ytdlp-detail-error"]').text()).toContain(
      "API rate limit exceeded",
    );
  });

  it("refreshes GitHub latest in background on mount", async () => {
    mountSection();
    await flushPromises();
    expect(App.refreshYTDLPLatestRelease).toHaveBeenCalledWith(false);
  });

  it("expands details with last check time", async () => {
    const wrapper = mountSection();
    await flushPromises();
    await wrapper.get('[data-testid="ytdlp-details-toggle"]').trigger("click");
    await flushPromises();
    expect(wrapper.get('[data-testid="ytdlp-last-check"]').text()).not.toBe(
      "未確認",
    );
  });
});
