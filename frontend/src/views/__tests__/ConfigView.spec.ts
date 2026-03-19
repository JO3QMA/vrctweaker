import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import ConfigView from "../ConfigView.vue";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [{ path: "/config", component: ConfigView }],
});

describe("ConfigView", () => {
  it("renders page title", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    expect(wrapper.find(".page-title").text()).toBe("その他の設定");
  });

  it("shows create button when config does not exist", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    expect(wrapper.find("[data-testid='create-config-btn']").exists()).toBe(
      true,
    );
  });

  it("has camera resolution preset toggles", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    // Simulate creating config to show editor
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    expect(wrapper.find("[data-testid='camera-preset-fhd']").exists()).toBe(
      true,
    );
    expect(wrapper.find("[data-testid='camera-preset-wqhd']").exists()).toBe(
      true,
    );
    expect(wrapper.find("[data-testid='camera-preset-4k']").exists()).toBe(
      true,
    );
    expect(wrapper.find("[data-testid='camera-preset-custom']").exists()).toBe(
      true,
    );
  });

  it("has screenshot resolution preset toggles", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    expect(wrapper.find("[data-testid='screenshot-preset-fhd']").exists()).toBe(
      true,
    );
    expect(wrapper.find("[data-testid='screenshot-preset-4k']").exists()).toBe(
      true,
    );
    expect(
      wrapper.find("[data-testid='screenshot-preset-custom']").exists(),
    ).toBe(true);
  });

  it("disables camera resolution inputs when preset is not custom", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    // Select FHD preset
    const fhdRadio = wrapper.find("[data-testid='camera-preset-fhd']");
    await fhdRadio.setValue(true);
    await wrapper.vm.$nextTick();

    const widthInput = wrapper.find("[data-testid='camera-width-input']");
    const heightInput = wrapper.find("[data-testid='camera-height-input']");
    expect((widthInput.element as HTMLInputElement).disabled).toBe(true);
    expect((heightInput.element as HTMLInputElement).disabled).toBe(true);
  });

  it("enables camera resolution inputs when preset is custom", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    const customRadio = wrapper.find("[data-testid='camera-preset-custom']");
    await customRadio.setValue(true);
    await wrapper.vm.$nextTick();

    const widthInput = wrapper.find("[data-testid='camera-width-input']");
    expect((widthInput.element as HTMLInputElement).disabled).toBe(false);
  });

  it("has save and delete buttons in editor", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    expect(wrapper.find("[data-testid='save-config-btn']").exists()).toBe(true);
    expect(wrapper.find("[data-testid='delete-config-btn']").exists()).toBe(
      true,
    );
  });

  it("has cache settings inputs", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    expect(wrapper.find("[data-testid='cache-size-input']").exists()).toBe(
      true,
    );
    expect(wrapper.find("[data-testid='cache-expiry-input']").exists()).toBe(
      true,
    );
    expect(wrapper.find("[data-testid='cache-directory-input']").exists()).toBe(
      true,
    );
  });

  it("has rich presence toggle", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    expect(
      wrapper.find("[data-testid='disable-rich-presence-checkbox']").exists(),
    ).toBe(true);
  });
});
