import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { ref } from "vue";
import GalleryView from "../GalleryView.vue";
import * as galleryThumbnailCache from "../galleryThumbnailCache";
import type { ScreenshotDTO, VRChatConfigDTO } from "../../wails/app";

/** jsdom ではスクロール計測が安定しないため、十分な行を可視として返す。 */
const MOCK_VIRTUAL_ROW_HEIGHT = 48;
const MOCK_VIRTUAL_ROW_COUNT = 48;

vi.mock("@tanstack/vue-virtual", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@tanstack/vue-virtual")>();
  return {
    ...actual,
    useVirtualizer: () =>
      ref({
        getVirtualItems: () =>
          Array.from({ length: MOCK_VIRTUAL_ROW_COUNT }, (_, index) => ({
            key: index,
            index,
            start: index * MOCK_VIRTUAL_ROW_HEIGHT,
            end: (index + 1) * MOCK_VIRTUAL_ROW_HEIGHT,
            size: MOCK_VIRTUAL_ROW_HEIGHT,
            lane: 0,
          })),
        getTotalSize: () => MOCK_VIRTUAL_ROW_COUNT * MOCK_VIRTUAL_ROW_HEIGHT,
        measure: () => {},
      }),
  };
});

const {
  mockScreenshots,
  mockSearchScreenshots,
  mockScanScreenshotDir,
  mockIsGalleryScanning,
  mockGetVRChatConfig,
  mockDefaultVRChatPictureFolder,
  mockJoinWorldFromScreenshot,
  mockScreenshotThumbnailDataURL,
  mockOpenScreenshotExternally,
  mockRevealScreenshotInFileManager,
} = vi.hoisted(() => ({
  mockScreenshots: vi.fn(),
  mockSearchScreenshots: vi.fn(),
  mockScanScreenshotDir: vi.fn(),
  mockIsGalleryScanning: vi.fn(),
  mockGetVRChatConfig: vi.fn(),
  mockDefaultVRChatPictureFolder: vi.fn(),
  mockJoinWorldFromScreenshot: vi.fn(),
  mockScreenshotThumbnailDataURL: vi.fn(),
  mockOpenScreenshotExternally: vi.fn(),
  mockRevealScreenshotInFileManager: vi.fn(),
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
      isGalleryScanning: mockIsGalleryScanning,
      getVRChatConfig: mockGetVRChatConfig,
      defaultVRChatPictureFolder: mockDefaultVRChatPictureFolder,
      joinWorldFromScreenshot: mockJoinWorldFromScreenshot,
      screenshotThumbnailDataURL: mockScreenshotThumbnailDataURL,
      openScreenshotExternally: mockOpenScreenshotExternally,
      revealScreenshotInFileManager: mockRevealScreenshotInFileManager,
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
  let wailsEventListeners: Record<string, (data?: unknown) => void>;

  beforeEach(() => {
    wailsEventListeners = {};
    Object.defineProperty(window, "runtime", {
      configurable: true,
      enumerable: true,
      value: {
        EventsOn: (eventName: string, cb: (data?: unknown) => void) => {
          wailsEventListeners[eventName] = cb;
          return () => {
            delete wailsEventListeners[eventName];
          };
        },
      },
    });

    host = document.createElement("div");
    host.style.width = "960px";
    host.style.height = "800px";
    host.style.display = "flex";
    host.style.flexDirection = "column";
    document.body.appendChild(host);

    vi.clearAllMocks();
    mockScreenshots.mockResolvedValue([sampleShot]);
    mockSearchScreenshots.mockResolvedValue([sampleShot]);
    mockIsGalleryScanning.mockResolvedValue(false);
    mockScanScreenshotDir.mockImplementation((_path: string) =>
      Promise.resolve(3).then((count) => {
        queueMicrotask(() => {
          wailsEventListeners["gallery:scan-done"]?.({ count });
        });
        return count;
      }),
    );
    mockGetVRChatConfig.mockResolvedValue({ ...defaultConfig });
    mockDefaultVRChatPictureFolder.mockResolvedValue("C:/Temp/Pictures/VRChat");
    mockJoinWorldFromScreenshot.mockResolvedValue(undefined);
    mockScreenshotThumbnailDataURL.mockResolvedValue(
      "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/wAARCAABAAEDAREAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAr/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCwAA8A/9k=",
    );
    mockOpenScreenshotExternally.mockResolvedValue(undefined);
    mockRevealScreenshotInFileManager.mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.useRealTimers();
    Reflect.deleteProperty(window, "runtime");
    host.remove();
  });

  it("loads all screenshots on mount via App.screenshots", async () => {
    mount(GalleryView, { attachTo: host });
    await flushPromises();
    expect(mockIsGalleryScanning).toHaveBeenCalled();
    expect(mockScreenshots).toHaveBeenCalledWith("");
    expect(mockSearchScreenshots).not.toHaveBeenCalled();
  });

  it("shows scan progress on mount when IsGalleryScanning is true", async () => {
    mockIsGalleryScanning.mockResolvedValue(true);
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    expect(wrapper.find('[data-testid="gallery-scan-progress"]').exists()).toBe(
      true,
    );
  });

  it("gallery:scan-done triggers list reload", async () => {
    mount(GalleryView, { attachTo: host });
    await flushPromises();
    mockScreenshots.mockClear();
    wailsEventListeners["gallery:scan-done"]?.({ count: 2 });
    await flushPromises();
    expect(mockScreenshots).toHaveBeenCalledWith("");
  });

  it("debounces reload when gallery:screenshots-changed fires", async () => {
    vi.useFakeTimers();
    const debounceMs = 400;
    mount(GalleryView, { attachTo: host });
    await flushPromises();
    mockScreenshots.mockClear();

    wailsEventListeners["gallery:screenshots-changed"]?.();
    expect(mockScreenshots).not.toHaveBeenCalled();

    await vi.advanceTimersByTimeAsync(debounceMs);
    await flushPromises();
    expect(mockScreenshots).toHaveBeenCalledWith("");

    mockScreenshots.mockClear();
    wailsEventListeners["gallery:screenshots-changed"]?.();
    wailsEventListeners["gallery:screenshots-changed"]?.();
    await vi.advanceTimersByTimeAsync(debounceMs);
    await flushPromises();
    expect(mockScreenshots).toHaveBeenCalledTimes(1);
  });

  it("debounced reload from gallery:screenshots-changed respects active world filter", async () => {
    vi.useFakeTimers();
    const debounceMs = 400;
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    const input = wrapper.find("[data-testid='gallery-world-filter']");
    await input.setValue("wrld_filtered");
    await input.trigger("keyup.enter");
    await flushPromises();
    mockScreenshots.mockClear();
    mockSearchScreenshots.mockClear();

    wailsEventListeners["gallery:screenshots-changed"]?.();
    await vi.advanceTimersByTimeAsync(debounceMs);
    await flushPromises();

    expect(mockSearchScreenshots).toHaveBeenCalledWith({
      worldId: "wrld_filtered",
    });
    expect(mockScreenshots).not.toHaveBeenCalled();
  });

  it("clears pending gallery:screenshots-changed debounce on unmount", async () => {
    vi.useFakeTimers();
    const debounceMs = 400;
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    mockScreenshots.mockClear();

    wailsEventListeners["gallery:screenshots-changed"]?.();
    wrapper.unmount();
    await vi.advanceTimersByTimeAsync(debounceMs);
    await flushPromises();

    expect(mockScreenshots).not.toHaveBeenCalled();
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

  it("shows determinate progress and current file when gallery:scan-progress fires during scan", async () => {
    let resolveScan!: (n: number) => void;
    mockScanScreenshotDir.mockImplementation(
      () =>
        new Promise<number>((resolve) => {
          resolveScan = (n: number) => {
            queueMicrotask(() => {
              wailsEventListeners["gallery:scan-done"]?.({ count: n });
            });
            resolve(n);
          };
        }),
    );

    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    expect(wailsEventListeners["gallery:scan-progress"]).toBeDefined();

    void wrapper.find("[data-testid='gallery-scan-folder']").trigger("click");
    await flushPromises();

    const panel = wrapper.find('[data-testid="gallery-scan-progress"]');
    expect(panel.exists()).toBe(true);

    wailsEventListeners["gallery:scan-progress"]?.({
      phase: "importing",
      current: 1,
      total: 2,
      item: "VRChat_foo.png",
    });
    await wrapper.vm.$nextTick();

    expect(panel.text()).toContain("VRChat_foo.png");
    expect(panel.text()).toContain("1 / 2");
    const prog = panel.find("progress");
    expect(prog.attributes("value")).toBe("1");
    expect(prog.attributes("max")).toBe("2");

    resolveScan(1);
    await flushPromises();
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

  it("falls back to default folder when getVRChatConfig rejects", async () => {
    mockGetVRChatConfig.mockRejectedValue(new Error("config read failed"));
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

  it("falls back to default folder when config.json is missing (Go error shape)", async () => {
    mockGetVRChatConfig.mockRejectedValue(
      new Error("config.json does not exist: no such file"),
    );
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

  it("shows detail preview image data URL when an item is selected", async () => {
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    await flushPromises();

    await wrapper.find(".grid-item").trigger("click");
    await wrapper.vm.$nextTick();

    const img = wrapper.find('[data-testid="gallery-detail-preview"]');
    expect(img.exists()).toBe(true);
    expect(img.attributes("src") ?? "").toContain("data:image/jpeg;base64,");
  });

  it("opens screenshot file via App when path button is clicked", async () => {
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    await flushPromises();

    await wrapper.find(".grid-item").trigger("click");
    await wrapper.vm.$nextTick();

    await wrapper
      .find('[data-testid="gallery-detail-open-file"]')
      .trigger("click");
    await flushPromises();

    expect(mockOpenScreenshotExternally).toHaveBeenCalledWith(sampleShot.id);
  });

  it("reveals folder via App when folder button is clicked", async () => {
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    await flushPromises();

    await wrapper.find(".grid-item").trigger("click");
    await wrapper.vm.$nextTick();

    await wrapper
      .find('[data-testid="gallery-detail-open-folder"]')
      .trigger("click");
    await flushPromises();

    expect(mockRevealScreenshotInFileManager).toHaveBeenCalledWith(
      sampleShot.id,
    );
  });

  it("renders date group headers for multiple screenshots", async () => {
    const june15: ScreenshotDTO = {
      ...sampleShot,
      id: "s-a",
      filePath: "C:/a.png",
      takenAt: "2024-06-15T10:00:00Z",
    };
    const june1: ScreenshotDTO = {
      ...sampleShot,
      id: "s-b",
      filePath: "C:/b.png",
      takenAt: "2024-06-01T10:00:00Z",
    };
    mockScreenshots.mockResolvedValue([june15, june1]);
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    await flushPromises();

    const txt = wrapper.text();
    expect(txt).toContain("2024年");
    expect(txt).toContain("2024年06月");
    expect(txt).toContain("2024年06月15日");
    expect(txt).toContain("2024年06月01日");
    expect(
      wrapper.findAll("[data-testid='gallery-group-header']").length,
    ).toBeGreaterThan(0);
  });

  it("collapsing year hides thumbnails under that year", async () => {
    const june15: ScreenshotDTO = {
      ...sampleShot,
      id: "s-a",
      filePath: "C:/a.png",
      takenAt: "2024-06-15T10:00:00Z",
    };
    const june1: ScreenshotDTO = {
      ...sampleShot,
      id: "s-b",
      filePath: "C:/b.png",
      takenAt: "2024-06-01T10:00:00Z",
    };
    mockScreenshots.mockResolvedValue([june15, june1]);
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    await flushPromises();

    expect(wrapper.findAll(".grid-item").length).toBe(2);

    const yearBtn = wrapper.find('[data-collapse-key="y:2024"]');
    expect(yearBtn.exists()).toBe(true);
    expect(yearBtn.attributes("aria-expanded")).toBe("true");

    await yearBtn.trigger("click");
    await wrapper.vm.$nextTick();

    expect(wrapper.findAll(".grid-item").length).toBe(0);
    expect(
      wrapper.find('[data-collapse-key="y:2024"]').attributes("aria-expanded"),
    ).toBe("false");
  });

  it("calls pruneThumbnailUrlMap when collapsing a year group", async () => {
    const june15: ScreenshotDTO = {
      ...sampleShot,
      id: "s-a",
      filePath: "C:/a.png",
      takenAt: "2024-06-15T10:00:00Z",
    };
    const june1: ScreenshotDTO = {
      ...sampleShot,
      id: "s-b",
      filePath: "C:/b.png",
      takenAt: "2024-06-01T10:00:00Z",
    };
    mockScreenshots.mockResolvedValue([june15, june1]);
    const pruneSpy = vi.spyOn(galleryThumbnailCache, "pruneThumbnailUrlMap");
    const wrapper = mount(GalleryView, { attachTo: host });
    await flushPromises();
    await flushPromises();
    pruneSpy.mockClear();

    await wrapper.find('[data-collapse-key="y:2024"]').trigger("click");
    await wrapper.vm.$nextTick();

    expect(pruneSpy).toHaveBeenCalled();
    pruneSpy.mockRestore();
  });

  it("debounced scroll schedules thumbnail cache prune", async () => {
    vi.useFakeTimers();
    const pruneSpy = vi.spyOn(galleryThumbnailCache, "pruneThumbnailUrlMap");
    try {
      const wrapper = mount(GalleryView, { attachTo: host });
      await flushPromises();
      await flushPromises();

      await wrapper
        .find("[data-testid='gallery-grid-scroll']")
        .trigger("scroll");
      const callsAfterScroll = pruneSpy.mock.calls.length;
      await vi.advanceTimersByTimeAsync(150);
      expect(pruneSpy.mock.calls.length).toBeGreaterThan(callsAfterScroll);
    } finally {
      pruneSpy.mockRestore();
      vi.useRealTimers();
    }
  });
});
