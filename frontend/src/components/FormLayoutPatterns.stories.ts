import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import VtInput from "./VtInput.vue";
import VtSelect from "./VtSelect.vue";
import VtCheckbox from "./VtCheckbox.vue";
import VtSwitch from "./VtSwitch.vue";
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

export const FormLayoutTopLabels: Story = {
  name: "Form layout (top labels)",
  render: () => ({
    components: { VtInput, VtSelect, VtCheckbox },
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
            <VtInput v-model="name" placeholder="Name" />
          </el-form-item>
          <el-form-item label="Language">
            <VtSelect v-model="lang">
              <el-option value="ja" label="日本語" />
              <el-option value="en" label="English" />
            </VtSelect>
          </el-form-item>
          <el-form-item>
            <VtCheckbox v-model="desktop">Desktop mode</VtCheckbox>
          </el-form-item>
        </el-form>
      </div>
    `,
  }),
};

export const SettingRow: Story = {
  name: "Setting row",
  render: () => ({
    components: { VtSwitch },
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
        <VtSwitch v-model="suppressSleep" class="vt-setting-row-story__control" />
      </div>
    `,
  }),
};

export const PathInputRow: Story = {
  name: "Path input row",
  render: () => ({
    components: { VtInput, VtButton },
    setup() {
      const path = ref("C:\\\\VRChat\\\\output_log.txt");
      return { path };
    },
    template: `
      <div class="vt-form-layout-story">
        <label>Output log path</label>
        <div class="vt-path-input-group-story">
          <VtInput v-model="path" />
          <VtButton variant="secondary">Browse</VtButton>
        </div>
      </div>
    `,
  }),
};
