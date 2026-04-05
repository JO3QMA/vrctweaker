import type { Meta, StoryObj } from "@storybook/vue3-vite";
import SettingsView from "./SettingsView.vue";
import {
  settingsViewWailsDecorator,
  type SettingsViewWailsPreset,
} from "./settingsViewStoryDecorator";

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
  return [settingsViewWailsDecorator(preset)];
}

/**
 * Wails スタブ無し。`window.go` が無いため callApp はフォールバック値になり、
 * 本番の「未接続」に近いが、資格情報ストーリーはデコレータ付きバリアントを参照。
 */
export const Default: Story = {};

/** 保存資格情報なし・再ログイン案内（unlock 警告なし） */
export const NeedsReloginNoWarning: Story = {
  decorators: withPreset("needsReloginClean"),
};

/** 起動時アンロック失敗（セッション切れ等）で警告 el-alert を表示 */
export const NeedsReloginWithUnlockWarning: Story = {
  decorators: withPreset("needsReloginUnlockError"),
};

/** ログイン済み・プロフィールカード表示 */
export const LoggedInWithProfile: Story = {
  decorators: withPreset("loggedIn"),
};
