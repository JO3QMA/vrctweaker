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
    expect(detail.text()).toContain("(Scheduled) Database Maintenance");
  });

  it("renders maintenance phase, window, and tooltip", async () => {
    vi.setSystemTime(new Date(2026, 11, 23, 15, 0, 0));
    mockGetServerStatus.mockResolvedValue({
      fetchState: "ok",
      summary: { indicator: "none", description: "All Systems Operational" },
      components: [],
      incidents: [],
      maintenances: [
        {
          name: "Later",
          status: "scheduled",
          scheduledFor: new Date(2026, 11, 25, 1, 0).toISOString(),
          scheduledUntil: new Date(2026, 11, 25, 3, 0).toISOString(),
        },
        {
          name: "Done",
          status: "completed",
          scheduledFor: new Date(2026, 11, 20, 1, 0).toISOString(),
          scheduledUntil: new Date(2026, 11, 20, 2, 0).toISOString(),
          updates: ["finished"],
        },
        {
          id: "live",
          name: "API",
          status: "in_progress",
          scheduledFor: new Date(2026, 11, 23, 12, 34).toISOString(),
          scheduledUntil: new Date(2026, 11, 23, 23, 45).toISOString(),
          updates: ["Edge cache restart"],
        },
        {
          name: "Overnight",
          status: "scheduled",
          scheduledFor: new Date(2026, 11, 23, 23, 0).toISOString(),
          scheduledUntil: new Date(2026, 11, 24, 2, 0).toISOString(),
        },
        {
          name: "Open",
          status: "scheduled",
          scheduledFor: new Date(2026, 11, 26, 9, 5).toISOString(),
        },
      ],
    });
    const wrapper = mountSection();
    await flushPromises();
    const rows = wrapper.findAll('[data-testid="server-status-maintenance"]');
    const lines = rows.map((row) =>
      row
        .get('[data-testid="server-status-maintenance-line"]')
        .text()
        .replace(/\s+/g, " ")
        .trim(),
    );
    expect(lines).toEqual([
      "(In progress) API until 11:45 PM",
      "(Scheduled) Overnight 12/23, 11:00 PM–12/24, 2:00 AM",
      "(Scheduled) Later 12/25, 1:00 AM–3:00 AM",
      "(Scheduled) Open 12/26, 9:05 AM",
    ]);
    expect(rows[0]?.classes()).toContain("server-status-maintenance--active");
    expect(rows[1]?.classes()).toContain(
      "server-status-maintenance--scheduled",
    );
    expect(wrapper.text()).not.toContain("Done");
    expect(wrapper.text()).not.toContain("finished");
    const tip = rows[0]?.getComponent({ name: "ElTooltip" });
    expect(tip?.props("content")).toBe("API\n\nEdge cache restart");
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
