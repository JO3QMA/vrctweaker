import { describe, it, expect } from "vitest";
import {
  AutomationItemParseError,
  dtoToEditor,
} from "../automationEditorMapping";
import type { AutomationItemDTO } from "../../wails/app";

const base: AutomationItemDTO = {
  id: "1",
  name: "n",
  kind: "rule",
  isEnabled: true,
  triggerType: "friend_joined",
  conditionsJson: "[]",
  actionsJson: JSON.stringify([
    { type: "change_status", payload: { status: "busy" } },
  ]),
};

describe("dtoToEditor", () => {
  it("maps valid JSON fields", () => {
    const state = dtoToEditor({
      ...base,
      scheduleJson: JSON.stringify({ weekdays: [1], hour: 9, minute: 30 }),
      conditionsJson: JSON.stringify([
        { type: "vrchat_running" },
        { type: "friend_is", vrcUserId: "usr_x" },
      ]),
    });
    expect(state.scheduleHour).toBe(9);
    expect(state.vrchatRunning).toBe(true);
    expect(state.friendUserId).toBe("usr_x");
    expect(state.actions[0]?.status).toBe("busy");
  });

  it("throws on invalid actionsJson instead of silently defaulting", () => {
    expect(() => dtoToEditor({ ...base, actionsJson: "{not-json" })).toThrow(
      AutomationItemParseError,
    );
  });

  it("throws on invalid conditionsJson", () => {
    expect(() =>
      dtoToEditor({ ...base, conditionsJson: "not-array{" }),
    ).toThrow(AutomationItemParseError);
  });

  it("throws on invalid scheduleJson", () => {
    expect(() => dtoToEditor({ ...base, scheduleJson: "{bad" })).toThrow(
      AutomationItemParseError,
    );
  });
});
