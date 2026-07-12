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

  it("loads maintain status", async () => {
    const wrapper = mountView();
    await flushPromises();
    expect(App.getYTDLPMaintainStatus).toHaveBeenCalled();
    expect(wrapper.text()).toContain("2025.04.01");
    expect(wrapper.find('[data-testid="ytdlp-maintain-switch"]').exists()).toBe(
      true,
    );
  });

  it("checks latest release", async () => {
    const wrapper = mountView();
    await flushPromises();
    await wrapper.get('[data-testid="ytdlp-check-latest"]').trigger("click");
    await flushPromises();
    expect(App.checkYTDLPLatestRelease).toHaveBeenCalled();
    expect(wrapper.text()).toContain("2025.05.01");
  });
});
