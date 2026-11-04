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

import "element-plus/es/components/checkbox/style/css";
import "element-plus/es/components/form/style/css";
import "element-plus/es/components/form-item/style/css";
import "element-plus/es/components/input/style/css";
import "element-plus/es/components/option/style/css";
import "element-plus/es/components/select/style/css";
import "element-plus/es/components/switch/style/css";

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
