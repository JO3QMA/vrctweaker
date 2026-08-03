import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises, type VueWrapper } from "@vue/test-utils";
import CookieLinkageSection from "../CookieLinkageSection.vue";
import { App } from "../../wails/app";

const maintainOfficial = {
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

vi.mock("../../wails/app", () => ({
  App: {
    getYTDLPCookieLinkageStatus: vi.fn(),
    getYTDLPMaintainStatus: vi.fn(),
    acknowledgeYTDLPCookieLinkageRisk: vi.fn(),
    setYTDLPCookieLinkageBrowser: vi.fn(),
    setYTDLPCookieLinkageCookiesFile: vi.fn(),
    disableYTDLPCookieLinkage: vi.fn(),
    openFileDialog: vi.fn(),
  },
}));

function mountSection() {
  return mount(CookieLinkageSection);
}

function cookieSwitch(wrapper: VueWrapper) {
  return wrapper.get('[data-testid="video-cookie-enable"]');
}

function isSwitchOn(wrapper: VueWrapper): boolean {
  return cookieSwitch(wrapper).classes().includes("is-checked");
}

async function selectCookieSource(
  wrapper: VueWrapper,
  source: "browser" | "file",
) {
  const group = wrapper.findComponent({ name: "ElRadioGroup" });
  await group.setValue(source);
  await group.vm.$emit("change", source);
  await flushPromises();
}

describe("CookieLinkageSection source kind switch", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(App.getYTDLPMaintainStatus).mockResolvedValue(maintainOfficial);
    vi.mocked(App.acknowledgeYTDLPCookieLinkageRisk).mockResolvedValue(
      undefined,
    );
    vi.mocked(App.setYTDLPCookieLinkageBrowser).mockResolvedValue(undefined);
    vi.mocked(App.setYTDLPCookieLinkageCookiesFile).mockResolvedValue(
      undefined,
    );
    vi.mocked(App.disableYTDLPCookieLinkage).mockResolvedValue(undefined);
  });

  it("disables quietly when switching browser to file with empty path", async () => {
    vi.mocked(App.getYTDLPCookieLinkageStatus)
      .mockResolvedValueOnce({
        supported: true,
        enabled: true,
        sourceKind: "browser",
        browser: "chrome",
        riskAcknowledged: true,
        cookiesFilePath: "",
      })
      .mockResolvedValue({
        supported: true,
        enabled: false,
        sourceKind: "",
        riskAcknowledged: true,
        cookiesFilePath: "",
      });

    const wrapper = mountSection();
    await flushPromises();

    await selectCookieSource(wrapper, "file");

    expect(App.disableYTDLPCookieLinkage).toHaveBeenCalledTimes(1);
    expect(App.setYTDLPCookieLinkageCookiesFile).not.toHaveBeenCalled();
    expect(isSwitchOn(wrapper)).toBe(false);
    const fileRadio = wrapper
      .get('[data-testid="video-cookie-source"]')
      .findAll(".el-radio-button")[1]!;
    expect(fileRadio.classes()).toContain("is-active");
  });

  it("writes browser source when switching file to browser while enabled", async () => {
    vi.mocked(App.getYTDLPCookieLinkageStatus)
      .mockResolvedValueOnce({
        supported: true,
        enabled: true,
        sourceKind: "file",
        cookiesFilePath: "C:\\cookies.txt",
        riskAcknowledged: true,
        browser: "chrome",
      })
      .mockResolvedValue({
        supported: true,
        enabled: true,
        sourceKind: "browser",
        browser: "chrome",
        riskAcknowledged: true,
        cookiesFilePath: "",
      });

    const wrapper = mountSection();
    await flushPromises();

    await selectCookieSource(wrapper, "browser");

    expect(App.setYTDLPCookieLinkageBrowser).toHaveBeenCalledWith("chrome");
    expect(App.disableYTDLPCookieLinkage).not.toHaveBeenCalled();
    expect(isSwitchOn(wrapper)).toBe(true);
  });

  it("reverts to file source when browser write fails after file to browser switch", async () => {
    vi.mocked(App.getYTDLPCookieLinkageStatus).mockResolvedValue({
      supported: true,
      enabled: true,
      sourceKind: "file",
      cookiesFilePath: "C:\\cookies.txt",
      riskAcknowledged: true,
      browser: "chrome",
    });
    vi.mocked(App.setYTDLPCookieLinkageBrowser).mockRejectedValue(
      new Error("cookie linkage config read: permission denied"),
    );

    const wrapper = mountSection();
    await flushPromises();

    await selectCookieSource(wrapper, "browser");

    expect(wrapper.text()).toContain(
      "yt-dlp の設定ファイルを読めませんでした。",
    );
    expect(isSwitchOn(wrapper)).toBe(true);
    const fileRadio = wrapper
      .get('[data-testid="video-cookie-source"]')
      .findAll(".el-radio-button")[1]!;
    expect(fileRadio.classes()).toContain("is-active");
  });

  it("reverts source radio on disable failure when switching to file with empty path", async () => {
    vi.mocked(App.getYTDLPCookieLinkageStatus).mockResolvedValue({
      supported: true,
      enabled: true,
      sourceKind: "browser",
      browser: "chrome",
      riskAcknowledged: true,
      cookiesFilePath: "",
    });
    vi.mocked(App.disableYTDLPCookieLinkage).mockRejectedValue(
      new Error("cookie linkage config read: permission denied"),
    );

    const wrapper = mountSection();
    await flushPromises();

    await selectCookieSource(wrapper, "file");

    expect(wrapper.text()).toContain(
      "yt-dlp の設定ファイルを読めませんでした。",
    );
    expect(isSwitchOn(wrapper)).toBe(true);
    const browserRadio = wrapper
      .get('[data-testid="video-cookie-source"]')
      .findAll(".el-radio-button")[0]!;
    expect(browserRadio.classes()).toContain("is-active");
  });
});

describe("CookieLinkageSection unsupported platform", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(App.getYTDLPMaintainStatus).mockResolvedValue(maintainOfficial);
  });

  it("shows action error when status fetch fails on unsupported platform", async () => {
    vi.mocked(App.getYTDLPCookieLinkageStatus).mockRejectedValue(
      new Error("cookie linkage config read: permission denied"),
    );

    const wrapper = mountSection();
    await flushPromises();

    expect(wrapper.find(".cookie-unsupported-error").exists()).toBe(true);
    expect(wrapper.text()).toContain(
      "yt-dlp の設定ファイルを読めませんでした。",
    );
    expect(wrapper.find('[data-testid="video-cookie-enable"]').exists()).toBe(
      false,
    );
  });
});
