import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { expect, userEvent, waitFor, within } from "storybook/test";
import ja from "../i18n/locales/ja.json";
import AutomationView from "./AutomationView.vue";
import {
  AUTOMATION_STORY_FRIEND_RULE_NAME,
  AUTOMATION_STORY_RUN_LOG,
  withAutomationWails,
} from "../stories/fixtures/automationWails";

const meta = {
  title: "Views/AutomationView",
  component: AutomationView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
  decorators: [withAutomationWails()],
} satisfies Meta<typeof AutomationView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Rules: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(
      await canvas.findByText(AUTOMATION_STORY_FRIEND_RULE_NAME),
    ).toBeVisible();
  },
};

export const EditingRule: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(
      await canvas.findByText(AUTOMATION_STORY_FRIEND_RULE_NAME),
    );
    await expect(await canvas.findByTestId("item-name-input")).toHaveValue(
      AUTOMATION_STORY_FRIEND_RULE_NAME,
    );
    await expect(
      await canvas.findByText(ja.automation.sectionWhen),
    ).toBeVisible();
  },
};

export const RunLog: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(
      await canvas.findByRole("tab", {
        name: new RegExp(ja.automation.runLogTitle),
      }),
    );
    const log = within(await canvas.findByTestId("automation-run-log"));
    await expect(
      await log.findByTestId("automation-log-event"),
    ).toHaveTextContent(ja.automation.triggerSchedule);
  },
};

export const NarrowList: Story = {
  decorators: [
    (story) => ({
      components: { story },
      template: '<div style="width: 600px; height: 720px"><story /></div>',
    }),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(
      await canvas.findByText(AUTOMATION_STORY_FRIEND_RULE_NAME),
    ).toBeVisible();
  },
};

const AUTOMATION_STORY_UNKNOWN_RULE_NAME = "未対応トリガーのルール";
const AUTOMATION_STORY_UNKNOWN_TRIGGER = "future.unknown_trigger";

export const UnknownTrigger: Story = {
  parameters: {
    automationItems: [
      {
        id: "rule-unknown",
        name: AUTOMATION_STORY_UNKNOWN_RULE_NAME,
        kind: "rule" as const,
        isEnabled: true,
        triggerType: AUTOMATION_STORY_UNKNOWN_TRIGGER,
        conditionsJson: "[]",
        actionsJson: "[]",
      },
    ],
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(
      await canvas.findByText(ja.automation.triggerUnknown),
    ).toBeVisible();
    expect(canvas.queryByText(AUTOMATION_STORY_UNKNOWN_TRIGGER)).toBeNull();
    await userEvent.click(
      await canvas.findByText(AUTOMATION_STORY_UNKNOWN_RULE_NAME),
    );
    const editor = within(await canvas.findByTestId("rule-editor"));
    await expect(
      await editor.findByText(ja.automation.triggerUnknown),
    ).toBeVisible();
    expect(editor.queryByText(AUTOMATION_STORY_UNKNOWN_TRIGGER)).toBeNull();
  },
};

/** 600px frame is under 720px, so the log shows four columns and the result tooltip. */
export const NarrowRunLog: Story = {
  decorators: [
    (story) => ({
      components: { story },
      template: '<div style="width: 600px; height: 720px"><story /></div>',
    }),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(
      await canvas.findByRole("tab", {
        name: new RegExp(ja.automation.runLogTitle),
      }),
    );
    const log = within(await canvas.findByTestId("automation-run-log"));
    const entry = AUTOMATION_STORY_RUN_LOG[0];
    if (!entry) throw new Error("missing run log fixture");
    await expect(
      await log.findByTestId("automation-log-col-event"),
    ).toBeVisible();
    await waitFor(() => {
      expect(log.queryByTestId("automation-log-col-progress")).toBeNull();
    });
    await userEvent.hover(await log.findByTestId("automation-log-result"));
    await expect(
      await within(document.body).findByTestId("automation-log-result-tooltip"),
    ).toHaveTextContent(
      `${ja.automation.runLogActions} ${entry.actionsCompleted}/${entry.actionsTotal}`,
    );
  },
};
