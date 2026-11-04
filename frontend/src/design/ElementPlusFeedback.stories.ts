import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { onMounted } from "vue";
import { ElMessageBox } from "element-plus";
import { showToast } from "../utils/showToast";
import { withWailsApp } from "../stories/wailsDecorator";

const meta = {
  title: "Design/Element Plus feedback",
  parameters: {
    controls: { disable: true },
    viewport: { defaultViewport: "desktop1280" },
  },
  decorators: [withWailsApp({})],
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

/** Visual check: toast + confirm dialog styles (message / message-box CSS). */
export const ToastAndConfirmDialog: Story = {
  name: "Toast and confirm dialog",
  render: () => ({
    template: '<div data-testid="feedback-story-host" />',
    setup() {
      onMounted(() => {
        showToast.success("保存しました");
        void ElMessageBox.confirm("この操作を実行しますか？", "確認", {
          confirmButtonText: "OK",
          cancelButtonText: "キャンセル",
          type: "warning",
        }).catch(() => undefined);
      });
    },
  }),
};
