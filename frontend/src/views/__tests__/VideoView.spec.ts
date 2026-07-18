import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import ElementPlus from "element-plus";
import { createI18n } from "vue-i18n";
import VideoView from "../VideoView.vue";
import { App } from "../../wails/app";
import ja from "../../i18n/locales/ja.json";

vi.mock("../../wails/app", () => ({
  App: {
    getYTDLPMaintainStatus: vi.fn(),
    acknowledgeYTDLPToolsReplaceRisk: vi.fn(),
    setYTDLPToolsReplaceMaintain: vi.fn(),
    checkYTDLPLatestRelease: vi.fn(),
    updateOfficialYTDLPCache: vi.fn(),
    openYTDLPCacheFolder: vi.fn(),
    openYTDLPToolsFolder: vi.fn(),
    getYTDLPCookieLinkageStatus: vi.fn(),
    acknowledgeYTDLPCookieLinkageRisk: vi.fn(),
    setYTDLPCookieLinkageBrowser: vi.fn(),
    setYTDLPCookieLinkageCookiesFile: vi.fn(),
    disableYTDLPCookieLinkage: vi.fn(),
    openFileDialog: vi.fn(),
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
  latestVersion: "",
  latestTag: "",
  latestDownloadUrl: "",
  latestError: "",
};

describe("VideoView", () => {
  beforeEach(() => {
    vi.mocked(App.getYTDLPMaintainStatus).mockResolvedValue({ ...baseStatus });
    vi.mocked(App.acknowledgeYTDLPToolsReplaceRisk).mockResolvedValue(
      undefined,
    );
    vi.mocked(App.setYTDLPToolsReplaceMaintain).mockResolvedValue(undefined);
    vi.mocked(App.checkYTDLPLatestRelease).mockResolvedValue({
      ...baseStatus,
      latestVersion: "2025.05.01",
      latestTag: "2025.05.01",
      latestDownloadUrl: "https://example.com/yt-dlp.exe",
    });
    vi.mocked(App.updateOfficialYTDLPCache).mockResolvedValue({
      ...baseStatus,
      cacheVersion: "2025.05.01",
      effectiveOfficial: true,
      maintainDesired: true,
    });
    vi.mocked(App.openYTDLPCacheFolder).mockResolvedValue(undefined);
    vi.mocked(App.openYTDLPToolsFolder).mockResolvedValue(undefined);
    vi.mocked(App.getYTDLPCookieLinkageStatus).mockResolvedValue({
      supported: false,
      enabled: false,
      sourceKind: "",
      riskAcknowledged: false,
    });
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

  it("loads without paths, ON/OFF labels, or duplicate detail rows", async () => {
    const wrapper = mountView();
    await flushPromises();
    expect(App.getYTDLPMaintainStatus).toHaveBeenCalled();
    expect(wrapper.text()).toContain("VRChat 同梱版");
    expect(wrapper.text()).toContain("yt-dlp の置換");
    expect(wrapper.text()).not.toContain("C:\\Tools\\yt-dlp.exe");
    expect(wrapper.find('[data-testid="ytdlp-maintain-switch"]').exists()).toBe(
      true,
    );
    // Switch has no active/inactive text
    expect(wrapper.find(".el-switch__label").exists()).toBe(false);
    // Details collapsed by default — version hidden
    expect(
      wrapper.find('[data-testid="ytdlp-cache-version"]').isVisible(),
    ).toBe(false);
    expect(wrapper.text()).not.toContain("置き換え設定");
  });

  it("uses a 2x2 action grid", async () => {
    const wrapper = mountView();
    await flushPromises();
    const grid = wrapper.get('[data-testid="ytdlp-action-grid"]');
    expect(grid.classes()).toContain("video-actions");
    expect(grid.findAll("button").length).toBeGreaterThanOrEqual(4);
  });

  it("expands details accordion to show versions", async () => {
    const wrapper = mountView();
    await flushPromises();
    await wrapper.get('[data-testid="ytdlp-details-toggle"]').trigger("click");
    await flushPromises();
    expect(wrapper.get('[data-testid="ytdlp-cache-version"]').text()).toContain(
      "2025.04.01",
    );
  });

  it("shows friendly GitHub rate-limit error once in alert area", async () => {
    vi.mocked(App.checkYTDLPLatestRelease).mockResolvedValue({
      ...baseStatus,
      latestError:
        'github api: 403 Forbidden: {"message":"API rate limit exceeded for xxx"}',
    });
    const wrapper = mountView();
    await flushPromises();
    await wrapper.get('[data-testid="ytdlp-check-latest"]').trigger("click");
    await flushPromises();
    const banner = wrapper.get('[data-testid="ytdlp-error-banner"]');
    expect(banner.text()).toContain("GitHub の通信制限");
    expect(wrapper.text()).not.toContain("API rate limit exceeded");
    expect(wrapper.text().match(/GitHub の通信制限/g)?.length).toBe(1);
  });

  it("checks latest release", async () => {
    const wrapper = mountView();
    await flushPromises();
    await wrapper.get('[data-testid="ytdlp-check-latest"]').trigger("click");
    await flushPromises();
    expect(App.checkYTDLPLatestRelease).toHaveBeenCalled();
    await wrapper.get('[data-testid="ytdlp-details-toggle"]').trigger("click");
    await flushPromises();
    expect(
      wrapper.get('[data-testid="ytdlp-latest-version"]').text(),
    ).toContain("2025.05.01");
  });

  it("keeps latest version in details after cache update", async () => {
    vi.mocked(App.checkYTDLPLatestRelease).mockResolvedValue({
      ...baseStatus,
      latestVersion: "2025.05.01",
      latestTag: "2025.05.01",
      latestDownloadUrl: "https://example.com/yt-dlp.exe",
    });
    vi.mocked(App.updateOfficialYTDLPCache).mockResolvedValue({
      ...baseStatus,
      cacheVersion: "2025.05.01",
      latestVersion: "",
      latestTag: "",
      latestDownloadUrl: "",
    });
    const wrapper = mountView();
    await flushPromises();
    await wrapper.get('[data-testid="ytdlp-check-latest"]').trigger("click");
    await flushPromises();
    await wrapper.get('[data-testid="ytdlp-update-cache"]').trigger("click");
    await flushPromises();
    await wrapper.get('[data-testid="ytdlp-details-toggle"]').trigger("click");
    await flushPromises();
    expect(wrapper.get('[data-testid="ytdlp-cache-version"]').text()).toContain(
      "2025.05.01",
    );
    expect(
      wrapper.get('[data-testid="ytdlp-latest-version"]').text(),
    ).toContain("2025.05.01");
  });

  it("opens cache folder", async () => {
    const wrapper = mountView();
    await flushPromises();
    await wrapper
      .get('[data-testid="ytdlp-open-cache-folder"]')
      .trigger("click");
    await flushPromises();
    expect(App.openYTDLPCacheFolder).toHaveBeenCalled();
  });

  it("shows cookie linkage section when supported", async () => {
    vi.mocked(App.getYTDLPCookieLinkageStatus).mockResolvedValue({
      supported: true,
      enabled: false,
      sourceKind: "",
      riskAcknowledged: false,
      browser: "chrome",
    });
    vi.mocked(App.getYTDLPMaintainStatus).mockResolvedValue({
      ...baseStatus,
      effectiveOfficial: false,
    });
    const wrapper = mountView();
    await flushPromises();
    expect(wrapper.find('[data-testid="video-cookie-linkage"]').exists()).toBe(
      true,
    );
    expect(
      wrapper.find('[data-testid="video-cookie-official-hint"]').exists(),
    ).toBe(true);
  });

  it("hides cookie linkage section when unsupported", async () => {
    const wrapper = mountView();
    await flushPromises();
    expect(wrapper.find('[data-testid="video-cookie-linkage"]').exists()).toBe(
      false,
    );
  });
});
