import { afterEach, describe, expect, it, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import AutomationRunLogPanel from "./AutomationRunLogPanel.vue";
import type { AutomationRunLogEntryDTO } from "../../wails/app";

const entry: AutomationRunLogEntryDTO = {
  at: "2026-10-21 21:04",
  itemId: "rule-1",
  itemName: "Rule",
  eventType: "schedule.tick",
  success: false,
  actionsCompleted: 1,
  actionsTotal: 2,
  contextLabel: "ctx",
};

const originalOffset = Object.getOwnPropertyDescriptor(
  HTMLElement.prototype,
  "offsetWidth",
);
const originalObserver = globalThis.ResizeObserver;

function stubOffsetWidth(width: number) {
  Object.defineProperty(HTMLElement.prototype, "offsetWidth", {
    configurable: true,
    get() {
      return width;
    },
  });
}

function restoreLayoutGlobals() {
  if (originalOffset) {
    Object.defineProperty(HTMLElement.prototype, "offsetWidth", originalOffset);
  }
  globalThis.ResizeObserver = originalObserver;
}

describe("AutomationRunLogPanel width", () => {
  afterEach(() => {
    restoreLayoutGlobals();
    vi.unstubAllGlobals();
  });

  it("uses offsetWidth before observing so a narrow log does not start at six columns", async () => {
    stubOffsetWidth(500);
    let observed = false;
    class SilentObserver implements ResizeObserver {
      observe(): void {
        observed = true;
      }
      unobserve(): void {}
      disconnect(): void {}
    }
    vi.stubGlobal("ResizeObserver", SilentObserver);

    const wrapper = mount(AutomationRunLogPanel, {
      props: { entries: [entry] },
    });
    await flushPromises();

    expect(observed).toBe(true);
    expect(
      wrapper.find('[data-testid="automation-log-col-event"]').exists(),
    ).toBe(true);
    expect(
      wrapper.find('[data-testid="automation-log-col-progress"]').exists(),
    ).toBe(false);
  });
});
