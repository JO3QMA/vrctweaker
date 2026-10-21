import type { Decorator, Meta, StoryObj } from "@storybook/vue3-vite";
import { expect, userEvent, within } from "storybook/test";
import { ref } from "vue";
import { useRouter } from "vue-router";
import SettingsView from "./SettingsView.vue";
import {
  withSettingsWails,
  type SettingsViewWailsPreset,
} from "../stories/fixtures/settingsWails";

const meta = {
  title: "Views/SettingsView",
  component: SettingsView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof SettingsView>;

export default meta;
type Story = StoryObj<typeof meta>;

function withPreset(preset: SettingsViewWailsPreset) {
  return [withSettingsWails(preset)];
}

/** Wait until the storybook router is on a settings sub-route, then render. */
function withSettingsRoute(name: string): Decorator {
  return (story) => ({
    components: { story },
    template: '<story v-if="ready" />',
    setup() {
      const router = useRouter();
      const ready = ref(false);
      void router.push({ name }).finally(() => {
        ready.value = true;
      });
      return { ready };
    },
  });
}

function withNarrowFrame(): Decorator {
  return (story) => ({
    components: { story },
    template: '<div style="width: 400px"><story /></div>',
  });
}

/**
 * Wails スタブ無し。`window.go` が無いため callApp はフォールバック値になり、
 * 本番の「未接続」に近いが、資格情報ストーリーはデコレータ付きバリアントを参照。
 */
export const Default: Story = {
  decorators: [withSettingsRoute("settings-account")],
};

/** 保存資格情報なし・再ログイン案内（unlock 警告なし） */
export const NeedsReloginNoWarning: Story = {
  decorators: [
    withSettingsRoute("settings-account"),
    ...withPreset("needsReloginClean"),
  ],
};

/** 起動時アンロック失敗（セッション切れ等）で警告 el-alert を表示 */
export const NeedsReloginWithUnlockWarning: Story = {
  decorators: [
    withSettingsRoute("settings-account"),
    ...withPreset("needsReloginUnlockError"),
  ],
};

/** ログイン済み・プロフィールカード表示 */
export const LoggedInWithProfile: Story = {
  decorators: [
    withSettingsRoute("settings-account"),
    ...withPreset("loggedIn"),
  ],
};

export const Account: Story = {
  decorators: [
    withSettingsRoute("settings-account"),
    ...withPreset("needsReloginClean"),
  ],
};

export const General: Story = {
  decorators: [
    withSettingsRoute("settings-general"),
    ...withPreset("loggedIn"),
  ],
};

export const Data: Story = {
  decorators: [withSettingsRoute("settings-data"), ...withPreset("loggedIn")],
};

export const About: Story = {
  decorators: [withSettingsRoute("settings-about"), ...withPreset("loggedIn")],
};

async function showFailedUpdate(canvasElement: HTMLElement): Promise<void> {
  const canvas = within(canvasElement);
  await userEvent.click(await canvas.findByTestId("settings-check-app-update"));
  await expect(
    await canvas.findByTestId("settings-app-update-error"),
  ).toBeVisible();
}

/** 更新確認が失敗し、長い詳細つきの警告がボタン直下に出る */
export const AboutUpdateCheckFailed: Story = {
  decorators: [
    withSettingsRoute("settings-about"),
    ...withPreset("updateCheckFailed"),
  ],
  play: async ({ canvasElement }) => {
    await showFailedUpdate(canvasElement);
  },
};

/** パネル幅 400px。失敗した更新確認の警告が横にはみ出さない */
export const NarrowAboutUpdateCheckFailed: Story = {
  decorators: [
    withNarrowFrame(),
    withSettingsRoute("settings-about"),
    ...withPreset("updateCheckFailed"),
  ],
  play: async ({ canvasElement }) => {
    await showFailedUpdate(canvasElement);
  },
};

export const Reset: Story = {
  decorators: [withSettingsRoute("settings-reset"), ...withPreset("loggedIn")],
};

/** パネル幅 400px。パスのボタンは入力の下、数値とスイッチはラベルの右に残る。 */
export const Narrow: Story = {
  decorators: [
    withNarrowFrame(),
    withSettingsRoute("settings-data"),
    ...withPreset("loggedIn"),
  ],
};

/** パネル幅 400px。言語セレクトはラベルの下へ全幅、スイッチは右に残る。 */
export const NarrowGeneral: Story = {
  decorators: [
    withNarrowFrame(),
    withSettingsRoute("settings-general"),
    ...withPreset("loggedIn"),
  ],
};
