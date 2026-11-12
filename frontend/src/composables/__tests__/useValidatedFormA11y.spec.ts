import { describe, expect, it, vi } from "vitest";
import { defineComponent, nextTick, reactive, ref } from "vue";
import { flushPromises, mount } from "@vue/test-utils";
import type { FormInstance } from "element-plus";
import { useValidatedFormA11y } from "../useValidatedFormA11y";

describe("useValidatedFormA11y", () => {
  it("clears aria-invalid when field validateState returns to success", async () => {
    const formEl = document.createElement("form");
    formEl.className = "el-form vt-validated-form";
    const itemEl = document.createElement("div");
    itemEl.className = "el-form-item is-error vt-form-item-with-range-hint";
    itemEl.innerHTML = `
      <div class="el-form-item__content">
        <div class="el-input-number"><input type="text" /></div>
        <p id="width-range-hint" class="form-field-range-hint">1280〜7680</p>
        <div class="el-form-item__error">bad</div>
      </div>
    `;
    formEl.append(itemEl);
    document.body.append(formEl);

    const inner = itemEl.querySelector("input") as HTMLInputElement;
    const field = reactive({
      propString: "width",
      validateState: "error",
      validateMessage: "bad",
      $el: itemEl,
    });
    const form = reactive({
      $el: formEl,
      fields: [field],
    });
    const formRef = ref(form as unknown as FormInstance);

    mount(
      defineComponent({
        setup() {
          useValidatedFormA11y(formRef);
          return () => null;
        },
      }),
    );
    await flushPromises();
    await vi.waitFor(() => {
      expect(inner.getAttribute("aria-invalid")).toBe("true");
    });

    field.validateState = "success";
    field.validateMessage = "";
    itemEl.classList.remove("is-error");
    await nextTick();
    await flushPromises();
    await vi.waitFor(() => {
      expect(inner.getAttribute("aria-invalid")).toBeNull();
    });
    expect(inner.getAttribute("aria-describedby")).toBe("width-range-hint");

    formEl.remove();
  });
});
