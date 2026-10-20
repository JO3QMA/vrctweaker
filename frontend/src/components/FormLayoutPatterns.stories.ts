import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ElForm, ElFormItem, ElOption } from "element-plus";
import { ref } from "vue";
import VtButton from "./VtButton.vue";
import "./FormControl.stories.css";

const catalogParameters = {
  controls: { disable: true },
};

const meta = {
  title: "Design/Form layout patterns",
  tags: ["autodocs"],
  parameters: catalogParameters,
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

export const FormControls: Story = {
  name: "Form controls",
  render: () => ({
    components: { ElOption },
    setup() {
      const name = ref("");
      const lang = ref("ja");
      const desktop = ref(true);
      const enabled = ref(true);
      const compact = ref(false);
      return { name, lang, desktop, enabled, compact };
    },
    template: `
      <div class="vt-form-layout-story">
        <el-input v-model="name" placeholder="Profile name" clearable />
        <el-input placeholder="Profile name" disabled />
        <el-select v-model="lang" placeholder="Language">
          <el-option value="ja" label="日本語" />
          <el-option value="en" label="English" />
        </el-select>
        <el-select v-model="lang" disabled placeholder="Language">
          <el-option value="ja" label="日本語" />
        </el-select>
        <el-checkbox v-model="desktop">Desktop mode</el-checkbox>
        <el-checkbox v-model="desktop" disabled>Desktop mode</el-checkbox>
        <el-switch v-model="enabled" aria-label="Enable feature" />
        <el-switch v-model="enabled" disabled aria-label="Enable feature" />
        <el-switch v-model="compact" size="small" aria-label="Enable list item" />
      </div>
    `,
  }),
};

export const FormLayoutTopLabels: Story = {
  name: "Form layout (top labels)",
  render: () => ({
    components: { ElForm, ElFormItem, ElOption },
    setup() {
      const name = ref("Desktop");
      const lang = ref("ja");
      const desktop = ref(true);
      return { name, lang, desktop };
    },
    template: `
      <div class="vt-form-layout-story">
        <el-form label-position="top" size="default">
          <el-form-item label="Profile name">
            <el-input v-model="name" placeholder="Name" />
          </el-form-item>
          <el-form-item label="Language">
            <el-select v-model="lang">
              <el-option value="ja" label="日本語" />
              <el-option value="en" label="English" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-checkbox v-model="desktop">Desktop mode</el-checkbox>
          </el-form-item>
        </el-form>
      </div>
    `,
  }),
};

export const SettingRow: Story = {
  name: "Setting row",
  render: () => ({
    setup() {
      const suppressSleep = ref(true);
      return { suppressSleep };
    },
    template: `
      <div class="vt-setting-row-story">
        <div class="vt-setting-row-story__label">
          <span>Suppress sleep while VRChat is running</span>
          <span class="vt-setting-row-story__hint">Keeps the PC awake during sessions.</span>
        </div>
        <el-switch v-model="suppressSleep" class="vt-setting-row-story__control" />
      </div>
    `,
  }),
};

export const PathInputRow: Story = {
  name: "Path input row",
  render: () => ({
    components: { VtButton },
    setup() {
      const path = ref("C:\\VRChat\\output_log.txt");
      return { path };
    },
    template: `
      <div class="vt-form-layout-story">
        <label for="output-log-path">Output log path</label>
        <div class="vt-path-input-group-story">
          <el-input id="output-log-path" v-model="path" />
          <VtButton variant="secondary">Browse</VtButton>
        </div>
      </div>
    `,
  }),
};
