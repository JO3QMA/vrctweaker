import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { createRouter, createWebHashHistory } from "vue-router";
import { ElSlider } from "element-plus";
import ConfigView from "../ConfigView.vue";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [{ path: "/config", component: ConfigView }],
});

/** ElRadioButton の root label 要素を返す（exists チェック用） */
function radioLabel(wrapper: ReturnType<typeof mount>, testId: string) {
  return wrapper.find(`[data-testid="${testId}"]`);
}

/** ElRadioButton 内の native radio input を返す（checked / click 用） */
function radioInput(wrapper: ReturnType<typeof mount>, testId: string) {
  return wrapper.find(`[data-testid="${testId}"] input`);
}

/** ElInputNumber 内の native input を返す（value / disabled / setValue 用） */
function numInput(wrapper: ReturnType<typeof mount>, testId: string) {
  return wrapper.find(`[data-testid="${testId}"] input`);
}

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
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    expect(radioLabel(wrapper, "camera-preset-fhd").exists()).toBe(true);
    expect(radioLabel(wrapper, "camera-preset-wqhd").exists()).toBe(true);
    expect(radioLabel(wrapper, "camera-preset-4k").exists()).toBe(true);
    expect(radioLabel(wrapper, "camera-preset-8k").exists()).toBe(true);
    expect(radioLabel(wrapper, "camera-preset-custom").exists()).toBe(true);
  });

  it("has screenshot resolution preset toggles", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    expect(radioLabel(wrapper, "screenshot-preset-fhd").exists()).toBe(true);
    expect(radioLabel(wrapper, "screenshot-preset-4k").exists()).toBe(true);
    expect(radioLabel(wrapper, "screenshot-preset-custom").exists()).toBe(true);
  });

  it("disables camera resolution inputs when preset is not custom", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    // FHD プリセットを選択（inner radio input に setValue で確実に変更）
    await radioInput(wrapper, "camera-preset-fhd").setValue(true);
    await wrapper.vm.$nextTick();

    // ElInputNumber 内の input で disabled を確認
    expect(
      (numInput(wrapper, "camera-width-input").element as HTMLInputElement)
        .disabled,
    ).toBe(true);
    expect(
      (numInput(wrapper, "camera-height-input").element as HTMLInputElement)
        .disabled,
    ).toBe(true);
  });

  it("enables camera resolution inputs when preset is custom", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    await radioInput(wrapper, "camera-preset-custom").setValue(true);
    await wrapper.vm.$nextTick();

    expect(
      (numInput(wrapper, "camera-width-input").element as HTMLInputElement)
        .disabled,
    ).toBe(false);
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

    // ElInputNumber の内側 input で value を確認
    expect(
      (numInput(wrapper, "cache-size-input").element as HTMLInputElement).value,
    ).toBe("30");
    expect(
      (numInput(wrapper, "cache-expiry-input").element as HTMLInputElement)
        .value,
    ).toBe("30");
  });

  it("clamps cache size to 30 on blur when value is less than 30", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    const cacheInner = numInput(wrapper, "cache-size-input");
    await cacheInner.setValue(20);
    await cacheInner.trigger("blur");
    await wrapper.vm.$nextTick();

    expect((cacheInner.element as HTMLInputElement).value).toBe("30");
  });

  it("clamps cache expiry to 30 on blur when value is less than 30", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    const expiryInner = numInput(wrapper, "cache-expiry-input");
    await expiryInner.setValue(10);
    await expiryInner.trigger("blur");
    await wrapper.vm.$nextTick();

    expect((expiryInner.element as HTMLInputElement).value).toBe("30");
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

  it("shows Steadycam FOV input as empty by default with placeholder 50", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    // ElInputNumber の内側 input で確認
    const fovInner = numInput(wrapper, "steadycam-fov-input")
      .element as HTMLInputElement;
    expect(fovInner.value).toBe("");
    expect(fovInner.placeholder).toBe("50");

    // ElSlider の modelValue がデフォルト値 (50) になっていることを確認
    const sliderComp = wrapper.findComponent(ElSlider);
    expect(sliderComp.props("modelValue")).toBe(50);
  });

  it("syncs Steadycam FOV slider and number input", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    const sliderComp = wrapper.findComponent(ElSlider);
    const fovInner = numInput(wrapper, "steadycam-fov-input");

    // スライダーの input イベントをエミット → 数値入力に反映
    await sliderComp.vm.$emit("input", 75);
    await wrapper.vm.$nextTick();
    expect((fovInner.element as HTMLInputElement).value).toBe("75");

    // 数値入力を変更 → スライダーに反映
    await fovInner.setValue("60");
    await fovInner.trigger("input");
    await wrapper.vm.$nextTick();
    expect(sliderComp.props("modelValue")).toBe(60);
  });

  it("clamps Steadycam FOV to 30-100 on blur", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    const fovInner = numInput(wrapper, "steadycam-fov-input");
    await fovInner.setValue(20);
    await fovInner.trigger("input");
    await fovInner.trigger("blur");
    await wrapper.vm.$nextTick();
    expect((fovInner.element as HTMLInputElement).value).toBe("30");
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

  it("sets aria-label on camera and screenshot resolution radio groups", async () => {
    await router.push("/config");
    await router.isReady();
    const wrapper = mount(ConfigView, {
      global: { plugins: [router] },
    });
    await wrapper.find("[data-testid='create-config-btn']").trigger("click");
    await wrapper.vm.$nextTick();

    const cam = wrapper
      .get("[data-testid='camera-preset-hd']")
      .element.closest("[role='radiogroup']");
    expect(cam?.getAttribute("aria-label")).toBe("カメラ解像度プリセット");

    const shot = wrapper
      .get("[data-testid='screenshot-preset-hd']")
      .element.closest("[role='radiogroup']");
    expect(shot?.getAttribute("aria-label")).toBe(
      "スクリーンショット解像度プリセット",
    );
  });
});
