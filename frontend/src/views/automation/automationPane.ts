/** Frame widths below this show one pane. Matches `@container automation-frame (max-width: 719px)`. */
export const AUTOMATION_SPLIT_MIN_PX = 720;

export function automationShowsSplit(frameWidth: number): boolean {
  return frameWidth >= AUTOMATION_SPLIT_MIN_PX;
}

/** Which pane is visible. Wide frames show both; narrow frames show one. */
export function automationPaneVisibility(
  frameWidth: number,
  hasEditor: boolean,
  preferList: boolean,
): { showList: boolean; showEditor: boolean } {
  if (automationShowsSplit(frameWidth)) {
    return { showList: true, showEditor: true };
  }
  const showEditor = hasEditor && !preferList;
  return { showList: !showEditor, showEditor };
}
