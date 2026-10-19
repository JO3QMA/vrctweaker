import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import LegalDocumentView from "../LegalDocumentView.vue";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: "/legal/privacy",
      component: LegalDocumentView,
      props: { kind: "privacy" },
    },
  ],
});

describe("LegalDocumentView", () => {
  it("renders bundled privacy markdown", async () => {
    await router.push("/legal/privacy");
    await router.isReady();
    const wrapper = mount(LegalDocumentView, {
      props: { kind: "privacy" },
      global: { plugins: [router] },
    });
    expect(wrapper.find(".page-title").text()).toBe("プライバシーポリシー");
    const body = wrapper.find('[data-testid="legal-markdown-body"]');
    expect(body.exists()).toBe(true);
    expect(body.html()).toContain("VRChat Tweaker");
    expect(body.html()).not.toContain("<script");
  });
});
