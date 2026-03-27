import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import VideoView from "../VideoView.vue";

const mockGetYTDLPBasics = vi.fn();
const mockGetYTDLPUpdateStatus = vi.fn();
const mockApplyYTDLP = vi.fn();

vi.mock("../../wails/app", () => ({
  App: {
    getYTDLPBasics: () => mockGetYTDLPBasics(),
    getYTDLPUpdateStatus: () => mockGetYTDLPUpdateStatus(),
    applyYTDLP: (u: string, t: string) => mockApplyYTDLP(u, t),
  },
}));

const router = createRouter({
  history: createWebHashHistory(),
  routes: [{ path: "/video", component: VideoView }],
});

const supportedBasics = {
  supported: true,
  targetPath: "C:\\VRChat\\VRChat\\Tools\\yt-dlp.exe",
  localVersion: "2024.01.01",
  latestVersion: "",
  latestTag: "",
  latestDownloadUrl: "",
  latestError: "",
};

describe("VideoView", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetYTDLPBasics.mockResolvedValue({ ...supportedBasics });
    mockGetYTDLPUpdateStatus.mockResolvedValue({
      ...supportedBasics,
      latestVersion: "2025.01.01",
      latestTag: "2025.01.01",
      latestDownloadUrl: "https://github.com/y/y/releases/download/ytdlp.exe",
      latestError: "",
    });
    mockApplyYTDLP.mockResolvedValue({
      ok: true,
      appliedVersion: "2025.01.01",
      message: "適用しました。",
      error: "",
    });
  });

  it("renders page title", async () => {
    await router.push("/video");
    await router.isReady();
    const wrapper = mount(VideoView, {
      global: { plugins: [router] },
    });
    await flushPromises();
    expect(wrapper.find(".page-title").text()).toBe("動画");
  });

  it("disables apply until latest is confirmed", async () => {
    await router.push("/video");
    await router.isReady();
    const wrapper = mount(VideoView, {
      global: { plugins: [router] },
    });
    await flushPromises();
    const applyBtn = wrapper.get('[data-testid="ytdlp-apply"]');
    expect(applyBtn.attributes("disabled")).toBeDefined();

    await wrapper.get('[data-testid="ytdlp-check-latest"]').trigger("click");
    await flushPromises();

    expect(mockGetYTDLPUpdateStatus).toHaveBeenCalled();
    expect(applyBtn.attributes("disabled")).toBeUndefined();
  });

  it("shows unsupported message when basics say so", async () => {
    mockGetYTDLPBasics.mockResolvedValue({
      supported: false,
      targetPath: "",
      localVersion: "",
      latestVersion: "",
      latestTag: "",
      latestDownloadUrl: "",
      latestError: "",
      unsupportedReason: "Windows のみ",
    });
    await router.push("/video");
    await router.isReady();
    const wrapper = mount(VideoView, {
      global: { plugins: [router] },
    });
    await flushPromises();
    expect(wrapper.text()).toContain("Windows のみ");
  });
});
