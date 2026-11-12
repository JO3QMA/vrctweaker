import type { App, Component } from "vue";
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

function registerNamed(
  app: App,
  component: Component & { name?: string },
): void {
  const name = component.name;
  if (!name) {
    throw new Error("Element Plus component is missing name");
  }
  app.component(name, component);
}

/**
 * String-template Storybook stories do not get unplugin-vue-components transforms.
 * When adding `<el-...>` to a string-template story, register the component here
 * and import its `element-plus/es/components/.../style/css` above.
 */
export function registerElementPlusFormComponents(app: App): void {
  registerNamed(app, ElInput);
  registerNamed(app, ElSelect);
  registerNamed(app, ElOption);
  registerNamed(app, ElCheckbox);
  registerNamed(app, ElSwitch);
  registerNamed(app, ElForm);
  registerNamed(app, ElFormItem);
}
