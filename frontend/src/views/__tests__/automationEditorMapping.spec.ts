import { describe, it, expect, vi, afterEach } from "vitest";
import {
  AutomationItemParseError,
  dtoToEditor,
  newAutomationId,
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

  it("maps set_vrchat_window_size width/height", () => {
    const state = dtoToEditor({
      ...base,
      actionsJson: JSON.stringify([
        {
          type: "set_vrchat_window_size",
          payload: { width: 960, height: 540 },
        },
      ]),
    });
    expect(state.actions[0]?.type).toBe("set_vrchat_window_size");
    expect(state.actions[0]?.windowWidth).toBe(960);
    expect(state.actions[0]?.windowHeight).toBe(540);
  });

  it("throws on invalid actionsJson instead of silently defaulting", () => {
    expect(() => dtoToEditor({ ...base, actionsJson: "{not-json" })).toThrow(
      AutomationItemParseError,
    );
  });

  it("throws when actionsJson is not an array", () => {
    expect(() => dtoToEditor({ ...base, actionsJson: "null" })).toThrow(
      AutomationItemParseError,
    );
    expect(() => dtoToEditor({ ...base, actionsJson: "{}" })).toThrow(
      AutomationItemParseError,
    );
  });

  it("throws when conditionsJson is not an array", () => {
    expect(() => dtoToEditor({ ...base, conditionsJson: "null" })).toThrow(
      AutomationItemParseError,
    );
    expect(() => dtoToEditor({ ...base, conditionsJson: "{}" })).toThrow(
      AutomationItemParseError,
    );
  });

  it("throws on invalid conditionsJson syntax", () => {
    expect(() =>
      dtoToEditor({ ...base, conditionsJson: "not-array{" }),
    ).toThrow(AutomationItemParseError);
  });

  it("throws on invalid scheduleJson", () => {
    expect(() => dtoToEditor({ ...base, scheduleJson: "{bad" })).toThrow(
      AutomationItemParseError,
    );
    expect(() => dtoToEditor({ ...base, scheduleJson: "[]" })).toThrow(
      AutomationItemParseError,
    );
  });

  it("throws on unknown kind", () => {
    expect(() =>
      dtoToEditor({ ...base, kind: "workflow" as AutomationItemDTO["kind"] }),
    ).toThrow(AutomationItemParseError);
  });
});

describe("newAutomationId", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("uses crypto.randomUUID when available", () => {
    vi.stubGlobal("crypto", {
      randomUUID: () => "11111111-1111-4111-8111-111111111111",
    });
    expect(newAutomationId()).toBe("11111111-1111-4111-8111-111111111111");
  });

  it("falls back when randomUUID is missing", () => {
    vi.stubGlobal("crypto", {});
    const id = newAutomationId();
    expect(id).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i,
    );
  });
});
