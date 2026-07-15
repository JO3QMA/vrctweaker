import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createRouter, createMemoryHistory } from "vue-router";
import { ElMessage } from "element-plus";
import DashboardLaunchBlock from "../DashboardLaunchBlock.vue";
import type { LaunchProfileDTO } from "../../wails/app";

const {
  mockGetDashboardLaunchBlock,
  mockLaunchVRChat,
  mockInstanceRejoin,
  mockEventsOn,
  triggerEncountersChanged,
} = vi.hoisted(() => {
  let encountersChangedCb: (() => void) | undefined;
  return {
    mockGetDashboardLaunchBlock: vi.fn(),
    mockLaunchVRChat: vi.fn(),
    mockInstanceRejoin: vi.fn(),
    mockEventsOn: vi.fn((_event: string, cb: () => void) => {
      encountersChangedCb = cb;
      return vi.fn();
    }),
    triggerEncountersChanged: () => encountersChangedCb?.(),
  };
});

vi.mock("../../wails/app", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../wails/app")>();
  return {
    ...actual,
    App: {
      ...actual.App,
      getDashboardLaunchBlock: mockGetDashboardLaunchBlock,
      launchVRChat: mockLaunchVRChat,
      instanceRejoin: mockInstanceRejoin,
    },
  };
});

vi.mock("../../wails/runtime", () => ({
  getRuntime: () => ({
    EventsOn: mockEventsOn,
  }),
}));

const defaultLaunchProfile: LaunchProfileDTO = {
  id: "p-default",
  name: "Default",
  arguments: "",
  isDefault: true,
};

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: "/launcher", name: "launcher", component: { template: "<div />" } },
  ],
});

function mountBlock() {
  return mount(DashboardLaunchBlock, {
    global: {
      plugins: [router],
    },
  });
}

