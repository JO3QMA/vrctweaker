// Wails runtime - provides window drag, minimize, maximize, quit
// When running in Wails, these are injected by the runtime

export interface WailsRuntime {
  WindowMinimise?: () => void
  WindowToggleMaximise?: () => void
  Quit?: () => void
}

declare global {
  interface Window {
    runtime?: WailsRuntime
  }
}

export function getRuntime(): WailsRuntime | undefined {
  return typeof window !== 'undefined' ? window.runtime : undefined
}
