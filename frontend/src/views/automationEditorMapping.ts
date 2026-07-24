import type { AutomationItemDTO } from "../wails/app";

export type ActionEditor = {
  type: string;
  status: string;
  powerPlanMode: "preset" | "guid";
  powerPlanPreset: string;
  powerPlanGuid: string;
  windowWidth: number;
  windowHeight: number;
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
    windowWidth: 1280,
    windowHeight: 720,
    continueOnError: false,
  };
}

/** UUID for new items; falls back when crypto.randomUUID is unavailable. */
export function newAutomationId(): string {
  const c = globalThis.crypto as Crypto | undefined;
  if (c && typeof c.randomUUID === "function") {
    return c.randomUUID();
  }
  // ponytail: non-crypto fallback for older WebViews / odd test hosts.
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (ch) => {
    const r = (Math.random() * 16) | 0;
    const v = ch === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

function parseJsonField<T>(raw: string, field: string): T {
  try {
    return JSON.parse(raw) as T;
  } catch {
    throw new AutomationItemParseError(field);
  }
}

function parseJsonArray<T>(raw: string, field: string): T[] {
  const value = parseJsonField<unknown>(raw, field);
  if (!Array.isArray(value)) {
    throw new AutomationItemParseError(field);
  }
  return value as T[];
}

/** Maps editor state into a persistable item DTO. */
export function editorToDto(state: EditorState): AutomationItemDTO {
  const dto: AutomationItemDTO = {
    id: state.id,
    name: state.name,
    kind: state.kind,
    isEnabled: state.isEnabled,
  };
  if (state.kind === "script") {
    dto.scriptSource = state.scriptSource;
    return dto;
  }
  dto.triggerType = state.triggerType;
  if (state.triggerType === "schedule.tick") {
    dto.scheduleJson = JSON.stringify({
      weekdays: state.scheduleWeekdays,
      hour: state.scheduleHour,
      minute: state.scheduleMinute,
    });
  }
  const conds: Array<{ type: string; vrcUserId?: string }> = [];
  if (state.vrchatRunning) conds.push({ type: "vrchat_running" });
  // friend_is only applies to friend_joined; leftover picker state must not block schedule/process.
  if (state.triggerType === "friend_joined" && state.friendUserId) {
    conds.push({ type: "friend_is", vrcUserId: state.friendUserId });
  }
  dto.conditionsJson = JSON.stringify(conds);
  dto.actionsJson = JSON.stringify(
    state.actions.map((a) => {
      const payload: Record<string, string | number> = {};
      if (a.type === "change_status") payload.status = a.status;
      if (a.type === "set_power_plan") {
        if (a.powerPlanMode === "guid" && a.powerPlanGuid) {
          payload.guid = a.powerPlanGuid;
        } else {
          payload.preset = a.powerPlanPreset;
        }
      }
      if (a.type === "set_vrchat_window_size") {
        payload.width = a.windowWidth;
        payload.height = a.windowHeight;
      }
      return {
        type: a.type,
        payload,
        continueOnError: a.continueOnError || undefined,
      };
    }),
  );
  return dto;
}

/** Maps a persisted item into editor state. Throws AutomationItemParseError on bad JSON. */
export function dtoToEditor(dto: AutomationItemDTO): EditorState {
  if (dto.kind !== "script" && dto.kind !== "rule") {
    throw new AutomationItemParseError("kind");
  }
  const state: EditorState = {
    id: dto.id,
    name: dto.name,
    kind: dto.kind,
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
    if (s === null || typeof s !== "object" || Array.isArray(s)) {
      throw new AutomationItemParseError("scheduleJson");
    }
    state.scheduleWeekdays = s.weekdays ?? [];
    state.scheduleHour = s.hour ?? 0;
    state.scheduleMinute = s.minute ?? 0;
  }
  if (dto.conditionsJson) {
    const conds = parseJsonArray<{ type?: string; vrcUserId?: string }>(
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
    const steps = parseJsonArray<{
      type?: string;
      payload?: Record<string, unknown>;
      continueOnError?: boolean;
    }>(dto.actionsJson, "actionsJson");
    if (steps.length) {
      state.actions = steps.map((step) => ({
        type: step.type || "change_status",
        status: String(step.payload?.status ?? "busy"),
        powerPlanMode: step.payload?.guid ? "guid" : "preset",
        powerPlanPreset: String(step.payload?.preset ?? "balanced"),
        powerPlanGuid: String(step.payload?.guid ?? ""),
        windowWidth: Number(step.payload?.width ?? 1280) || 1280,
        windowHeight: Number(step.payload?.height ?? 720) || 720,
        continueOnError: !!step.continueOnError,
      }));
    }
  }
  return state;
}
