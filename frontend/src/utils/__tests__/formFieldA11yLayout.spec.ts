import { readFileSync } from "node:fs";
import path from "node:path";
import { afterAll, beforeAll, describe, expect, it } from "vitest";
import { defineComponent, ref } from "vue";
import { flushPromises, mount } from "@vue/test-utils";
import ElementPlus from "element-plus";
import type { FormInstance } from "element-plus";
import { parseStyleSheet } from "../../design/cssCustomProperties";
import { syncValidatedFormFieldA11y } from "../formFieldA11y";

const styleCssPath = path.resolve(process.cwd(), "src/assets/style.css");
const styleCss = readFileSync(styleCssPath, "utf8");
const parsedStyle = parseStyleSheet(styleCss);

let globalStyleEl: HTMLStyleElement;

beforeAll(() => {
  globalStyleEl = document.createElement("style");
  globalStyleEl.setAttribute("data-testid", "vt-global-style");
  globalStyleEl.textContent = styleCss;
  document.head.append(globalStyleEl);
});

afterAll(() => {
  globalStyleEl?.remove();
});

function declarationValue(
  selectorPart: string,
  property: string,
): string | undefined {
  return parsedStyle.declarations.find(
    (d) => d.selector.includes(selectorPart) && d.property === property,
  )?.value;
}

const GridHintForm = defineComponent({
  setup() {
    const formRef = ref<FormInstance>();
    const model = ref({ width: 100 });
    return { formRef, model };
  },
  template: `
    <el-form ref="formRef" :model="model" class="vt-validated-form">
      <el-form-item prop="width" class="vt-form-item-with-range-hint resolution-form-item">
        <el-input-number v-model="model.width" data-testid="width-input" />
        <p id="width-range-hint" class="form-field-range-hint">1280〜7680</p>
      </el-form-item>
    </el-form>
  `,
});

describe("vt-form-item-with-range-hint layout", () => {
  it("declares grid layout on form item content in style.css", () => {
    expect(declarationValue("vt-form-item-with-range-hint", "display")).toBe(
      "grid",
    );
  });

  it("uses CSS grid on el-form-item__content so hint is below the control", async () => {
    const wrapper = mount(GridHintForm, {
      global: { plugins: [ElementPlus] },
      attachTo: document.body,
    });
    await flushPromises();
    document.head.append(globalStyleEl);

    const content = wrapper.element.querySelector(
      ".vt-form-item-with-range-hint .el-form-item__content",
    ) as HTMLElement | null;
    expect(content).toBeTruthy();
    expect(getComputedStyle(content!).display).toBe("grid");

    const hint = wrapper.get(".form-field-range-hint").element as HTMLElement;
    const inputWrap = wrapper.get(".el-input-number").element as HTMLElement;
    expect(
      inputWrap.compareDocumentPosition(hint) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();

    wrapper.unmount();
  });
});

describe("syncValidatedFormFieldA11y on el-input-number", () => {
  it("sets aria on the inner input, not the wrapper", async () => {
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

    const inner = itemEl.querySelector(
      ".el-input-number input",
    ) as HTMLInputElement;
    const formRef = {
      $el: formEl,
      fields: [
        {
          propString: "width",
          validateState: "error",
          $el: itemEl,
        },
      ],
    } as unknown as FormInstance;

    await syncValidatedFormFieldA11y(formRef);

    expect(inner.getAttribute("aria-invalid")).toBe("true");
    expect(inner.getAttribute("aria-describedby")).toContain(
      "width-range-hint",
    );

    formRef.fields[0].validateState = "";
    await syncValidatedFormFieldA11y(formRef);
    expect(inner.getAttribute("aria-invalid")).toBeNull();
    const describedBy = inner.getAttribute("aria-describedby") ?? "";
    expect(describedBy).toBe("width-range-hint");
    expect(describedBy).not.toContain("-error");

    formEl.remove();
  });

  it("mirrors disabled to aria-disabled on inner input", async () => {
    const formEl = document.createElement("form");
    const itemEl = document.createElement("div");
    itemEl.className = "el-form-item";
    itemEl.innerHTML =
      '<div class="el-form-item__content"><div class="el-input-number"><input disabled /></div></div>';
    formEl.append(itemEl);
    document.body.append(formEl);
    const inner = itemEl.querySelector("input") as HTMLInputElement;
    const formRef = {
      $el: formEl,
      fields: [{ propString: "w", validateState: "", $el: itemEl }],
    } as unknown as FormInstance;
    await syncValidatedFormFieldA11y(formRef);
    expect(inner.getAttribute("aria-disabled")).toBe("true");
    inner.disabled = false;
    await syncValidatedFormFieldA11y(formRef);
    expect(inner.getAttribute("aria-disabled")).toBeNull();
    formEl.remove();
  });
});

describe("resolution-fields responsive layout", () => {
  it("allows flex-wrap so row width does not exceed container", () => {
    const row = document.createElement("div");
    row.className = "resolution-fields";
    document.body.append(row);
    expect(getComputedStyle(row).flexWrap).toBe("wrap");
    row.remove();
  });

  it("declares scroll padding on automation editor scroll area", () => {
    expect(
      declarationValue("automation-editor__scroll", "padding-bottom"),
    ).toContain("--automation-action-bar-inset");
  });

  it("uses 8px margin on automation inline form items", () => {
    const root = document.createElement("div");
    root.className = "automation-rule-editor";
    const form = document.createElement("form");
    form.className = "el-form vt-validated-form";
    const item = document.createElement("div");
    item.className =
      "el-form-item el-form-item--default automation-inline-form-item";
    form.append(item);
    root.append(form);
    document.body.append(root);
    expect(getComputedStyle(item).marginBottom).toBe("var(--space-8)");
    root.remove();
  });
});
