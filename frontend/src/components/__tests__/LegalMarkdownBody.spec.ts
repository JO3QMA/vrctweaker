import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import LegalMarkdownBody from "../LegalMarkdownBody.vue";
import * as openExternal from "@/utils/openExternalHttpUrl";

vi.mock("@/utils/openExternalHttpUrl", () => ({
  openValidatedExternalHttpUrl: vi.fn(),
}));

describe("LegalMarkdownBody", () => {
  it("opens the decoded href through the validated opener", async () => {
    vi.mocked(openExternal.openValidatedExternalHttpUrl).mockResolvedValue(
      true,
    );
    const wrapper = mount(LegalMarkdownBody, {
      props: {
        markdown: "See [Example](https://example.com/path?q=1).",
      },
    });

    await wrapper.get(".legal-md-external-link").trigger("click");

    expect(openExternal.openValidatedExternalHttpUrl).toHaveBeenCalledWith(
      "https://example.com/path?q=1",
    );
  });
});
