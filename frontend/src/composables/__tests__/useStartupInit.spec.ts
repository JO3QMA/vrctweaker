import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent } from "vue";
import { describe, expect, it, vi, beforeEach } from "vitest";

const getStartupInitStatus = vi.fn();
const eventsOn = vi.fn();

vi.mock("../../wails/app", () => ({
  App: {
    getStartupInitStatus: (...args: unknown[]) => getStartupInitStatus(...args),
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
});
