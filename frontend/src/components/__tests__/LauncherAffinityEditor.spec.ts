import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import LauncherAffinityEditor from "../launcher/LauncherAffinityEditor.vue";

vi.mock("../../wails/app", () => ({
  App: {
    getLogicalProcessorCount: vi.fn().mockResolvedValue(4),
  },
}));

describe("LauncherAffinityEditor", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("does not emit on mount when modelValue is empty", async () => {
    const wrapper = mount(LauncherAffinityEditor, {
      props: { modelValue: "" },
    });
    await flushPromises();
    expect(wrapper.emitted("update:modelValue")).toBeUndefined();
  });

  it("shows validation when all visible cores are off", async () => {
    const wrapper = mount(LauncherAffinityEditor, {
      props: { modelValue: "0" },
    });
    await flushPromises();
    expect(wrapper.find('[data-testid="affinity-validation"]').exists()).toBe(
      true,
    );
  });
});
