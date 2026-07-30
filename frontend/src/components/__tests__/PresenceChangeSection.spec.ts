import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount, flushPromises, type VueWrapper } from "@vue/test-utils";
import { ElMessage } from "element-plus";
import PresenceChangeSection from "../PresenceChangeSection.vue";

const {
  mockGetPresenceChangeSection,
  mockApplyPresenceChange,
  mockEventsOn,
  eventHandlers,
} = vi.hoisted(() => {
  const eventHandlers: Record<string, () => void> = {};
  return {
    mockGetPresenceChangeSection: vi.fn(),
    mockApplyPresenceChange: vi.fn(),
    mockEventsOn: vi.fn((event: string, cb: () => void) => {
      eventHandlers[event] = cb;
      return () => {
        delete eventHandlers[event];
      };
    }),
    eventHandlers,
  };
});

vi.mock("../../wails/runtime", () => ({
  getRuntime: () => ({ EventsOn: mockEventsOn }),
}));

vi.mock("../../wails/app", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../wails/app")>();
  return {
    ...actual,
    App: {
      ...actual.App,
      getPresenceChangeSection: mockGetPresenceChangeSection,
      applyPresenceChange: mockApplyPresenceChange,
    },
  };
});

function loggedInSection(overrides: Record<string, unknown> = {}) {
  return {
    loggedIn: true,
    status: "active",
    statusDescription: "",
    history: ["作業中"],
    ...overrides,
  };
}

function mountSection() {
  return mount(PresenceChangeSection, {
    global: {
      stubs: {
        RouterLink: { template: "<a><slot /></a>" },
        ElAutocomplete: {
          name: "ElAutocomplete",
          props: ["modelValue"],
          emits: ["update:modelValue", "select"],
          template:
            '<input data-testid="presence-change-description-input" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
        },
      },
    },
  });
}

function applyButton(wrapper: VueWrapper) {
  return wrapper.find('[data-testid="presence-change-apply"]');
}

function expectApplyDisabled(wrapper: VueWrapper, disabled: boolean) {
  const btn = applyButton(wrapper);
  if (disabled) {
    expect(btn.classes()).toContain("is-disabled");
  } else {
    expect(btn.classes()).not.toContain("is-disabled");
  }
}

describe("PresenceChangeSection", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    for (const key of Object.keys(eventHandlers)) {
      delete eventHandlers[key];
    }
    mockGetPresenceChangeSection.mockResolvedValue(loggedInSection());
    mockApplyPresenceChange.mockResolvedValue({
      status: "join me",
      statusDescription: "hello",
    });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("calls getPresenceChangeSection on mount", async () => {
    const wrapper = mountSection();
    await flushPromises();
    expect(mockGetPresenceChangeSection).toHaveBeenCalledTimes(1);
    expect(wrapper.find('[data-testid="presence-change-section"]').exists()).toBe(
      true,
    );
  });

  it("shows presence change form with four color buttons", async () => {
    const wrapper = mountSection();
    await flushPromises();
    expect(
      wrapper.find('[data-testid="presence-change-color-join-me"]').exists(),
    ).toBe(true);
    expect(
      wrapper.find('[data-testid="presence-change-color-active"]').exists(),
    ).toBe(true);
    expect(
      wrapper.find('[data-testid="presence-change-color-ask-me"]').exists(),
    ).toBe(true);
    expect(
      wrapper.find('[data-testid="presence-change-color-busy"]').exists(),
    ).toBe(true);
  });

  it("shows login required state when not logged in", async () => {
    mockGetPresenceChangeSection.mockResolvedValue({
      loggedIn: false,
      status: "",
      statusDescription: "",
      history: [],
    });
    const wrapper = mountSection();
    await flushPromises();
    expect(
      wrapper.find('[data-testid="presence-change-login-required"]').exists(),
    ).toBe(true);
    expectApplyDisabled(wrapper, true);
  });

  it("shows inline error on load failure", async () => {
    mockGetPresenceChangeSection.mockRejectedValueOnce(new Error("db down"));
    const wrapper = mountSection();
    await flushPromises();
    expect(
      wrapper.find('[data-testid="presence-change-load-error"]').exists(),
    ).toBe(true);
  });

  it("color click updates draft only without calling apply", async () => {
    const wrapper = mountSection();
    await flushPromises();
    await wrapper
      .find('[data-testid="presence-change-color-join-me"]')
      .trigger("click");
    await flushPromises();
    expect(mockApplyPresenceChange).not.toHaveBeenCalled();
    expectApplyDisabled(wrapper, false);
  });

  it("disables apply when unchanged", async () => {
    const wrapper = mountSection();
    await flushPromises();
    expectApplyDisabled(wrapper, true);
  });

  it("applies presence change and shows success message", async () => {
    const successSpy = vi
      .spyOn(ElMessage, "success")
      .mockImplementation(() => ({ close: () => {} }));
    const wrapper = mountSection();
    await flushPromises();
    await wrapper
      .find('[data-testid="presence-change-color-busy"]')
      .trigger("click");
    await flushPromises();
    await applyButton(wrapper).trigger("click");
    await flushPromises();
    expect(mockApplyPresenceChange).toHaveBeenCalledWith("busy", "");
    expect(successSpy).toHaveBeenCalledWith("ステータスを更新しました");
    successSpy.mockRestore();
  });

  it("shows error when apply fails", async () => {
    mockApplyPresenceChange.mockRejectedValueOnce(new Error("apply failed"));
    const errorSpy = vi
      .spyOn(ElMessage, "error")
      .mockImplementation(() => ({ close: () => {} }));
    const wrapper = mountSection();
    await flushPromises();
    await wrapper
      .find('[data-testid="presence-change-color-busy"]')
      .trigger("click");
    await flushPromises();
    await applyButton(wrapper).trigger("click");
    await flushPromises();
    expect(errorSpy).toHaveBeenCalledWith("apply failed");
    errorSpy.mockRestore();
  });

  it("reloads on self cache changed when not dirty", async () => {
    vi.useFakeTimers();
    mountSection();
    await flushPromises();
    mockGetPresenceChangeSection.mockClear();
    eventHandlers["identity:self-cache-changed"]?.();
    await vi.advanceTimersByTimeAsync(300);
    await flushPromises();
    expect(mockGetPresenceChangeSection).toHaveBeenCalledTimes(1);
  });

  it("skips reload when dirty", async () => {
    vi.useFakeTimers();
    const wrapper = mountSection();
    await flushPromises();
    await wrapper
      .find('[data-testid="presence-change-color-busy"]')
      .trigger("click");
    await flushPromises();
    mockGetPresenceChangeSection.mockClear();
    eventHandlers["identity:self-cache-changed"]?.();
    await vi.advanceTimersByTimeAsync(300);
    await flushPromises();
    expect(mockGetPresenceChangeSection).not.toHaveBeenCalled();
  });
});
