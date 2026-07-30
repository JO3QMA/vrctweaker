import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import DashboardView from "../DashboardView.vue";

vi.mock("../../components/DashboardLaunchBlock.vue", () => ({
  default: {
    name: "DashboardLaunchBlock",
    template: '<div data-testid="dashboard-launch-block-stub" />',
  },
}));

vi.mock("../../components/ServerStatusSection.vue", () => ({
  default: {
    name: "ServerStatusSection",
    template: '<div data-testid="server-status-section-stub" />',
  },
}));

vi.mock("../../components/PresenceChangeSection.vue", () => ({
  default: {
    name: "PresenceChangeSection",
    template: '<div data-testid="presence-change-section-stub" />',
  },
}));

describe("DashboardView", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders dashboard sections", async () => {
    const wrapper = mount(DashboardView);
    await flushPromises();
    expect(wrapper.find(".page-title").text()).toBe("ダッシュボード");
    expect(
      wrapper.find('[data-testid="dashboard-launch-block-stub"]').exists(),
    ).toBe(true);
    expect(
      wrapper.find('[data-testid="server-status-section-stub"]').exists(),
    ).toBe(true);
    expect(
      wrapper.find('[data-testid="presence-change-section-stub"]').exists(),
    ).toBe(true);
  });
});