describe("DashboardLaunchBlock", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetDashboardLaunchBlock.mockResolvedValue({
      profiles: [],
      selectedProfileId: "",
      rejoin: null,
    });
    mockLaunchVRChat.mockResolvedValue(undefined);
    mockInstanceRejoin.mockResolvedValue(undefined);
  });

  it("calls getDashboardLaunchBlock on mount", async () => {
    mountBlock();
    await flushPromises();
    expect(mockGetDashboardLaunchBlock).toHaveBeenCalledTimes(1);
  });

  it("shows launch block with rejoin button", async () => {
    mockGetDashboardLaunchBlock.mockResolvedValue({
      profiles: [defaultLaunchProfile],
      selectedProfileId: "p-default",
      rejoin: {
        playSessionId: "ps-1",
        worldDisplayName: "Test World",
      },
    });
    const wrapper = mountBlock();
    await flushPromises();
    expect(
      wrapper.find('[data-testid="dashboard-launch-block"]').exists(),
    ).toBe(true);
    expect(wrapper.find('[data-testid="launch-block-rejoin-btn"]').text()).toBe(
      "Test World に参加",
    );
  });

  it("shows launch block without rejoin button", async () => {
    mockGetDashboardLaunchBlock.mockResolvedValue({
      profiles: [defaultLaunchProfile],
      selectedProfileId: "p-default",
      rejoin: null,
    });
    const wrapper = mountBlock();
    await flushPromises();
    expect(
      wrapper.find('[data-testid="launch-block-rejoin-btn"]').exists(),
    ).toBe(false);
    expect(wrapper.find('[data-testid="launch-block-quick-btn"]').text()).toBe(
      "VRChat を起動",
    );
  });

  it("shows empty state with launcher link", async () => {
    const wrapper = mountBlock();
    await flushPromises();
    expect(
      wrapper.find('[data-testid="launch-block-empty-state"]').exists(),
    ).toBe(true);
    expect(
      wrapper
        .find('[data-testid="launch-block-launcher-link"]')
        .attributes("href"),
    ).toBe("/launcher");
    expect(
      (
        wrapper.find('[data-testid="launch-block-quick-btn"]')
          .element as HTMLButtonElement
      ).disabled,
    ).toBe(true);
  });

  it("shows inline error on load failure without toast", async () => {
    mockGetDashboardLaunchBlock.mockRejectedValueOnce(new Error("db down"));
    const errorSpy = vi.spyOn(ElMessage, "error").mockImplementation(() => ({
      close: () => {},
    }));
    const consoleSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);
    const wrapper = mountBlock();
    await flushPromises();
    expect(
      wrapper.find('[data-testid="launch-block-load-error"]').exists(),
    ).toBe(true);
    expect(errorSpy).not.toHaveBeenCalled();
    expect(consoleSpy).toHaveBeenCalledWith(
      "DashboardLaunchBlock load failed:",
      expect.any(Error),
    );
    errorSpy.mockRestore();
    consoleSpy.mockRestore();
  });

  it("shows error on quick launch failure", async () => {
    mockGetDashboardLaunchBlock.mockResolvedValue({
      profiles: [defaultLaunchProfile],
      selectedProfileId: "p-default",
      rejoin: null,
    });
    mockLaunchVRChat.mockRejectedValueOnce(new Error("launch failed"));
    const errorSpy = vi.spyOn(ElMessage, "error").mockImplementation(() => ({
      close: () => {},
    }));
    const wrapper = mountBlock();
    await flushPromises();
    await wrapper
      .find('[data-testid="launch-block-quick-btn"]')
      .trigger("click");
    await flushPromises();
    expect(mockLaunchVRChat).toHaveBeenCalledWith("p-default");
    expect(errorSpy).toHaveBeenCalledWith("launch failed");
    errorSpy.mockRestore();
  });

  it("shows error on rejoin failure and reloads block", async () => {
    mockGetDashboardLaunchBlock.mockResolvedValue({
      profiles: [defaultLaunchProfile],
      selectedProfileId: "p-default",
      rejoin: { playSessionId: "ps-1", worldDisplayName: "" },
    });
    mockInstanceRejoin.mockRejectedValueOnce(new Error("rejoin failed"));
    const errorSpy = vi.spyOn(ElMessage, "error").mockImplementation(() => ({
      close: () => {},
    }));
    const wrapper = mountBlock();
    await flushPromises();
    await wrapper
      .find('[data-testid="launch-block-rejoin-btn"]')
      .trigger("click");
    await flushPromises();
    expect(mockInstanceRejoin).toHaveBeenCalledWith("p-default", "ps-1");
    expect(errorSpy).toHaveBeenCalledWith("rejoin failed");
    expect(mockGetDashboardLaunchBlock).toHaveBeenCalledTimes(2);
    errorSpy.mockRestore();
  });

  it("refreshes on activity event", async () => {
    vi.useFakeTimers();
    mockGetDashboardLaunchBlock.mockClear();
    mountBlock();
    await flushPromises();
    expect(mockGetDashboardLaunchBlock).toHaveBeenCalledTimes(1);
    triggerEncountersChanged();
    await vi.advanceTimersByTimeAsync(400);
    expect(mockGetDashboardLaunchBlock).toHaveBeenCalledTimes(2);
    vi.useRealTimers();
  });

  it("retries refresh after in-flight load when debounced event fires", async () => {
    vi.useFakeTimers();
    let resolveFirst: (v: unknown) => void = () => {};
    mockGetDashboardLaunchBlock
      .mockReturnValueOnce(
        new Promise((resolve) => {
          resolveFirst = resolve;
        }),
      )
      .mockResolvedValueOnce({
        profiles: [defaultLaunchProfile],
        selectedProfileId: "p-default",
        rejoin: null,
      });

    mountBlock();
    await flushPromises();

    triggerEncountersChanged();
    await vi.advanceTimersByTimeAsync(400);

    resolveFirst({
      profiles: [],
      selectedProfileId: "",
      rejoin: null,
    });
    await flushPromises();

    expect(mockGetDashboardLaunchBlock).toHaveBeenCalledTimes(2);
    vi.useRealTimers();
  });

  it("skips state update after unmount", async () => {
    let resolveLoad: (v: unknown) => void = () => {};
    mockGetDashboardLaunchBlock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveLoad = resolve;
      }),
    );
    const wrapper = mountBlock();
    wrapper.unmount();
    resolveLoad({
      profiles: [defaultLaunchProfile],
      selectedProfileId: "p-default",
      rejoin: null,
    });
    await flushPromises();
    expect(
      wrapper.find('[data-testid="launch-block-quick-btn"]').exists(),
    ).toBe(false);
  });
});
