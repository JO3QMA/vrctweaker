import { describe, expect, it } from "vitest";
import {
  AUTOMATION_SPLIT_MIN_PX,
  automationPaneVisibility,
  automationShowsSplit,
} from "./automationPane";

describe("automationPaneVisibility", () => {
  it("keeps both panes at the split width and above", () => {
    expect(AUTOMATION_SPLIT_MIN_PX).toBe(720);
    for (const width of [720, 960, 1280]) {
      expect(automationShowsSplit(width)).toBe(true);
      expect(automationPaneVisibility(width, true, false)).toEqual({
        showList: true,
        showEditor: true,
      });
      expect(automationPaneVisibility(width, false, false)).toEqual({
        showList: true,
        showEditor: true,
      });
    }
  });

  it("shows one pane under 720px", () => {
    expect(automationShowsSplit(719)).toBe(false);
    expect(automationPaneVisibility(600, false, false)).toEqual({
      showList: true,
      showEditor: false,
    });
    expect(automationPaneVisibility(600, true, false)).toEqual({
      showList: false,
      showEditor: true,
    });
    expect(automationPaneVisibility(719, true, true)).toEqual({
      showList: true,
      showEditor: false,
    });
  });
});
