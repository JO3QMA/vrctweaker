/** Read the log width before observing so the first paint is not the six-column layout. */
export function startLogWidthObserver(
  el: HTMLElement | null | undefined,
  onWidth: (width: number) => void,
): () => void {
  onWidth(el?.offsetWidth ?? 0);
  if (typeof ResizeObserver === "undefined" || !el) return () => {};
  const observer = new ResizeObserver(() => {
    onWidth(el.offsetWidth ?? 0);
  });
  observer.observe(el);
  return () => observer.disconnect();
}
