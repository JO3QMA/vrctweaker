/** Read a `:root` custom property (Storybook loads `style.css` in preview). */
export function readRootCssCustomProperty(varName: `--${string}`): string {
  if (typeof document === "undefined") {
    return "";
  }
  const prop = varName.startsWith("--") ? varName : `--${varName}`;
  return getComputedStyle(document.documentElement)
    .getPropertyValue(prop)
    .trim();
}
