/**
 * jsdom does not implement ResizeObserver; TanStack Virtual and the gallery grid rely on it.
 */
class ResizeObserverMock implements ResizeObserver {
  constructor(private cb: ResizeObserverCallback) {}

  observe(target: Element): void {
    const w = target.clientWidth > 0 ? target.clientWidth : 480;
    const h = target.clientHeight > 0 ? target.clientHeight : 400;
    this.cb(
      [
        {
          target,
          contentRect: {
            x: 0,
            y: 0,
            width: w,
            height: h,
            top: 0,
            left: 0,
            bottom: h,
            right: w,
            toJSON() {
              return {};
            },
          },
          borderBoxSize: [],
          contentBoxSize: [],
          devicePixelContentBoxSize: [],
        } as ResizeObserverEntry,
      ],
      this,
    );
  }

  unobserve(): void {}

  disconnect(): void {}
}

globalThis.ResizeObserver = ResizeObserverMock;
