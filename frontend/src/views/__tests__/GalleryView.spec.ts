import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { ref } from "vue";
import GalleryView from "../GalleryView.vue";
import type { ScreenshotDTO, VRChatConfigDTO } from "../../wails/app";

vi.mock("@tanstack/vue-virtual", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@tanstack/vue-virtual")>();
  return {
    ...actual,
    useVirtualizer: () =>
      ref({
        getVirtualItems: () => [
          { key: 0, index: 0, start: 0, end: 120, size: 120, lane: 0 },
        ],
        getTotalSize: () => 120,
        measure: () => {},
      }),
  };
});

const {
  mockScreenshots,
  mockSearchScreenshots,
  mockScanScreenshotDir,
  mockGetVRChatConfig,
  mockDefaultVRChatPictureFolder,
  mockJoinWorldFromScreenshot,
  mockScreenshotThumbnailDataURL,
} = vi.hoisted(() => ({
  mockScreenshots: vi.fn(),
  mockSearchScreenshots: vi.fn(),
  mockScanScreenshotDir: vi.fn(),
  mockGetVRChatConfig: vi.fn(),
  mockDefaultVRChatPictureFolder: vi.fn(),
  mockJoinWorldFromScreenshot: vi.fn(),
  mockScreenshotThumbnailDataURL: vi.fn(),
}));

vi.mock("../../wails/app", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../wails/app")>();
  return {
    ...actual,
    App: {
      ...actual.App,
      screenshots: mockScreenshots,
      searchScreenshots: mockSearchScreenshots,
      scanScreenshotDir: mockScanScreenshotDir,
      getVRChatConfig: mockGetVRChatConfig,
      defaultVRChatPictureFolder: mockDefaultVRChatPictureFolder,
      joinWorldFromScreenshot: mockJoinWorldFromScreenshot,
      screenshotThumbnailDataURL: mockScreenshotThumbnailDataURL,
    },
  };
});

const sampleShot: ScreenshotDTO = {
  id: "s1",
  filePath: "C:/VRChat/2024/shot.png",
  worldId: "wrld_abc",
  worldName: "Test World",
  takenAt: "2024-01-15T12:00:00Z",
  fileSizeBytes: 12345,
};

const defaultConfig: VRChatConfigDTO = {
  cameraResWidth: 1920,
  cameraResHeight: 1080,
  screenshotResWidth: 1920,
  screenshotResHeight: 1080,
  pictureOutputFolder: "C:/Pictures/VRChat",
  pictureOutputSplitByDate: true,
  fpvSteadycamFov: 90,
  cacheDirectory: "",
  cacheSize: 0,
  cacheExpiryDelay: 0,
  disableRichPresence: null,
};

