import { describe, it, expect, vi, afterEach } from "vitest";
import { createRouter, createMemoryHistory } from "vue-router";
import { openEncounterHistoryWindow } from "../openEncounterHistoryWindow";

describe("openEncounterHistoryWindow", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    delete (window as unknown as { go?: unknown }).go;
  });

  it("uses router.push when window.open returns null", () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: "/", name: "home", component: { template: "<div/>" } },
        {
          path: "/activity/encounter-history",
          name: "encounter-history",
          component: { template: "<div/>" },
        },
      ],
    });
    const pushSpy = vi.spyOn(router, "push").mockResolvedValue(undefined);

    vi.spyOn(window, "open").mockReturnValue(null);

    openEncounterHistoryWindow(router, "user", "usr_x");

    expect(pushSpy).toHaveBeenCalledWith({
      name: "encounter-history",
      query: { kind: "user", vrcUserId: "usr_x" },
    });
  });

  it("uses router.push without window.open when Wails App bindings exist", () => {
    (
      window as unknown as { go: { main: { App: Record<string, unknown> } } }
    ).go = {
      main: { App: {} },
    };

    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: "/", name: "home", component: { template: "<div/>" } },
        {
          path: "/activity/encounter-history",
          name: "encounter-history",
          component: { template: "<div/>" },
        },
      ],
    });
    const pushSpy = vi.spyOn(router, "push").mockResolvedValue(undefined);
    const openSpy = vi.spyOn(window, "open");

    openEncounterHistoryWindow(router, "world", "wrld_abc");

    expect(openSpy).not.toHaveBeenCalled();
    expect(pushSpy).toHaveBeenCalledWith({
      name: "encounter-history",
      query: { kind: "world", worldId: "wrld_abc" },
    });
  });
});
