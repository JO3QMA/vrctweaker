import { describe, it, expect, afterEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import LegalDocLinks from "../LegalDocLinks.vue";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: "/legal/privacy", component: { template: "<div />" } },
    { path: "/legal/terms", component: { template: "<div />" } },
  ],
});

describe("LegalDocLinks", () => {
  afterEach(() => {
    document.body.innerHTML = "";
  });

  it("uses in-app routes by default", async () => {
    await router.push("/");
    await router.isReady();
    const wrapper = mount(LegalDocLinks, {
      global: { plugins: [router] },
    });
    const privacy = wrapper.find('[data-testid="legal-link-privacy"]');
    expect(privacy.attributes("href")).toContain("/legal/privacy");
    expect(wrapper.find('[data-testid="legal-github-privacy"]').exists()).toBe(
      true,
    );
  });

  it("opens dialog in dialog presentation", async () => {
    await router.push("/legal/privacy");
    await router.isReady();
    const wrapper = mount(LegalDocLinks, {
      props: { presentation: "dialog", showGitHubSupplement: false },
      global: { plugins: [router] },
      attachTo: document.body,
    });
    await wrapper.find('[data-testid="legal-link-terms"]').trigger("click");
    await flushPromises();
    expect(
      document.querySelector('[data-testid="legal-document-dialog"]'),
    ).not.toBeNull();
    wrapper.unmount();
  });
});
