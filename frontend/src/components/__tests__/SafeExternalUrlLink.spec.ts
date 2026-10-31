import { describe, expect, it, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import { createI18n } from "vue-i18n";
import ja from "../../i18n/locales/ja.json";
import SafeExternalUrlLink from "../SafeExternalUrlLink.vue";
import * as openExternal from "@/utils/openExternalHttpUrl";
import { showToast } from "@/utils/showToast";

vi.mock("@/utils/openExternalHttpUrl", () => ({
  openValidatedExternalHttpUrl: vi.fn(),
}));

vi.mock("@/utils/showToast", () => ({
  showToast: {
    warning: vi.fn(),
    error: vi.fn(),
  },
}));

function mountLink(url: string) {
  const i18n = createI18n({
    legacy: false,
    locale: "ja",
    messages: { ja },
  });
  return mount(SafeExternalUrlLink, {
    props: { url },
    global: { plugins: [i18n] },
  });
}

describe("SafeExternalUrlLink", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("opens the validated href instead of re-parsing the raw URL", async () => {
    vi.mocked(openExternal.openValidatedExternalHttpUrl).mockResolvedValue(
      true,
    );
    const wrapper = mountLink(" https://example.com/path ");

    await wrapper.get("a").trigger("click");

    expect(openExternal.openValidatedExternalHttpUrl).toHaveBeenCalledWith(
      "https://example.com/path",
    );
  });

  it("shows toast when openValidatedExternalHttpUrl rejects", async () => {
    vi.mocked(openExternal.openValidatedExternalHttpUrl).mockRejectedValue(
      new Error("unavailable"),
    );
    const wrapper = mountLink("https://example.com/");
    await wrapper.get("a").trigger("click");
    await vi.waitFor(() => {
      expect(showToast.error).toHaveBeenCalled();
    });
  });

  it("shows warning when open returns false", async () => {
    vi.mocked(openExternal.openValidatedExternalHttpUrl).mockResolvedValue(
      false,
    );
    const wrapper = mountLink("https://example.com/");
    await wrapper.get("a").trigger("click");
    await vi.waitFor(() => {
      expect(showToast.warning).toHaveBeenCalled();
    });
  });
});