describe("GalleryView", () => {
  let host: HTMLDivElement;

  beforeEach(() => {
    host = document.createElement("div");
    host.style.width = "960px";
    host.style.height = "800px";
    host.style.display = "flex";
    host.style.flexDirection = "column";
    document.body.appendChild(host);

    vi.clearAllMocks();
    mockScreenshots.mockResolvedValue([sampleShot]);
    mockSearchScreenshots.mockResolvedValue([sampleShot]);
    mockScanScreenshotDir.mockResolvedValue(3);
    mockGetVRChatConfig.mockResolvedValue({ ...defaultConfig });
    mockDefaultVRChatPictureFolder.mockResolvedValue("C:/Temp/Pictures/VRChat");
    mockJoinWorldFromScreenshot.mockResolvedValue(undefined);
    mockScreenshotThumbnailDataURL.mockResolvedValue(
      "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/wAARCAABAAEDAREAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAr/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCwAA8A/9k=",
    );
  });

  afterEach(() => {
    vi.useRealTimers();
    host.remove();
  });

  it("loads all screenshots on mount via App.screenshots", async () => {
    mount(GalleryView, { attachTo: host });
    await flushPromises();
    expect(mockScreenshots).toHaveBeenCalledWith("");
    expect(mockSearchScreenshots).not.toHaveBeenCalled();
  });

  it("fetches thumbnail data URLs via App.screenshotThumbnailDataURL", async () => {
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    await flushPromises();

    expect(mockScreenshotThumbnailDataURL).toHaveBeenCalledWith(sampleShot.id);
    const img = wrapper.find(".thumbnail");
    expect(img.attributes("src") ?? "").toContain("data:image/jpeg;base64,");
  });

  it("renders virtualized grid scroll container when list is non-empty", async () => {
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    await flushPromises();
    expect(wrapper.find("[data-testid='gallery-grid-scroll']").exists()).toBe(
      true,
    );
  });

  it("searches when World ID filter is submitted with Enter", async () => {
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    mockScreenshots.mockClear();

    const input = wrapper.find("[data-testid='gallery-world-filter']");
    await input.setValue("wrld_test");
    await input.trigger("keyup.enter");
    await flushPromises();

    expect(mockSearchScreenshots).toHaveBeenCalledWith({
      worldId: "wrld_test",
    });
  });

  it("debounces realtime filter and calls searchScreenshots", async () => {
    vi.useFakeTimers();
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    mockScreenshots.mockClear();
    mockSearchScreenshots.mockClear();

    const input = wrapper.find("[data-testid='gallery-world-filter']");
    await input.setValue("wrld_slow");
    await wrapper.vm.$nextTick();

    expect(mockSearchScreenshots).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(450);
    await flushPromises();

    expect(mockSearchScreenshots).toHaveBeenCalledWith({
      worldId: "wrld_slow",
    });
  });

  it("disables Scan Folder while getVRChatConfig is pending", async () => {
    let resolveConfig!: (value: VRChatConfigDTO) => void;
    mockGetVRChatConfig.mockImplementation(
      () =>
        new Promise<VRChatConfigDTO>((resolve) => {
          resolveConfig = resolve;
        }),
    );
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();

    void wrapper.find("[data-testid='gallery-scan-folder']").trigger("click");
    await wrapper.vm.$nextTick();

    const btn = wrapper.find("[data-testid='gallery-scan-folder']");
    expect(btn.attributes("disabled")).toBeDefined();

    resolveConfig({ ...defaultConfig });
    await flushPromises();
    expect(mockScanScreenshotDir).toHaveBeenCalledWith("C:/Pictures/VRChat");
  });

  it("Scan Folder loads config, scans directory, then refreshes list", async () => {
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    mockScreenshots.mockClear();

    await wrapper.find("[data-testid='gallery-scan-folder']").trigger("click");
    await flushPromises();

    expect(mockGetVRChatConfig).toHaveBeenCalled();
    expect(mockScanScreenshotDir).toHaveBeenCalledWith("C:/Pictures/VRChat");
    expect(mockScreenshots.mock.calls.length).toBeGreaterThanOrEqual(1);
  });

  it("uses default Pictures/VRChat when picture output folder is empty", async () => {
    mockGetVRChatConfig.mockResolvedValue({
      ...defaultConfig,
      pictureOutputFolder: "",
    });
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    mockScreenshots.mockClear();

    await wrapper.find("[data-testid='gallery-scan-folder']").trigger("click");
    await flushPromises();

    expect(mockDefaultVRChatPictureFolder).toHaveBeenCalled();
    expect(mockScanScreenshotDir).toHaveBeenCalledWith(
      "C:/Temp/Pictures/VRChat",
    );
    expect(mockScreenshots.mock.calls.length).toBeGreaterThanOrEqual(1);
  });

  it("shows scan error when scanScreenshotDir rejects", async () => {
    mockScanScreenshotDir.mockRejectedValue(new Error("directory scan failed"));
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    mockScreenshots.mockClear();

    await wrapper.find("[data-testid='gallery-scan-folder']").trigger("click");
    await flushPromises();

    expect(mockScreenshots).not.toHaveBeenCalled();
    expect(wrapper.find('[role="status"]').text()).toContain(
      "directory scan failed",
    );
    expect(wrapper.find('[role="alert"]').exists()).toBe(false);
  });

  it("shows scan error when getVRChatConfig rejects", async () => {
    mockGetVRChatConfig.mockRejectedValue(new Error("config read failed"));
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();

    await wrapper.find("[data-testid='gallery-scan-folder']").trigger("click");
    await flushPromises();

    expect(mockScanScreenshotDir).not.toHaveBeenCalled();
    expect(wrapper.find('[role="status"]').text()).toContain(
      "config read failed",
    );
  });

  it("shows scan error when default folder path cannot be resolved", async () => {
    mockGetVRChatConfig.mockResolvedValue({
      ...defaultConfig,
      pictureOutputFolder: "",
    });
    mockDefaultVRChatPictureFolder.mockResolvedValue("  ");
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();

    await wrapper.find("[data-testid='gallery-scan-folder']").trigger("click");
    await flushPromises();

    expect(mockScanScreenshotDir).not.toHaveBeenCalled();
    expect(wrapper.text()).toMatch(/ピクチャ|マイ ピクチャ|解決できません/);
  });

  it("shows file basename in detail when an item is selected", async () => {
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();

    await wrapper.find(".grid-item").trigger("click");
    await wrapper.vm.$nextTick();

    expect(wrapper.text()).toContain("shot.png");
    expect(wrapper.text()).toContain("ファイルサイズ");
    expect(wrapper.text()).toMatch(/12(\.0)? KB/);
  });
});
