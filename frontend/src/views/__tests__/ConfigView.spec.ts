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
    expect(wrapper.find("[data-testid='camera-preset-8k']").exists()).toBe(
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

  it("shows cache size and expiry default to 30", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    const cacheSizeInput = wrapper.find("[data-testid='cache-size-input']");
    const cacheExpiryInput = wrapper.find("[data-testid='cache-expiry-input']");
    expect((cacheSizeInput.element as HTMLInputElement).value).toBe("30");
    expect((cacheExpiryInput.element as HTMLInputElement).value).toBe("30");
  });

  it("clamps cache size to 30 on blur when value is less than 30", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    const cacheSizeInput = wrapper.find("[data-testid='cache-size-input']");
    await cacheSizeInput.setValue(20);
    await cacheSizeInput.trigger("blur");
    await wrapper.vm.$nextTick();

    expect((cacheSizeInput.element as HTMLInputElement).value).toBe("30");
  });

  it("clamps cache expiry to 30 on blur when value is less than 30", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    const cacheExpiryInput = wrapper.find("[data-testid='cache-expiry-input']");
    await cacheExpiryInput.setValue(10);
    await cacheExpiryInput.trigger("blur");
    await wrapper.vm.$nextTick();

    expect((cacheExpiryInput.element as HTMLInputElement).value).toBe("30");
  });

  it("has Steadycam FOV slider and number input", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    expect(wrapper.find("[data-testid='steadycam-fov-slider']").exists()).toBe(
      true,
    );
    expect(wrapper.find("[data-testid='steadycam-fov-input']").exists()).toBe(
      true,
    );
  });

  it("shows Steadycam FOV as empty by default with placeholder 50", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    const input = wrapper.find("[data-testid='steadycam-fov-input']");
    const slider = wrapper.find("[data-testid='steadycam-fov-slider']");
    expect((input.element as HTMLInputElement).value).toBe("");
    expect((input.element as HTMLInputElement).placeholder).toBe("50");
    expect((slider.element as HTMLInputElement).value).toBe("50");
  });

  it("syncs Steadycam FOV slider and number input", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    const slider = wrapper.find("[data-testid='steadycam-fov-slider']");
    const input = wrapper.find("[data-testid='steadycam-fov-input']");

    await slider.setValue(75);
    await wrapper.vm.$nextTick();
    expect((input.element as HTMLInputElement).value).toBe("75");

    await input.setValue("60");
    await input.trigger("input");
    await wrapper.vm.$nextTick();
    expect((slider.element as HTMLInputElement).value).toBe("60");
  });

  it("clamps Steadycam FOV to 30-100 on blur", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    const input = wrapper.find("[data-testid='steadycam-fov-input']");
    await input.setValue(20);
    await input.trigger("input");
    await input.trigger("blur");
    await wrapper.vm.$nextTick();
    expect((input.element as HTMLInputElement).value).toBe("30");
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
