const SCROLL_END_EPSILON = 1;

export function shouldShowLicensesTableScrollFade(
  scrollLeft: number,
  scrollWidth: number,
  clientWidth: number,
): boolean {
  const overflow = scrollWidth > clientWidth + SCROLL_END_EPSILON;
  if (!overflow) {
    return false;
  }
  const atEnd = scrollLeft + clientWidth >= scrollWidth - SCROLL_END_EPSILON;
  return !atEnd;
}

export function syncLicensesTableScrollFade(el: HTMLElement | null): boolean {
  if (!el) {
    return false;
  }
  return shouldShowLicensesTableScrollFade(
    el.scrollLeft,
    el.scrollWidth,
    el.clientWidth,
  );
}
