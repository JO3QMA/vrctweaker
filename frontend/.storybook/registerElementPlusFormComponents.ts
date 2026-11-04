import type { App } from "vue";
import {
  ElCheckbox,
  ElForm,
  ElFormItem,
  ElInput,
  ElOption,
  ElSelect,
  ElSwitch,
} from "element-plus";

/** String-template stories do not get unplugin-vue-components transforms. */
export function registerElementPlusFormComponents(app: App): void {
  app.component(ElInput.name!, ElInput);
  app.component(ElSelect.name!, ElSelect);
  app.component(ElOption.name!, ElOption);
  app.component(ElCheckbox.name!, ElCheckbox);
  app.component(ElSwitch.name!, ElSwitch);
  app.component(ElForm.name!, ElForm);
  app.component(ElFormItem.name!, ElFormItem);
}
