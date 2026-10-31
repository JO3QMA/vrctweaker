import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import LegalMarkdownBody from "../LegalMarkdownBody.vue";
import * as openExternal from "@/utils/openExternalHttpUrl";

vi.mock("@/utils/openExternalHttpUrl", () => ({
  openExternalHttpUrl: vi.fn(),
}));

describe("LegalMarkdownBody", () => {
  it("revalidates the decoded href before opening it", async () => {
    vi.mocked(openExternal.openExternalHttpUrl).mockResolvedValue(true);
    const wrapper = mount(LegalMarkdownBody, {
      props: {
        markdown: "See [Example](https://example.com/path?q=1).",
      },
    });

    await wrapper.get(".legal-md-external-link").trigger("click");

    expect(openExternal.openExternalHttpUrl).toHaveBeenCalledWith(
      "https://example.com/path?q=1",
    );
  });
});
