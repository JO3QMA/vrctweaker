import { describe, it, expect, vi } from "vitest";
import type { FormInstance } from "element-plus";
import { nextTick } from "vue";
import { validateElForm } from "../elFormValidate";
import * as formFieldA11y from "../formFieldA11y";
import {
  formFieldValidationErrorId,
  scrollFormItemIntoView,
  syncValidatedFormFieldA11y,
} from "../formFieldA11y";

function mockErrorFormItem(options: {
  prop: string;
  withRangeHint?: boolean;
}): {
  formEl: HTMLFormElement;
  itemEl: HTMLDivElement;
  input: HTMLInputElement;
} {
  const formEl = document.createElement("form");
  formEl.className = "el-form vt-validated-form";

  const itemEl = document.createElement("div");
  itemEl.className = "el-form-item is-error";

  const input = document.createElement("input");
  input.type = "text";
  input.dataset.testid = "field-input";

  const error = document.createElement("div");
  error.className = "el-form-item__error";
  error.textContent = "validation failed";

  itemEl.append(input, error);
  if (options.withRangeHint) {
    const hint = document.createElement("p");
    hint.id = "width-range-hint";
    hint.className = "form-field-range-hint";
    hint.textContent = "1280〜7680";
    itemEl.append(hint);
  }

  formEl.append(itemEl);
  document.body.append(formEl);

  return { formEl, itemEl, input };
}

function mockFormRef(
  formEl: HTMLFormElement,
  itemEl: HTMLDivElement,
  prop: string,
  validateState: "" | "error" = "error",
): FormInstance {
  return {
    $el: formEl,
    fields: [
      {
        propString: prop,
        validateState,
        $el: itemEl,
      },
    ],
    scrollToField: vi.fn(),
    validate: vi.fn(),
  } as unknown as FormInstance;
}

describe("syncValidatedFormFieldA11y", () => {
  it("sets aria-invalid and aria-describedby to error id and range hint", async () => {
    const { formEl, itemEl, input } = mockErrorFormItem({
      prop: "width",
      withRangeHint: true,
    });
    const formRef = mockFormRef(formEl, itemEl, "width");

    await syncValidatedFormFieldA11y(formRef);

    const scopeId = formEl.getAttribute("data-vt-form-a11y-scope");
    expect(scopeId).toBeTruthy();
    const errorId = formFieldValidationErrorId(scopeId!, "width");

    expect(input.getAttribute("aria-invalid")).toBe("true");
    expect(input.getAttribute("aria-describedby")).toBe(
      `${errorId} width-range-hint`,
    );
    expect(itemEl.querySelector(`#${CSS.escape(errorId)}`)).toBeTruthy();

    formEl.remove();
  });

  it("clears aria-invalid and keeps range hint when field is valid", async () => {
    const { formEl, itemEl, input } = mockErrorFormItem({
      prop: "width",
      withRangeHint: true,
    });
    const formRef = mockFormRef(formEl, itemEl, "width", "");

    await syncValidatedFormFieldA11y(formRef);

    expect(input.getAttribute("aria-invalid")).toBeNull();
    expect(input.getAttribute("aria-describedby")).toBe("width-range-hint");

    formEl.remove();
  });
});

describe("validateElForm", () => {
  it("syncs a11y after validation failure", async () => {
    const { formEl, itemEl, input } = mockErrorFormItem({
      prop: "scheduleHour",
    });
    itemEl.scrollIntoView = vi.fn();
    const formRef = mockFormRef(formEl, itemEl, "scheduleHour");
    formRef.validate = vi.fn().mockRejectedValue({
      scheduleHour: [{ message: "bad" }],
    });

    const ok = await validateElForm(formRef);
    expect(ok).toBe(false);
    expect(itemEl.scrollIntoView).toHaveBeenCalled();

    await nextTick();
    expect(input.getAttribute("aria-invalid")).toBe("true");
    expect(input.getAttribute("aria-describedby")).toContain(
      "-scheduleHour-error",
    );

    formEl.remove();
  });

  it("returns true when validate passes even if a11y sync throws", async () => {
    const { formEl, itemEl } = mockErrorFormItem({ prop: "name" });
    const formRef = mockFormRef(formEl, itemEl, "name", "");
    formRef.validate = vi.fn().mockResolvedValue(true);
    const syncSpy = vi
      .spyOn(formFieldA11y, "syncValidatedFormFieldA11y")
      .mockRejectedValue(new Error("a11y failed"));

    const ok = await validateElForm(formRef);
    expect(ok).toBe(true);
    syncSpy.mockRestore();
    formEl.remove();
  });
});

describe("scrollFormItemIntoView", () => {
  it("scrolls down when the field bottom is under the sticky bar inset (600px)", () => {
    const item = document.createElement("div");
    vi.spyOn(item, "getBoundingClientRect").mockReturnValue({
      top: 500,
      bottom: 580,
      left: 0,
      right: 120,
      width: 120,
      height: 80,
      x: 0,
      y: 500,
      toJSON: () => ({}),
    } as DOMRect);
    item.scrollIntoView = vi.fn();

    Object.defineProperty(window, "innerHeight", {
      configurable: true,
      value: 600,
    });
    const scrollBy = vi.fn();
    window.scrollBy = scrollBy;

    scrollFormItemIntoView(item, 88);

    expect(item.scrollIntoView).toHaveBeenCalled();
    expect(scrollBy).toHaveBeenCalledWith({
      top: 580 - (600 - 88),
      behavior: "auto",
    });
  });
});
