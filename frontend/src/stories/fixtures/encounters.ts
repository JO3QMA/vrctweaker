import type { UserEncounterDTO } from "../../wails/app";

export const sampleActivityEncounters: UserEncounterDTO[] = [
  {
    id: "story-1",
    vrcUserId: "usr_story",
    displayName: "ストーリー太郎",
    instanceId: "inst_story",
    worldId: "wrld_story",
    worldDisplayName: "Sample World",
    joinedAt: new Date(Date.now() - 3_600_000).toISOString(),
    leftAt: new Date(Date.now() - 1_800_000).toISOString(),
    isFirstEncounter: false,
  },
];

export const encounterByUserSample: UserEncounterDTO[] = [
  {
    id: "eh-user-1",
    vrcUserId: "usr_enc_story",
    displayName: "Encounter User",
    instanceId: "inst_u1",
    worldId: "wrld_u1",
    worldDisplayName: "User Story World",
    joinedAt: "2026-03-01T12:00:00+09:00",
    leftAt: "2026-03-01T13:00:00+09:00",
    isFirstEncounter: false,
  },
];

export const encounterByWorldSample: UserEncounterDTO[] = [
  {
    id: "eh-world-1",
    vrcUserId: "usr_w1",
    displayName: "Visitor One",
    instanceId: "inst_w1",
    worldId: "wrld_enc_story",
    worldDisplayName: "World Story",
    joinedAt: "2026-03-02T09:00:00+09:00",
    leftAt: "2026-03-02T10:30:00+09:00",
    isFirstEncounter: false,
  },
];

export const userProfileEncounters: UserEncounterDTO[] = [
  {
    id: "up-story-1",
    vrcUserId: "usr_story",
    displayName: "Story User",
    instanceId: "inst_1",
    worldId: "wrld_1",
    worldDisplayName: "Test World",
    joinedAt: "2026-02-01T10:00:00+09:00",
    leftAt: "2026-02-01T11:00:00+09:00",
    isFirstEncounter: false,
  },
];
