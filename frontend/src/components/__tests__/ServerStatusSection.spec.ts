import { mount, flushPromises } from "@vue/test-utils";
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { createI18n } from "vue-i18n";
import ServerStatusSection from "../ServerStatusSection.vue";
import en from "../../i18n/locales/en.json";

const mockGetServerStatus = vi.fn();

vi.mock("../../wails/app", () => ({
  App: {
    getServerStatus: (...args: unknown[]) => mockGetServerStatus(...args),
  },
}));

function mountSection() {
  const i18n = createI18n({
    legacy: false,
    locale: "en",
    messages: { en },
  });
  return mount(ServerStatusSection, {
    global: { plugins: [i18n] },
  });
}

describe("ServerStatusSection", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    mockGetServerStatus.mockReset();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("shows loading before first fetch completes", () => {
    mockGetServerStatus.mockReturnValue(new Promise(() => {}));
    const wrapper = mountSection();
    expect(wrapper.find('[data-testid="server-status-loading"]').exists()).toBe(
      true,
    );
    expect(wrapper.text()).not.toContain("Could not load server status.");
  });

  it("shows server status fetch failure", async () => {
    mockGetServerStatus.mockResolvedValue({
      fetchState: "unavailable",
      summary: { indicator: "", description: "" },
      components: [],
      incidents: [],
      maintenances: [],
    });
    const wrapper = mountSection();
    await flushPromises();
    expect(wrapper.text()).toContain("Could not load server status.");
    expect(wrapper.find('[data-testid="server-status-detail"]').exists()).toBe(
      false,
    );
  });

  it("shows partial server status detail failure", async () => {
    mockGetServerStatus.mockResolvedValue({
      fetchState: "partial",
      summary: { indicator: "none", description: "All Systems Operational" },
      components: [],
      incidents: [],
      maintenances: [],
    });
    const wrapper = mountSection();
    await flushPromises();
    expect(wrapper.text()).toContain("All systems operational");
    expect(
      wrapper.find('[data-testid="server-status-detail-unavailable"]').exists(),
    ).toBe(true);
  });

  it("hides server status detail when operational", async () => {
    mockGetServerStatus.mockResolvedValue({
      fetchState: "ok",
      summary: { indicator: "none", description: "All Systems Operational" },
      components: [],
      incidents: [],
      maintenances: [],
    });
    const wrapper = mountSection();
    await flushPromises();
    expect(wrapper.find('[data-testid="server-status-detail"]').exists()).toBe(
      false,
    );
  });

  it("shows non-operational components only", async () => {
    mockGetServerStatus.mockResolvedValue({
      fetchState: "ok",
      summary: {
        indicator: "maintenance",
        description: "Service Under Maintenance",
      },
      components: [
        { name: "Authentication / Login", status: "under_maintenance" },
      ],
      incidents: [{ name: "API Degraded" }],
      maintenances: [{ name: "Database Maintenance" }],
    });
    const wrapper = mountSection();
    await flushPromises();
    const detail = wrapper.find('[data-testid="server-status-detail"]');
    expect(detail.exists()).toBe(true);
    expect(detail.text()).toContain("Authentication / Login");
    expect(detail.text()).toContain("Under maintenance");
    expect(detail.text()).toContain("API Degraded");
    expect(detail.text()).toContain("Database Maintenance");
  });

  it("keeps prior state when Wails IPC fails after initial load", async () => {
    mockGetServerStatus.mockResolvedValueOnce({
      fetchState: "ok",
      summary: { indicator: "none", description: "" },
      components: [],
      incidents: [],
      maintenances: [],
    });
    mockGetServerStatus.mockRejectedValueOnce(new Error("ipc failed"));

    const wrapper = mountSection();
    await flushPromises();
    expect(wrapper.text()).toContain("All systems operational");

    await vi.advanceTimersByTimeAsync(5 * 60 * 1000);
    await flushPromises();
    expect(wrapper.text()).toContain("All systems operational");
    expect(wrapper.text()).not.toContain("Could not load server status.");
  });

  it("shows unavailable when Wails IPC fails on initial load", async () => {
    mockGetServerStatus.mockRejectedValue(new Error("ipc failed"));
    const wrapper = mountSection();
    await flushPromises();
    expect(wrapper.text()).toContain("Could not load server status.");
  });

  it("does not update server status after unmount", async () => {
    let resolveFetch!: (value: unknown) => void;
    mockGetServerStatus.mockReturnValue(
      new Promise((resolve) => {
        resolveFetch = resolve;
      }),
    );
    const wrapper = mountSection();
    wrapper.unmount();
    resolveFetch({
      fetchState: "ok",
      summary: { indicator: "critical", description: "Major outage" },
      components: [],
      incidents: [],
      maintenances: [],
    });
    await flushPromises();
    expect(mockGetServerStatus).toHaveBeenCalledTimes(1);
  });

  it("skips overlapping server status poll", async () => {
    let resolveFirst!: (value: unknown) => void;
    mockGetServerStatus.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveFirst = resolve;
        }),
    );
    mockGetServerStatus.mockResolvedValue({
      fetchState: "ok",
      summary: { indicator: "none", description: "" },
      components: [],
      incidents: [],
      maintenances: [],
    });

    mountSection();
    expect(mockGetServerStatus).toHaveBeenCalledTimes(1);

    await vi.advanceTimersByTimeAsync(5 * 60 * 1000);
    expect(mockGetServerStatus).toHaveBeenCalledTimes(1);

    resolveFirst({
      fetchState: "ok",
      summary: { indicator: "none", description: "" },
      components: [],
      incidents: [],
      maintenances: [],
    });
    await flushPromises();

    await vi.advanceTimersByTimeAsync(5 * 60 * 1000);
    expect(mockGetServerStatus).toHaveBeenCalledTimes(2);
  });
});
