import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent } from "vue";
import { describe, expect, it, vi, beforeEach } from "vitest";

/* eslint-disable vue/one-component-per-file -- inline harness components per test */

const getStartupInitStatus = vi.fn();
const retryStartupInit = vi.fn();
const eventsOn = vi.fn();

vi.mock("../../wails/app", () => ({
  App: {
    getStartupInitStatus: (...args: unknown[]) => getStartupInitStatus(...args),
    retryStartupInit: (...args: unknown[]) => retryStartupInit(...args),
  },
}));

vi.mock("../../wails/runtime", () => ({
  getRuntime: () => ({
    EventsOn: (...args: unknown[]) => eventsOn(...args),
  }),
}));

import { useStartupInit } from "../useStartupInit";

describe("useStartupInit", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    getStartupInitStatus.mockResolvedValue({
      ready: false,
      loading: true,
      phase: "opening",
      summaryKey: "",
      detail: "",
      dataDir: "",
      canOpenDataFolder: false,
    });
    retryStartupInit.mockResolvedValue({
      ready: true,
      loading: false,
      phase: "",
      summaryKey: "",
      detail: "",
      dataDir: "",
      canOpenDataFolder: false,
    });
    eventsOn.mockReturnValue(() => undefined);
  });

  it("subscribes to startup events before fetching status", async () => {
    const order: string[] = [];
    eventsOn.mockImplementation(() => {
      order.push("subscribe");
      return () => undefined;
    });
    getStartupInitStatus.mockImplementation(async () => {
      order.push("fetch");
      return {
        ready: false,
        loading: false,
        phase: "",
        summaryKey: "app.startup.failed.databaseOpen",
        detail: "",
        dataDir: "",
        canOpenDataFolder: false,
      };
    });

    const Comp = defineComponent({
      setup() {
        useStartupInit();
        return () => null;
      },
    });

    mount(Comp);
    await flushPromises();

    expect(order).toEqual(["subscribe", "fetch"]);
  });

  it("shows failure when initial fetch rejects", async () => {
    getStartupInitStatus.mockRejectedValue(new Error("offline"));

    const Comp = defineComponent({
      setup() {
        const { status } = useStartupInit();
        return () => status.value.summaryKey;
      },
    });

    const wrapper = mount(Comp);
    await flushPromises();

    expect(wrapper.text()).toBe("app.startup.failed.statusFetch");
  });

  it("does not overwrite event payload with stale fetch snapshot", async () => {
    let eventHandler: ((data?: unknown) => void) | undefined;
    eventsOn.mockImplementation((_event, handler) => {
      eventHandler = handler as (data?: unknown) => void;
      return () => undefined;
    });

    let resolveFetch: (value: unknown) => void = () => undefined;
    getStartupInitStatus.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveFetch = resolve;
        }),
    );

    const Comp = defineComponent({
      setup() {
        const { status } = useStartupInit();
        return () => `${status.value.loading}:${status.value.phase}`;
      },
    });

    const wrapper = mount(Comp);
    eventHandler?.({
      ready: false,
      loading: true,
      phase: "dedupe",
      summaryKey: "",
      detail: "",
      dataDir: "",
      canOpenDataFolder: false,
    });
    resolveFetch({
      ready: false,
      loading: false,
      phase: "",
      summaryKey: "app.startup.failed.databaseOpen",
      detail: "",
      dataDir: "",
      canOpenDataFolder: false,
    });
    await flushPromises();

    expect(wrapper.text()).toBe("true:dedupe");
  });

  it("refreshes status when retry rejects", async () => {
    retryStartupInit.mockRejectedValue(new Error("retry failed"));
    getStartupInitStatus.mockResolvedValue({
      ready: false,
      loading: false,
      phase: "",
      summaryKey: "app.startup.failed.databaseOpen",
      detail: "still broken",
      dataDir: "",
      canOpenDataFolder: false,
    });

    let retryInitFn: (() => Promise<void>) | undefined;
    const Comp = defineComponent({
      setup() {
        const { status, retryInit } = useStartupInit();
        retryInitFn = retryInit;
        return () => status.value.summaryKey;
      },
    });

    const wrapper = mount(Comp);
    await flushPromises();
    await retryInitFn?.();
    await flushPromises();

    expect(getStartupInitStatus).toHaveBeenCalledTimes(2);
    expect(wrapper.text()).toBe("app.startup.failed.databaseOpen");
  });
});
