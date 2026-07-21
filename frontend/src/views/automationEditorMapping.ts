import type { AutomationItemDTO } from "../wails/app";

export type ActionEditor = {
  type: string;
  status: string;
  powerPlanMode: "preset" | "guid";
  powerPlanPreset: string;
  powerPlanGuid: string;
  continueOnError: boolean;
};

export type EditorState = {
  id: string;
  name: string;
  kind: "rule" | "script";
  isEnabled: boolean;
  isNew?: boolean;
  triggerType: string;
  scheduleWeekdays: number[];
  scheduleHour: number;
  scheduleMinute: number;
  vrchatRunning: boolean;
  friendUserId: string;
  actions: ActionEditor[];
  scriptSource: string;
};

export class AutomationItemParseError extends Error {
  constructor(field: string) {
    super(`invalid automation item JSON: ${field}`);
    this.name = "AutomationItemParseError";
  }
}

export function defaultAction(): ActionEditor {
  return {
    type: "change_status",
    status: "busy",
    powerPlanMode: "preset",
    powerPlanPreset: "balanced",
    powerPlanGuid: "",
    continueOnError: false,
  };
}

function parseJsonField<T>(raw: string, field: string): T {
  try {
    return JSON.parse(raw) as T;
  } catch {
    throw new AutomationItemParseError(field);
  }
}

/** Maps a persisted item into editor state. Throws AutomationItemParseError on bad JSON. */
export function dtoToEditor(dto: AutomationItemDTO): EditorState {
  const state: EditorState = {
    id: dto.id,
    name: dto.name,
    kind: dto.kind === "script" ? "script" : "rule",
    isEnabled: dto.isEnabled,
    triggerType: dto.triggerType || "friend_joined",
    scheduleWeekdays: [],
    scheduleHour: 0,
    scheduleMinute: 0,
    vrchatRunning: false,
    friendUserId: "",
    actions: [defaultAction()],
    scriptSource: dto.scriptSource || "",
  };
  if (dto.scheduleJson) {
    const s = parseJsonField<{
      weekdays?: number[];
      hour?: number;
      minute?: number;
    }>(dto.scheduleJson, "scheduleJson");
    state.scheduleWeekdays = s.weekdays ?? [];
    state.scheduleHour = s.hour ?? 0;
    state.scheduleMinute = s.minute ?? 0;
  }
  if (dto.conditionsJson) {
    const conds = parseJsonField<Array<{ type?: string; vrcUserId?: string }>>(
      dto.conditionsJson,
      "conditionsJson",
    );
    for (const c of conds) {
      if (c.type === "vrchat_running") state.vrchatRunning = true;
      if (c.type === "friend_is" && c.vrcUserId) {
        state.friendUserId = c.vrcUserId;
      }
    }
  }
  if (dto.actionsJson) {
    const steps = parseJsonField<
      Array<{
        type?: string;
        payload?: Record<string, string>;
        continueOnError?: boolean;
      }>
    >(dto.actionsJson, "actionsJson");
    if (steps.length) {
      state.actions = steps.map((step) => ({
        type: step.type || "change_status",
        status: step.payload?.status ?? "busy",
        powerPlanMode: step.payload?.guid ? "guid" : "preset",
        powerPlanPreset: step.payload?.preset ?? "balanced",
        powerPlanGuid: step.payload?.guid ?? "",
        continueOnError: !!step.continueOnError,
      }));
    }
  }
  return state;
}
